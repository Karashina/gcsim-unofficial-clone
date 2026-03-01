package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	attackFrames          [][]int
	attackHitmarks        = []int{20, 19, 26, 15, 45}
	attackPoiseDMG        = []float64{108.1, 80.0, 100.0, 100.0, 130.0}
	attackHitlagHaltFrame = []float64{.1, 0.08, 0.06, 0.08, 0.08}
	attackHitboxes        = [][]float64{{2.5}, {2.5, 3}, {2.5, 3}, {2.5, 3}, {3, 3.5}}
	attackOffsets         = []float64{0.5, -0.5, -0.5, 0.5, -0.5}
	attackFanAngles       = []float64{300, 360, 360, 300, 360}

	// S&D attack frames
	sturmFrames          [][]int
	sturmHitmarks        = []int{19, 16, 28, 14, 40}
	sturmPoiseDMG        = []float64{108.1, 80.0, 100.0, 100.0, 130.0}
	sturmHitlagHaltFrame = []float64{0.03, 0.06, 0.12, 0.1, 0.12}
	sturmSecondHitDelays = []int{8, 8, 8, 8, 10} // delay between 1st and 2nd sub-hit (S&D)

	// Delay (frames) from 1st to 2nd sub-hit for each normal segment.
	// Index 0 is unused (N1 is single-hit); N2-N5 are indices 1-4.
	attackSecondHitDelays = []int{0, 11, 17, 6, 1}
)

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 20)
	attackFrames[0][action.ActionAttack] = 29
	attackFrames[0][action.ActionCharge] = 21

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 30)
	attackFrames[1][action.ActionAttack] = 33

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 43)
	attackFrames[2][action.ActionAttack] = 54

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3], 21)
	attackFrames[3][action.ActionAttack] = 35

	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4], 46)
	attackFrames[4][action.ActionAttack] = 81

	sturmFrames = make([][]int, sturmNormalHitNum)
	sturmFrames[0] = frames.InitNormalCancelSlice(sturmHitmarks[0], 27)
	sturmFrames[0][action.ActionAttack] = 25
	sturmFrames[0][action.ActionCharge] = 34

	sturmFrames[1] = frames.InitNormalCancelSlice(sturmHitmarks[1], 27)
	sturmFrames[1][action.ActionAttack] = 29

	sturmFrames[2] = frames.InitNormalCancelSlice(sturmHitmarks[2], 44)
	sturmFrames[2][action.ActionAttack] = 54

	sturmFrames[3] = frames.InitNormalCancelSlice(sturmHitmarks[3], 21)
	sturmFrames[3][action.ActionAttack] = 39

	sturmFrames[4] = frames.InitNormalCancelSlice(sturmHitmarks[4], 41)
	sturmFrames[4][action.ActionAttack] = 66
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.sturmActive {
		return c.sturmAttack(p)
	}
	return c.normalAttack(p)
}

// normalAttack handles the Physical normal attack sequence
func (c *char) normalAttack(p map[string]int) (action.Info, error) {
	// N1 is a single hit
	if c.NormalCounter == 0 {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           attackPoiseDMG[0],
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               attack_1[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[0] * 60,
			CanBeDefenseHalted: true,
		}
		ap := combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[0]},
			attackHitboxes[0][0],
			attackFanAngles[0],
		)
		c.Core.QueueAttack(ai, ap, attackHitmarks[0], attackHitmarks[0])
	} else {
		// N2-N5 are multi-hit (2 sub-hits each)
		idx := c.NormalCounter
		multiIdx := idx - 1 // maps N2→0, N3→1, N4→2, N5→3
		subHits := attackMulti[multiIdx+1]

		for i, mult := range subHits {
			ai := combat.AttackInfo{
				ActorIndex:         c.Index,
				Abil:               fmt.Sprintf("Normal %v (Hit %v)", c.NormalCounter, i+1),
				AttackTag:          attacks.AttackTagNormal,
				ICDTag:             attacks.ICDTagNormalAttack,
				ICDGroup:           attacks.ICDGroupDefault,
				StrikeType:         attacks.StrikeTypeBlunt,
				PoiseDMG:           attackPoiseDMG[idx],
				Element:            attributes.Physical,
				Durability:         25,
				Mult:               mult[c.TalentLvlAttack()],
				HitlagFactor:       0.01,
				HitlagHaltFrames:   attackHitlagHaltFrame[idx] * 60,
				CanBeDefenseHalted: true,
			}
			delay := attackHitmarks[idx]
			if i == 1 {
				delay += attackSecondHitDelays[idx]
			}
			ap := combat.NewCircleHitOnTargetFanAngle(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[idx]},
				attackHitboxes[idx][0],
				attackFanAngles[idx],
			)
			if len(attackHitboxes[idx]) > 1 {
				ap = combat.NewBoxHitOnTarget(
					c.Core.Combat.Player(),
					geometry.Point{Y: attackOffsets[idx]},
					attackHitboxes[idx][0],
					attackHitboxes[idx][1],
				)
			}
			c.Core.QueueAttack(ai, ap, delay, delay)
		}
	}

	defer func() {
		c.AdvanceNormalIndex()
		c.savedNormalCounter = c.NormalCounter
	}()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

