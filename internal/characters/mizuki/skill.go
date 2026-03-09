package mizuki

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillHitmark                  = 2
	skillActivateDmgName          = "Aisa Utamakura Pilgrimage"
	skillActivateDmgRadius        = 5.5
	skillActivatePoise            = 30
	skillActivateDurability       = 25
	skillCdDelay                  = 23
	skillCd                       = 15 * 60
	skillParticleGenerations      = 4
	skillParticleGenerationIcd    = 0.5 * 60
	skillParticleGenerationIcdKey = "mizuki-particle-icd"
	cloudDmgName                  = "Dreamdrifter Continuous Attack"
	cloudPoise                    = 20
	cloudDurability               = 25
	cloudExplosionRadius          = 4
	cloudTravelTime               = 30
	cloudFirstHit                 = 18
	cloudHitInterval              = 45
	dreamDrifterStateKey          = "dreamdrifter-state"
	dreamDrifterBaseDuration      = 5 * 60
	dreamDrifterSwirlBuffKey      = "mizuki-swirl-buff"
	mizukiSwapOutKey              = "mizuki-exit"
)

func init() {
	skillFrames = frames.InitAbilSlice(50) // E -> E
	skillFrames[action.ActionBurst] = 34   // E -> Q
	skillFrames[action.ActionSwap] = 30    // E -> Swap
}

// 美しい夢の記憶を紡ぎ、地面の上に浮遊するDreamdrifter状態に入り、
// 周囲の敵に範囲風元素ダメージを1回与える。
//
// Dreamdrifter
//
//   - Dreamdrifter状態中、夢見月瑞希は継続的に前方に漂いながら、
//     一定間隔で周囲の敵に範囲風元素ダメージを与える。
//
//   - この間、夢見月瑞希は漂う方向を制御でき、元素爆発「安楽秘湯浴」の
//     夢見式特製おやつの拾得距離が増加する。
//
//   - 夢見月瑞希の元素熟知に基づいて、周囲のパーティメンバーの拡散ダメージが増加する。
//
//     瑞希がフィールドを離れるか、再度元素スキルを使用するとDreamdrifterは終了する。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	// Dreamdrifter状態中に使用した場合、状態をキャンセルする。
	if c.StatusIsActive(dreamDrifterStateKey) {
		c.cancelDreamDrifterState()
		return action.Info{
			Frames:          frames.NewAbilFunc(skillFrames),
			AnimationLength: skillFrames[action.InvalidAction],
			CanQueueAfter:   skillFrames[action.ActionSwap], // 最速キャンセルはスワップ
			State:           action.SkillState,
		}, nil
	}

	// 発動ダメージ
	activationAttack := combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         skillActivateDmgName,
		AttackTag:    attacks.AttackTagElementalArt,
		ICDTag:       attacks.ICDTagNone,
		ICDGroup:     attacks.ICDGroupDefault,
		StrikeType:   attacks.StrikeTypeDefault,
		PoiseDMG:     skillActivatePoise,
		Element:      attributes.Anemo,
		Durability:   skillActivateDurability,
		Mult:         skill[c.TalentLvlSkill()],
		HitlagFactor: 0.05,
	}

	c.Core.QueueAttack(
		activationAttack,
		combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			nil,
			skillActivateDmgRadius,
		),
		0,
		skillHitmark,
		c.particleCB,
	)

	c.particleGenerationsRemaining = skillParticleGenerations

	if c.Base.Ascension >= 1 {
		c.dreamDrifterExtensionsRemaining = dreamDrifterExtensions
	}

	travel, ok := p["travel"]
	if !ok {
		travel = cloudTravelTime
	}
	c.applyDreamDrifterEffect(travel)

	c.SetCDWithDelay(action.ActionSkill, skillCd, skillCdDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // 最速キャンセルは交代
		State:           action.SkillState,
	}, nil
}

