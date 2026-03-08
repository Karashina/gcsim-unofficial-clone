package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	chargeFrames        []int
	sturmChargeFrames   []int
	azureDevourFrames   []int
	chargeHitmarks      = []int{41, 42}
	sturmChargeHitmarks = []int{36, 37}
	azureDevourHitmarks = []int{39, 40, 62, 63}
)

func init() {
	chargeFrames = frames.InitAbilSlice(42)
	chargeFrames[action.ActionAttack] = 65

	sturmChargeFrames = frames.InitAbilSlice(44)
	sturmChargeFrames[action.ActionAttack] = 65

	azureDevourFrames = frames.InitAbilSlice(62)
	azureDevourFrames[action.ActionAttack] = 65
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.sturmActive {
		// Update FWA charges from CD timer
		c.updateFWACharges()
		// In S&D mode: if FWA charges available, perform Azure Devour
		if c.fwaCharges > 0 {
			return c.azureDevour(p)
		}
		// C6: if window is active from FWA, perform Azure Devour without charges
		if c.Base.Cons >= 6 && c.StatusIsActive(c6FWAWindowKey) {
			return c.azureDevour(p)
		}
		// Otherwise, perform S&D charged attack
		return c.sturmCharge(p)
	}
	return c.normalCharge(p)
}

// normalCharge handles the basic charged attack (outside S&D)
func (c *char) normalCharge(p map[string]int) (action.Info, error) {
	// 2 sub-hits, both Physical
	subHits := [][]float64{charge_a, charge_b}
	for i, mult := range subHits {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Charged Attack (Hit %v)", i+1),
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             attacks.ICDTagExtraAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           120.0,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.1 * 60,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3, 4),
			chargeHitmarks[i], chargeHitmarks[i],
		)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}

// sturmCharge handles the S&D charged attack (no FWA charges available)
func (c *char) sturmCharge(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()

	// S&D Charged: 1st=Other, 2nd=Anemo
	type hitInfo struct {
		mult    float64
		element attributes.Element
		icdTag  attacks.ICDTag
	}

	hits := []hitInfo{
		{sturmCA_a[lvl], c.otherElement, attacks.ICDTagVarkaCAOther},
		{sturmCA_b[lvl], attributes.Anemo, attacks.ICDTagVarkaCAAnemo},
	}

	if !c.hasOtherEle {
		hits[0].element = attributes.Anemo
	}

	for i, h := range hits {
		mult := h.mult
		// Apply A1 multiplier
		if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
			mult *= c.a1MultFactor
		}

		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Sturm und Drang Charged (Hit %v)", i+1),
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             h.icdTag,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           120.0,
			Element:            h.element,
			Durability:         25,
			Mult:               mult,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.1 * 60,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3, 4),
			sturmChargeHitmarks[i], sturmChargeHitmarks[i],
		)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(sturmChargeFrames),
		AnimationLength: sturmChargeFrames[action.InvalidAction],
		CanQueueAfter:   sturmChargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}

// azureDevour handles the special charged attack that consumes FWA charges
func (c *char) azureDevour(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()

	// C6: check if we should not consume charges
	consumeCharge := true
	if c.Base.Cons >= 6 {
		if c.StatusIsActive(c6FWAWindowKey) {
			// After FWA, tap charged triggers additional Azure Devour without consuming charges
			consumeCharge = false
			c.DeleteStatus(c6FWAWindowKey)
		} else if c.StatusIsActive(c6AzureWindowKey) {
			consumeCharge = false
			c.DeleteStatus(c6AzureWindowKey)
		}
	}

	if consumeCharge {
		c.fwaCharges--
	}

	// C1: Lyrical Libation effect - first Azure Devour after entering S&D deals 200% DMG
	c1Mult := 1.0
	if c.Base.Cons >= 1 && c.StatusIsActive(c1LyricalKey) {
		c1Mult = 2.0
		c.DeleteStatus(c1LyricalKey)
	}

	// Azure Devour: 4 hits
	// 1st: Other, 2nd: Anemo, 3rd: Other, 4th: Anemo
	// All use ICDTagVarkaCAOther
	type hitInfo struct {
		mult    float64
		element attributes.Element
	}

	hits := []hitInfo{
		{azureOther[lvl], c.otherElement},
		{azureAnemo[lvl], attributes.Anemo},
		{azureOther[lvl], c.otherElement},
		{azureAnemo[lvl], attributes.Anemo},
	}

	if !c.hasOtherEle {
		hits[0].element = attributes.Anemo
		hits[2].element = attributes.Anemo
	}

	for i, h := range hits {
		mult := h.mult
		// Apply A1 multiplier to Azure Devour
		if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
			mult *= c.a1MultFactor
		}
		// Apply C1 Lyrical Libation multiplier
		mult *= c1Mult

		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Azure Devour (Hit %v)", i+1),
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             attacks.ICDTagVarkaCAOther, // All 4 hits share this tag
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           120.0,
			Element:            h.element,
			Durability:         25,
			Mult:               mult,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.08 * 60,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3, 4),
			azureDevourHitmarks[i], azureDevourHitmarks[i],
		)
	}

	// C2: Additional Anemo strike equal to 800% ATK
	if c.Base.Cons >= 2 {
		c.c2Strike(azureDevourHitmarks[3] + 4)
	}

	// C6: after Azure Devour, open window for additional FWA
	// Only set window when this was a normal Azure Devour (not a C6 chain trigger)
	if c.Base.Cons >= 6 && consumeCharge {
		c.AddStatus(c6AzureWindowKey, 60, true) // ~1s window
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(azureDevourFrames),
		AnimationLength: azureDevourFrames[action.InvalidAction],
		CanQueueAfter:   azureDevourHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}
