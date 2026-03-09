package chevreuse

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

const (
	skillPressCDStart      = 18
	skillPressHitmark      = 25
	skillPressArkheHitmark = 59

	skillHoldCDStart      = 13
	skillHoldHitmark      = 19
	skillHoldArkheHitmark = 55

	skillHealKey      = "chev-skill-heal"
	skillHealInterval = 120
	particleICDKey    = "chev-particle-icd"
	arkheICDKey       = "chev-arkhe-icd"
)

var skillPressFrames []int
var skillHoldFrames []int

func init() {
	// skill (press) -> x
	skillPressFrames = frames.InitAbilSlice(31) // E -> N1/Q
	skillPressFrames[action.ActionDash] = 23
	skillPressFrames[action.ActionJump] = 25
	skillPressFrames[action.ActionWalk] = 24
	skillPressFrames[action.ActionSwap] = 24

	// skill (hold) -> x
	skillHoldFrames = frames.InitAbilSlice(26) // E -> Q
	skillHoldFrames[action.ActionAttack] = 25
	skillHoldFrames[action.ActionDash] = 21
	skillHoldFrames[action.ActionJump] = 23
	skillHoldFrames[action.ActionWalk] = 24
	skillHoldFrames[action.ActionSwap] = 23
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if p["hold"] == 0 {
		return c.skillPress(), nil
	}
	return c.skillHold(p), nil
}

func (c *char) skillPress() action.Info {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Short-Range Rapid Interdiction Fire",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       skillPress[c.TalentLvlSkill()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: -0.5}, 2, 6),
		skillPressHitmark,
		skillPressHitmark,
		c.particleCB,
		c.arkhe(skillPressArkheHitmark-skillPressHitmark),
	)

	c.skillHeal(skillPressCDStart)
	c.SetCDWithDelay(action.ActionSkill, 15*60, skillPressCDStart)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) skillHold(p map[string]int) action.Info {
	hold := p["hold"]
	// 最も早い長押しヒットマークは約19f
	// 最も遅い長押しヒットマークは約319f
	// hold=1で19f、hold=301で319fのヒットマークまでの遅延を得る
	if hold < 1 {
		hold = 1
	}
	if hold > 301 {
		hold = 301
	}
	// 長押しを示すために>0を供給する必要があるため1を引く
	hold -= 1
	hitmark := hold + skillHoldHitmark
	cdStart := hold + skillHoldCDStart

	var ai combat.AttackInfo
	var ap combat.AttackPattern

	if c.overChargedBall {
		ai = combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Short-Range Rapid Interdiction Fire [Overcharged]",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			Element:    attributes.Pyro,
			Durability: 25,
			PoiseDMG:   125,
			Mult:       skillOvercharged[c.TalentLvlSkill()],
		}

		ap = combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5)
		// 過充填弾発射後にステータスを削除
		c.overChargedBall = false
		c.Core.Tasks.Add(c.a4, cdStart)
	} else {
		ai = combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Short-Range Rapid Interdiction Fire [Hold]",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Pyro,
			Durability: 25,
			Mult:       skillHold[c.TalentLvlSkill()],
		}
		ap = combat.NewBoxHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: -0.5}, 3, 7)
	}

	// 4凸
	if c.StatModIsActive(c4StatusKey) {
		c.c4ShotsLeft -= 1
		if c.c4ShotsLeft == 0 {
			c.DeleteStatus(c4StatusKey)
		}
	} else {
		c.SetCDWithDelay(action.ActionSkill, 15*60, cdStart)
	}

	c.Core.QueueAttack(
		ai,
		ap,
		hitmark,
		hitmark,
		c.particleCB,
		c.c2(),
		c.arkhe(skillHoldArkheHitmark-skillHoldHitmark),
	)

	c.skillHeal(cdStart)

	return action.Info{
		Frames:          func(next action.Action) int { return hold + skillHoldFrames[next] },
		AnimationLength: hold + skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   hold + skillHoldFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) arkhe(delay int) combat.AttackCBFunc {
	// 敵だけでなく何かに命中した時にトリガー
	return func(a combat.AttackCB) {
		if c.StatusIsActive(arkheICDKey) {
			return
		}
		c.AddStatus(arkheICDKey, 10*60, true)

		aiArkhe := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Surging Blade (" + c.Base.Key.Pretty() + ")",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Pyro,
			Durability: 0,
			Mult:       arkhe[c.TalentLvlSkill()],
		}

		c.Core.QueueAttack(
			aiArkhe,
			combat.NewCircleHitOnTarget(a.Target.Pos(), nil, 2),
			delay,
			delay,
		)
	}
}

func (c *char) skillHeal(delay int) {
	skillDur := 12*60 + 1 // 有効期限の最後のティックで回復
	c.Core.Tasks.Add(func() {
		// C4+でスキル回復がまだ有効な場合のみステータスをリフレッシュ
		if c.StatusIsActive(skillHealKey) {
			c.AddStatus(skillHealKey, skillDur, false) // ヒットラグ延長なし
			return
		}
		c.AddStatus(skillHealKey, skillDur, false)               // ヒットラグ延長なし
		c.Core.Tasks.Add(c.startSkillHealing, skillHealInterval) // 最初の回復は2秒後
		// 既にキュー済みの6凸チーム回復がある場合は再キューしない
		if c.c6HealQueued {
			return
		}
		c.c6HealQueued = true
		c.Core.Tasks.Add(c.c6TeamHeal, 12*60)
	}, delay)
}

func (c *char) startSkillHealing() {
	if !c.StatusIsActive(skillHealKey) {
		return
	}

	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  c.Core.Player.Active(),
		Message: "Short-Range Rapid Interdiction Fire Healing",
		Src:     skillHpRegen[c.TalentLvlSkill()]*c.MaxHP() + skillHpFlat[c.TalentLvlSkill()],
		Bonus:   c.Stat(attributes.Heal),
	})
	c.c6(c.Core.Player.ActiveChar())
	c.Core.Tasks.Add(c.startSkillHealing, skillHealInterval)
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 10*60, false) // シュヴルーズは10秒の粒子ICD、ヒットラグ延長なし
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Pyro, c.ParticleDelay)
}

func (c *char) overchargedBallEventSub() {
	c.Core.Events.Subscribe(event.OnOverload, func(args ...interface{}) bool {
		// ガジェットには発動しない
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		c.overChargedBall = true
		return false
	}, "chev-overcharged-ball")
}
