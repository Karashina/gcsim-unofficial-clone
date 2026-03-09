package chongyun

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillHitmark   = 36
	skillFieldKey  = "chongyunfield"
	particleICDKey = "chongyun-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(52) // E -> N1
	skillFrames[action.ActionBurst] = 51   // E -> Q
	skillFrames[action.ActionDash] = 35    // E -> D
	skillFrames[action.ActionJump] = 35    // E -> J
	skillFrames[action.ActionSwap] = 49    // E -> J
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Spirit Blade: Chonghua's Layered Frost",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagElementalArt,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150,
		Element:            attributes.Cryo,
		Durability:         50,
		Mult:               skill[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.09 * 60,
		CanBeDefenseHalted: false,
	}

	// 祭礁大剣/大量のCD短縮によるフィールド終了時の固有天賦4処理
	// フィールドティックと固有天賦4タスクを無効化するためにsrcが必要
	src := c.Core.F
	c.fieldSrc = c.Core.F
	// フィールドがまだ存在する場合、既存の固有天賦4タスクを無効化し、新しい固有天賦4のスナップショット前にダメージを与える
	// 新しいスキル範囲が決定される前に実行する必要がある
	if c.Core.Status.Duration(skillFieldKey) > 0 {
		c.a4(skillHitmark+45, c.Core.F, true) // フィールド消滅まで約45フレーム
	}

	// フィールドダメージ/スキル範囲を処理
	c.skillArea = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.5}, 8)
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.skillArea.Shape.Pos(), nil, 2.5),
		0,
		skillHitmark,
		c.particleCB,
		c.makeC4Callback(),
	)

	// フィールド生成を処理
	c.QueueCharTask(func() {
		c.Core.Status.Add(skillFieldKey, 600)
	}, skillHitmark)

	// フィールドティックを処理
	// TODO: 霜フィールドのティック開始間の遅延？
	for i := 0; i <= 600; i += 60 {
		c.Core.Tasks.Add(func() {
			if src != c.fieldSrc {
				return
			}
			if !c.Core.Combat.Player().IsWithinArea(c.skillArea) {
				return
			}
			active := c.Core.Player.ActiveChar()
			c.infuse(active)
		}, i+skillHitmark)
	}

	// 期限切れによるフィールド終了時の固有天賦4処理
	c.a4(655, c.Core.F, false)

	c.SetCDWithDelay(action.ActionSkill, 900, 34)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
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
	c.AddStatus(particleICDKey, 0.2*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Cryo, c.ParticleDelay)
}

func (c *char) onSwapHook() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.Core.Status.Duration("chongyunfield") == 0 {
			return false
		}
		// 交代時に元素付与を追加
		dur := int(infuseDur[c.TalentLvlSkill()] * 60)
		c.Core.Log.NewEvent("chongyun adding infusion on swap", glog.LogCharacterEvent, c.Index).
			Write("expiry", c.Core.F+dur)
		active := c.Core.Player.ActiveChar()
		c.infuse(active)
		return false
	}, "chongyun-field")
}

func (c *char) infuse(active *character.CharWrapper) {
	dur := int(infuseDur[c.TalentLvlSkill()] * 60)
	// 2凸はCDを15%短縮
	if c.Base.Cons >= 2 {
		active.AddCooldownMod(character.CooldownMod{
			Base: modifier.NewBaseWithHitlag("chongyun-c2", dur),
			Amount: func(a action.Action) float64 {
				if a == action.ActionSkill || a == action.ActionBurst {
					return -0.15
				}
				return 0
			},
		})
	}

	// 武器元素付与と固有天賦1
	switch active.Weapon.Class {
	case info.WeaponClassClaymore, info.WeaponClassSpear, info.WeaponClassSword:
		c.Core.Player.AddWeaponInfuse(
			active.Index,
			"chongyun-ice-weapon",
			attributes.Cryo,
			dur,
			true,
			attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
		)
		c.Core.Log.NewEvent("chongyun adding infusion", glog.LogCharacterEvent, c.Index).
			Write("expiry", c.Core.F+dur)
		// 固有天賦1:
		// フィールド内の片手剣・両手剣・長柄武器キャラクターの通常攻撃速度が8%増加する。
		if c.Base.Ascension >= 1 {
			m := make([]float64, attributes.EndStatType)
			m[attributes.AtkSpd] = 0.08
			active.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("chongyun-field", dur),
				AffectedStat: attributes.NoStat,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}
	default:
		return
	}
}
