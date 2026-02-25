package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var burstFrames []int

const (
	burstHitmark1 = 97
	burstHitmark2 = 113
)

func init() {
	burstFrames = frames.InitAbilSlice(103)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 1st Hit DMG (Geo, DEF-scaling, Durability 25, No ICD)
	ai1 := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Tri-Sphere Eminence 1-Hit",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
		UseDef:     true,
		Mult:       burstInitial[c.TalentLvlBurst()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai1, ap, 0, 0)
	}, burstHitmark1)

	// 2nd Hit DMG (Lunar-Crystallize Reaction DMG)
	ai2 := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Tri-Sphere Eminence 2-Hit (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// Calculate 2nd hit multiplier with bonuses
	secondHitMult := 1.6 * burstSecond[c.TalentLvlBurst()]

	// DEF scaling with Lunar-Crystallize formula
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * secondHitMult
	emBonus := (6 * em) / (2000 + em)
	ai2.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai2)) * (1 + emBonus + c.LCrsReactBonus(ai2))
	ai2.FlatDmg *= (1 + c.ElevationBonus(ai2))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai2, snap, ap, 0)
	}, burstHitmark2)

	// If in Lunar Phase Shift mode, extend duration by 1.7s
	if c.lunarPhaseShiftActive {
		c.extendLunarPhaseShift(102) // 1.7s = 102 frames
	}

	// Set cooldown (15s)
	c.SetCDWithDelay(action.ActionBurst, 15*60, burstHitmark1)
	// Consume energy
	c.ConsumeEnergy(4)

	c.Core.Log.NewEvent("Zibai uses Tri-Sphere Eminence", glog.LogCharacterEvent, c.Index).
		Write("lunar_phase_shift_extended", c.lunarPhaseShiftActive)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}
