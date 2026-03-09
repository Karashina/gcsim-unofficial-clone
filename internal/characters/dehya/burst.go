package dehya

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

const (
	burstKey              = "dehya-burst"
	burstDuration         = 4.1 * 60
	kickKey               = "dehya-burst-kick"
	burstPunch1Hitmark    = 105
	burstPunchSlowHitmark = 50
	burstKickHitmark      = 46
)

var (
	kickFrames    []int
	punchHitmarks = []int{30, 30, 28, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27}
)

func init() {
	kickFrames = frames.InitAbilSlice(76) // Q -> N1
	kickFrames[action.ActionSkill] = 71
	kickFrames[action.ActionDash] = 73
	kickFrames[action.ActionJump] = 73
	kickFrames[action.ActionSwap] = burstKickHitmark
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.c6Count = 0
	c.sanctumSavedDur = 0
	if c.StatusIsActive(dehyaFieldKey) {
		// 開始時にフィールドを回収
		c.pickUpField()
	}

	c.Core.Tasks.Add(func() {
		c.burstHitSrc = 0
		c.burstCounter = 0
		c.AddStatus(burstKey, burstDuration, true)
		c.burstPunchFunc(c.burstHitSrc)()
	}, burstPunch1Hitmark)

	c.ConsumeEnergy(15) //TODO: これがping依存なら、0pingでは1に近い可能性がある
	c.SetCDWithDelay(action.ActionBurst, 18*60, 1)

	return action.Info{
		Frames:          func(action.Action) int { return burstPunch1Hitmark },
		AnimationLength: burstPunch1Hitmark,
		CanQueueAfter:   burstPunch1Hitmark,
		State:           action.BurstState,
	}, nil
}

func (c *char) burstPunchFunc(src int) func() {
	return func() {
		if c.burstHitSrc != src {
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flame-Mane's Fist",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   50,
			Element:    attributes.Pyro,
			Durability: 25,
			Mult:       burstPunchAtk[c.TalentLvlBurst()],
			FlatDmg:    (c.c1FlatDmgRatioQ + burstPunchHP[c.TalentLvlBurst()]) * c.MaxHP(),
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -2.8}, 5, 7.8),
			0,
			0,
			c.c4CB(),
			c.c6CB(),
		)
		if !c.StatusIsActive(burstKey) {
			c.burstHitSrc++
			c.AddStatus(kickKey, burstKickHitmark, true)
			c.Core.Tasks.Add(c.burstKickFunc(c.burstHitSrc), burstKickHitmark)
			return
		}
		c.burstCounter++
		c.burstHitSrc++
		c.Core.Tasks.Add(c.burstPunchFunc(c.burstHitSrc), burstPunchSlowHitmark)
	}
}

func (c *char) burstKickFunc(src int) func() {
	return func() {
		if src != c.burstHitSrc { // 重複を防止
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Incineration Drive",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   100,
			Element:    attributes.Pyro,
			Durability: 25,
			Mult:       burstKickAtk[c.TalentLvlBurst()],
			FlatDmg:    (c.c1FlatDmgRatioQ + burstKickHP[c.TalentLvlBurst()]) * c.MaxHP(),
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 6.5),
			0,
			0,
			c.c4CB(),
		)
		if dur := c.sanctumSavedDur; dur > 0 { // 自己トリガーを回避するため1フレーム遅延で領域を設置
			c.sanctumSavedDur = 0
			c.Core.Tasks.Add(func() {
				c.AddStatus(skillICDKey, c.sanctumICD, false)
				c.addField(dur)
			}, 1)
		}
	}
}

func (c *char) UseBurstAction() *action.Info {
	var out action.Info
	c.burstHitSrc++
	if c.StatusIsActive(kickKey) {
		out = c.burstKick(c.burstHitSrc)
		return &out
	}
	if c.StatusIsActive(burstKey) {
		out = c.burstPunch(c.burstHitSrc, false)
		return &out
	}
	return nil
}

func (c *char) burstPunch(src int, auto bool) action.Info {
	hitmark := burstPunchSlowHitmark
	if !auto {
		hitmark = punchHitmarks[c.burstCounter]
	}

	c.Core.Tasks.Add(c.burstPunchFunc(src), hitmark)

	return action.Info{
		Frames:          func(action.Action) int { return hitmark },
		AnimationLength: hitmark,
		CanQueueAfter:   hitmark,
		State:           action.Idle, // TODO: 元素爆発ステートは無敵フレームを意味するため使用不可
	}
}

func (c *char) burstKick(src int) action.Info {
	c.Core.Tasks.Add(c.burstKickFunc(src), burstKickHitmark)
	return action.Info{
		Frames:          frames.NewAbilFunc(kickFrames),
		AnimationLength: kickFrames[action.ActionAttack],
		CanQueueAfter:   kickFrames[action.ActionSwap], // 最速キャンセル
		State:           action.Idle,                   // TODO: 元素爆発ステートは無敵フレームを意味するため使用不可
	}
}
