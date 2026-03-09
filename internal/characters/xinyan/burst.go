package xinyan

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int
var c2PulseHitmarks = []int{65, 83}

const burstInitialHitmark = 22
const burstShieldStart = 43
const burstDoT1Hitmark = 57

func init() {
	burstFrames = frames.InitAbilSlice(87) // Q -> E/D/J
	burstFrames[action.ActionAttack] = 86  // Q -> N1
	burstFrames[action.ActionSwap] = 86    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Riff Revolution",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Physical,
		Durability:         100,
		Mult:               burstDmg[c.TalentLvlBurst()],
		CanBeDefenseHalted: true,
	}
	c1CB := c.makeC1CB()
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3),
		burstInitialHitmark,
		burstInitialHitmark,
		c1CB,
	)

	// 7ヒット
	ai = combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Riff Revolution (DoT)",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagElementalBurstPyro,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Pyro,
		Durability:         25,
		Mult:               burstDot[c.TalentLvlBurst()],
		CanBeDefenseHalted: true,
	}
	// 1回目のDoT
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 4),
			0,
			0,
			c1CB,
		)
		ai.CanBeDefenseHalted = false // 最初のDoTのみヒットラグあり
		// 2回目以降のDoT
		c.QueueCharTask(func() {
			for i := 0; i < 6; i++ {
				c.Core.QueueAttack(
					ai,
					combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 4),
					i*17,
					i*17,
					c1CB,
				)
			}
		}, 17)
	}, burstDoT1Hitmark)

	if c.Base.Cons >= 2 {
		// TODO: スナップショットのタイミング？
		defFactor := c.TotalDef(false)
		c.QueueCharTask(func() {
			c.updateShield(3, defFactor)
		}, burstShieldStart)

		// 2凸でレベル3シールドDoTの追加パルスが発生
		// fpsにより変動、フレーム動画では2パルス
		// 参照: https://library.keqingmains.com/evidence/characters/pyro/xinyan#xinyan-c2-shield-formation-pulses-extra-times
		ai := c.getAttackInfoShieldDoT()
		for i := 0; i < 2; i++ {
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3),
				1,
				c2PulseHitmarks[i],
				c.makeC1CB(),
			)
		}
	}

	c.ConsumeEnergy(5)
	c.SetCDWithDelay(action.ActionBurst, 15*60, 1)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionAttack], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