func (c *char) applyDreamDrifterEffect(travel int) {
	c.AddStatus(dreamDrifterStateKey, dreamDrifterBaseDuration, true)

	c.startCloudAttacks(travel)

	if c.Base.Cons >= 1 {
		// テストによるとデバフは3.5秒かからずに適用されるが、スキル発動時の初回拡散では発動しない。
		// 最初のクラウド（スキル発動後0.45秒）で発動可能なので数フレーム後にキューに入れる
		c.c1Task(c.cloudSrc, skillHitmark+2)
	}
}

func (c *char) skillInit() {
	for _, char := range c.Core.Player.Chars() {
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(dreamDrifterSwirlBuffKey, -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if !c.StatusIsActive(dreamDrifterStateKey) {
					return 0, false
				}
				// これらのフラグはAoE拡散を意味し、その場合この拡散ダメージボーナスは適用されない。
				// 先行のコールバック呼び出しで計算済みのため。この場合は他の反応ボーナスが
				// 代わりに適用される（例：蒸発ダメージボーナス、激化ダメージボーナス等）
				if ai.Amped || ai.Catalyzed {
					return 0, false
				}
				switch ai.AttackTag {
				case attacks.AttackTagSwirlCryo:
				case attacks.AttackTagSwirlElectro:
				case attacks.AttackTagSwirlHydro:
				case attacks.AttackTagSwirlPyro:
				default:
					return 0, false
				}

				return swirlDMG[c.TalentLvlSkill()] * c.Stat(attributes.EM), false
			},
		})
	}

	// フィールドを離れた時にDreamdrifter状態を解除する
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)

		if prev == c.Index && c.StatusIsActive(dreamDrifterStateKey) {
			c.cancelDreamDrifterState()
		}

		return false
	}, mizukiSwapOutKey)
}

func (c *char) startCloudAttacks(travel int) {
	// クラウドのダメージは発動時にスナップショット
	c.cloudAttack = combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         cloudDmgName,
		AttackTag:    attacks.AttackTagElementalArt,
		ICDTag:       attacks.ICDTagElementalArt,
		ICDGroup:     attacks.ICDGroupMizukiSkill,
		StrikeType:   attacks.StrikeTypeDefault,
		PoiseDMG:     cloudPoise,
		Element:      attributes.Anemo,
		Durability:   cloudDurability,
		Mult:         cloudDMG[c.TalentLvlSkill()],
		HitlagFactor: 0.05,
	}
	c.cloudSnap = c.Snapshot(&c.cloudAttack)

	// 最初のクラウドはスキル発動後約20フレームで発射される。
	c.cloudSrc = c.Core.F
	c.cloudTask(travel, c.cloudSrc, cloudFirstHit)
}

// 発動時またはクラウドの各Eダメージにつき最大4個の粒子を生成する。
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}

	if c.StatusIsActive(skillParticleGenerationIcdKey) {
		return
	}

	if c.particleGenerationsRemaining > 0 {
		c.AddStatus(skillParticleGenerationIcdKey, skillParticleGenerationIcd, false)
		c.particleGenerationsRemaining--
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Anemo, c.ParticleDelay)
	}
}

func (c *char) cancelDreamDrifterState() {
	c.DeleteStatus(dreamDrifterStateKey)
	c.cloudSrc = -1

	c.Core.Log.NewEvent("DreamDrifter effect cancelled", glog.LogCharacterEvent, c.Index)
}

func (c *char) cloudTask(travel, src, hitmark int) {
	c.QueueCharTask(func() {
		if c.cloudSrc != src {
			return
		}
		if !c.StatusIsActive(dreamDrifterStateKey) {
			return
		}
		c.Core.QueueAttackWithSnap(
			c.cloudAttack,
			c.cloudSnap,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, cloudExplosionRadius),
			travel,
			c.particleCB,
		)
		c.cloudTask(travel, src, cloudHitInterval)
	}, hitmark)
}
