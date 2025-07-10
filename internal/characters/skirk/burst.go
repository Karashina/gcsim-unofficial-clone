package skirk

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var (
	burstFrames       []int
	skillburstFrames  []int
	burstHitmarks     = []int{111, 117, 121, 128, 134}
	burstfinalhitmark = 160
)

const (
	burstCD  = 15 * 60
	burstkey = "skirk-burst"
)

func init() {
	burstFrames = frames.InitAbilSlice(109)
	skillburstFrames = frames.InitAbilSlice(78)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.onSevenPhaseFlash {
		c.c6("burst")
		c.SetCD(action.ActionBurst, burstCD)
		return c.spBurst()
	}

	excess := max(c.serpentsSubtlety-50, 0)
	burstbuff := min(excess, 12)
	if c.Base.Cons >= 2 {
		excess = max(c.serpentsSubtlety+10-50, 0)
		burstbuff = min(excess, 12)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Havoc: Ruin (Slash DMG)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()] + burstbuff*burstDMGUp[c.TalentLvlBurst()],
	}
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1}, 11.2, 9)
	for _, v := range burstHitmarks {
		// TODO: what's the size of this??
		c.Core.QueueAttack(ai, ap, v, v, c.particleCB)
	}

	aifinal := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Havoc: Ruin (Final Slash DMG)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       burstlast[c.TalentLvlBurst()] + burstbuff*burstDMGUp[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(aifinal, ap, burstfinalhitmark, burstfinalhitmark, c.particleCB)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) spBurst() (action.Info, error) {
	c.burstbuff()
	c.a1(false)
	c.AddStatus(burstkey, -1, true)
	c.burstbuffcount = 10
	return action.Info{
		Frames:          frames.NewAbilFunc(skillburstFrames),
		AnimationLength: skillburstFrames[action.InvalidAction],
		CanQueueAfter:   skillburstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) burstbuff() {
	m := make([]float64, attributes.EndStatType)
	switch c.voidrift {
	case 1:
		m[attributes.DmgP] = 0.06 + 0.006*float64(c.TalentLvlBurst())
	case 2:
		m[attributes.DmgP] = 0.08 + 0.008*float64(c.TalentLvlBurst())
	case 3:
		m[attributes.DmgP] = 0.10 + 0.01*float64(c.TalentLvlBurst())
	default:
		m[attributes.DmgP] = 0.03 + 0.005*float64(c.TalentLvlBurst())
	}
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("skirk-burst-buff", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if !c.StatusIsActive(burstkey) {
				return nil, false
			}
			if !c.onSevenPhaseFlash {
				return nil, false
			}
			if atk.Info.AttackTag != attacks.AttackTagNormal {
				return nil, false
			}
			if c.burstbuffcount <= 0 {
				return nil, false
			}
			c.burstbuffcount--
			return m, true
		},
	})
}
