package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	burstFrames      []int
	tsFrames         []int
	NascentHitmark   = []int{111, 125, 185}           // middle phase x2 + final phase x1
	AscendantHitmark = []int{111, 125, 129, 138, 164} // middle phase x4 + final phase x1
)

const (
	initialHitmark = 96
	tsHitmark      = 45
	tsAddHitmark   = 66
)

func init() {
	burstFrames = frames.InitAbilSlice(113)
	tsFrames = frames.InitAbilSlice(55)
}

// Q
// Ancient Ritual: Cometh the Night (Q)
// Flins deals single AoE Electro DMG and after a short delay, dealing 2 instances of middle-phase and 1 instance of final-phase AoE Electro DMG, all of which are considered Lunar-Charged DMG.
// When the moonsign is Moonsign: Ascendant Gleam, this ability is enhanced: If there are thunderclouds nearby, Flins will deal an additional 2 instances of middle-phase Lunar-Charged AoE Electro DMG.

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// After using the special Elemental Skill: Northland Spearstorm, Flins's Elemental Burst Ancient Ritual: Cometh the Night will be replaced with the special Elemental Burst Thunderous Symphony for the next 6s.
	// Thunderous Symphony
	// Consume less Elemental Energy to unleash a special Elemental Burst. Flins deals a single instance of AoE Electro DMG that is considered Lunar-Charged DMG.
	// When the moonsign is Moonsign: Ascendant Gleam, Flins's Skill is enhanced: If there are thunderclouds nearby, Flins will deal an additional instance of Lunar-Charged AoE Electro DMG.
	if c.StatusIsActive(northlandKey) {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins TS Dummy",
			FlatDmg:    0,
		}
		aiadd := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins TSADD Dummy",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), tsHitmark, tsHitmark)
		if c.MoonsignAscendant && c.lcCloudCheck() {
			c.Core.QueueAttack(aiadd, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), tsAddHitmark, tsAddHitmark)
		}

		c.DeleteStatus(northlandKey)
		c.ConsumeEnergyPartial(7, 30)

		return action.Info{
			Frames:          frames.NewAbilFunc(tsFrames),
			AnimationLength: tsFrames[action.InvalidAction],
			CanQueueAfter:   tsFrames[action.ActionSwap], // earliest cancel
			State:           action.BurstState,
		}, nil
	} else {
		// initial hit
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Initial Skill DMG (Q)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
		}
		c.QueueCharTask(func() {
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.5}, 7),
				0, 0,
			)
		}, initialHitmark)

		aimid := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins QMid Dummy",
			FlatDmg:    0,
		}
		aifin := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins QFin Dummy",
			FlatDmg:    0,
		}
		if c.MoonsignAscendant && c.lcCloudCheck() {
			// middle phase x4 + final phase x1
			for i, hitmark := range AscendantHitmark {
				if i < 4 {
					// middle phase
					c.Core.QueueAttack(aimid, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				} else {
					// final phase
					c.Core.QueueAttack(aifin, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				}
			}
		} else {
			// middle phase x2 + final phase x1
			for i, hitmark := range NascentHitmark {
				if i < 2 {
					// middle phase
					c.Core.QueueAttack(aimid, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				} else {
					// final phase
					c.Core.QueueAttack(aifin, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				}
			}
		}
		c.SetCD(action.ActionBurst, 20*60)
		c.ConsumeEnergy(7)

		return action.Info{
			Frames:          frames.NewAbilFunc(burstFrames),
			AnimationLength: burstFrames[action.InvalidAction],
			CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
			State:           action.BurstState,
		}, nil
	}
}

func (c *char) lcCloudCheck() bool {
	for _, target := range c.Core.Combat.Enemies() {
		if c.HasLCCloudOn(target) {
			return true
		}
	}
	return false
}