// sturmAttack handles the Sturm und Drang enhanced normal attacks
func (c *char) sturmAttack(p map[string]int) (action.Info, error) {
	idx := c.NormalCounter

	// S&D elemental assignment per hit (from spec):
	// N1: Other only
	// N2: Anemo + Other
	// N3: Anemo + Other
	// N4: Other + Anemo
	// N5: Other + Anemo
	type hitInfo struct {
		mult    []float64
		element attributes.Element
		icdTag  attacks.ICDTag
	}

	var hits []hitInfo
	lvl := c.TalentLvlSkill()

	switch idx {
	case 0: // N1: Other only (single hit)
		hits = []hitInfo{
			{sturmN1, c.otherElement, attacks.ICDTagVarkaNAOther},
		}
	case 1: // N2: 1st=Anemo, 2nd=Other
		hits = []hitInfo{
			{sturmN2a, attributes.Anemo, attacks.ICDTagVarkaNAAnemo},
			{sturmN2b, c.otherElement, attacks.ICDTagVarkaNAOther},
		}
	case 2: // N3: 1st=Anemo, 2nd=Other
		hits = []hitInfo{
			{sturmN3a, attributes.Anemo, attacks.ICDTagVarkaNAAnemo},
			{sturmN3b, c.otherElement, attacks.ICDTagVarkaNAOther},
		}
	case 3: // N4: 1st=Other, 2nd=Anemo
		hits = []hitInfo{
			{sturmN4a, c.otherElement, attacks.ICDTagVarkaNAOther},
			{sturmN4b, attributes.Anemo, attacks.ICDTagVarkaNAAnemo},
		}
	case 4: // N5: 1st=Other, 2nd=Anemo
		hits = []hitInfo{
			{sturmN5a, c.otherElement, attacks.ICDTagVarkaNAOther},
			{sturmN5b, attributes.Anemo, attacks.ICDTagVarkaNAAnemo},
		}
	}

	// If no other element, all become Anemo
	if !c.hasOtherEle {
		for i := range hits {
			hits[i].element = attributes.Anemo
		}
	}

	for i, h := range hits {
		mult := h.mult[lvl]
		// Apply A1 multiplier to S&D attacks
		if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
			mult *= c.a1MultFactor
		}

		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Sturm und Drang %v (Hit %v)", idx+1, i+1),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             h.icdTag,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           sturmPoiseDMG[idx],
			Element:            h.element,
			Durability:         25,
			Mult:               mult,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   sturmHitlagHaltFrame[idx] * 60,
			CanBeDefenseHalted: true,
		}

		delay := sturmHitmarks[idx]
		if i > 0 {
			delay += sturmSecondHitDelays[idx]
		}

		ap := combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[idx]},
			attackHitboxes[idx][0],
			attackFanAngles[idx],
		)
		if len(attackHitboxes[idx]) > 1 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[idx]},
				attackHitboxes[idx][0],
				attackHitboxes[idx][1],
			)
		}

		// Add CD reduction callback for first hit only
		var cb combat.AttackCBFunc
		if i == 0 {
			cb = c.sturmNAHitCB
		}
		c.Core.QueueAttack(ai, ap, delay, delay, cb)
	}

	defer func() {
		c.AdvanceNormalIndex()
		c.savedNormalCounter = c.NormalCounter
	}()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, sturmFrames),
		AnimationLength: sturmFrames[idx][action.InvalidAction],
		CanQueueAfter:   sturmHitmarks[idx],
		State:           action.NormalAttackState,
	}, nil
}

// sturmNAHitCB is the callback when S&D normal attacks hit an enemy
// Reduces Four Winds' Ascension CD by 0.5s (or 1s with Hexerei)
func (c *char) sturmNAHitCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !c.sturmActive {
		return
	}
	if c.cdReductionCount >= c.cdReductionMax {
		return
	}

	c.cdReductionCount++
	reduction := c.getCDReductionAmount()

	// Reduce FWA charge cooldown timer
	c.fwaCDEndFrame -= reduction

	// Immediately check if charges became ready
	c.updateFWACharges()
}
