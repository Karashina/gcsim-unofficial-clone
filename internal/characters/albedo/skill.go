package albedo

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const skillHitmark = 25

func init() {
	skillFrames = frames.InitAbilSlice(33) // E -> Q
	skillFrames[action.ActionAttack] = 32  // E -> N1
	skillFrames[action.ActionDash] = 29    // E -> D
	skillFrames[action.ActionJump] = 28    // E -> J
	skillFrames[action.ActionSwap] = 31    // E -> Swap
}

const (
	skillICDKey    = "albedo-skill-icd"
	particleICDKey = "albedo-particle-icd"
)

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Abiogenesis: Solar Isotoma",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   50,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	// TODO: ダメージフレーム
	c.bloomSnapshot = c.Snapshot(&ai)

	player := c.Core.Combat.Player()
	skillDir := player.Direction()
	// 元素スキル単押しのヒットボックスオフセットを想定
	skillPos := geometry.CalcOffsetPoint(c.Core.Combat.Player().Pos(), geometry.Point{Y: 3}, player.Direction())
	c.skillArea = combat.NewCircleHitOnTarget(skillPos, nil, 10)

	c.Core.QueueAttackWithSnap(ai, c.bloomSnapshot, combat.NewCircleHitOnTarget(skillPos, nil, 5), skillHitmark)

	// Tick用のスナップショット
	ai.Abil = "Abiogenesis: Solar Isotoma (Tick)"
	ai.ICDTag = attacks.ICDTagElementalArt
	ai.Mult = skillTick[c.TalentLvlSkill()]
	ai.UseDef = true
	c.skillAttackInfo = ai
	c.skillSnapshot = c.Snapshot(&c.skillAttackInfo)

	// 設置物を生成
	// 設置物はヒット着弾後まで完全に形成されない（正確なタイミングは不明）
	c.Core.Tasks.Add(func() {
		c.Core.Constructs.New(c.newConstruct(1800, skillDir, skillPos), true)
		c.lastConstruct = c.Core.F
		c.skillActive = true
		// 設置物生成後にICDをリセット
		c.DeleteStatus(skillICDKey)

		// Hexerei: Secret Riteバフをパーティに適用
		c.hexereiSecretRite()

		// 1凸追加バフ: スキル使用時に防御力+50%を20秒間
		if c.Base.Cons >= 1 {
			c.c1DEFBuff()
		}

		// 4凸と6凸のチェックを追加
		if c.Base.Cons >= 4 {
			c.Core.Tasks.Add(c.c4(c.Core.F), 18) // 0.3秒後にチェック開始
		}
		if c.Base.Cons >= 6 {
			c.Core.Tasks.Add(c.c6(c.Core.F), 18) // 0.3秒後にチェック開始
		}
	}, skillHitmark)

	c.SetCDWithDelay(action.ActionSkill, 240, 23)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillHitmark,
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 1*60, false)
	if c.Core.Rand.Float64() < 0.67 {
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Geo, c.ParticleDelay)
	}
}

func (c *char) skillHook() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		trg := args[0].(combat.Target)
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		if !c.skillActive {
			return false
		}
		if c.StatusIsActive(skillICDKey) {
			return false
		}
		// 更新時に自身ではトリガーされない
		if atk.Info.Abil == "Abiogenesis: Solar Isotoma" {
			return false
		}
		if dmg == 0 {
			return false
		}
		// 命中したターゲットがスキル範囲外なら発動しない
		if !trg.IsWithinArea(c.skillArea) {
			return false
		}

		// このICDはおそらく設置物に紐づいているため、ヒットラグで延長されない
		c.AddStatus(skillICDKey, 120, false) // 2秒ごとに発動

		c.Core.QueueAttackWithSnap(
			c.skillAttackInfo,
			c.skillSnapshot,
			combat.NewCircleHitOnTarget(trg, nil, 3.4),
			1,
			c.particleCB,
		)

		// 1凸: スキル発動時にエネルギー1.2回復
		if c.Base.Cons >= 1 {
			c.AddEnergy("albedo-c1", 1.2)
			c.Core.Log.NewEvent("c1 restoring energy", glog.LogCharacterEvent, c.Index)
		}

		// 2凸: スキル発動時にスタック付与、30秒持続; 各スタックで元素爆発ダメージが防御力の30%分増加、最大4スタック
		if c.Base.Cons >= 2 {
			if !c.StatusIsActive(c2key) {
				c.c2stacks = 0
			}
			c.AddStatus(c2key, 1800, true) // 30秒持続
			c.c2stacks++
			if c.c2stacks > 4 {
				c.c2stacks = 4
			}
		}

		return false
	}, "albedo-skill")
}
