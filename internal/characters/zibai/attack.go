package zibai

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
	attackHitmarks        = [][]int{{12}, {14}, {17, 30}, {27}}
	attackHitlagHaltFrame = [][]float64{{0.1}, {0.1}, {0.06, 0.03}, {0.06}}
	attackDefHalt         = [][]bool{{true}, {true}, {false, true}, {true}}
	attackHitboxes        = [][]float64{{2.0}, {2.0}, {2.0, 2.0}, {2.5}}
	attackOffsets         = []float64{0.5, 0.5, 0.5, 1.0}
)

const normalHitNum = 4

// N4 additional hit delay (distinct from Spirit Steed's Stride)
const n4AdditionalHitmark = 35

func init() {
	attackFrames = make([][]int, normalHitNum)

	// N1 -> x
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 24)
	attackFrames[0][action.ActionAttack] = 27

	// N2 -> x
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 25)
	attackFrames[1][action.ActionAttack] = 25

	// N3 -> x (2 hits)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 30)
	attackFrames[2][action.ActionAttack] = 42

	// N4 -> x
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 40)
	attackFrames[3][action.ActionAttack] = 61
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	// Check if in Lunar Phase Shift mode
	if c.lunarPhaseShiftActive {
		return c.lunarPhaseShiftAttack(p)
	}

	// Normal attack (Physical)
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
		)

		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

// lunarPhaseShiftAttack handles attacks during Lunar Phase Shift mode (Geo, DEF-scaling)
func (c *char) lunarPhaseShiftAttack(p map[string]int) (action.Info, error) {
	// C4: Normal Attack sequence will not reset in Lunar Phase Shift mode
	if c.Base.Cons >= 4 {
		c.NormalCounter = c.savedNormalCounter
	}

	for i, mult := range lunarPhaseShiftAttack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Lunar Phase Shift N%v", c.NormalCounter+1),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Geo,
			Durability:         25,
			UseDef:             true,
			Mult:               mult[c.TalentLvlSkill()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
		)

		hitmark := attackHitmarks[c.NormalCounter][i] - 2
		currentN := c.NormalCounter

		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, c.radianceGainCB, c.particleCB)

			// N4 additional hit when Moonsign is Ascendant Gleam
			if currentN == 3 && c.isMoonsignAscendant() {
				c.queueN4AdditionalHit()
			}
		}, hitmark)
	}

	// C4: After Spirit Steed's Stride hits, next N4 additional attack deals 250% damage
	// (checked in queueN4AdditionalHit)

	defer func() {
		c.AdvanceNormalIndex()
		if c.Base.Cons >= 4 {
			c.savedNormalCounter = c.NormalCounter
		}
	}()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

// queueN4AdditionalHit queues the additional Lunar-Crystallize damage on N4
func (c *char) queueN4AdditionalHit() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Lunar Phase Shift N4 Additional (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// C4 bonus
	mult := 1.0

	// C4: Scattermoon Splendor - 250% of original damage
	if c.c4ScattermoonUsed {
		mult = 2.5
		c.c4ScattermoonUsed = false
	}

	// DEF scaling with Lunar-Crystallize formula (includes ×1.6 LCrs base multiplier)
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * 1.6 * lunarPhaseShift4Additional[c.TalentLvlSkill()]
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * mult * (1 + c.LCrsBaseReactBonus(ai)) * (1 + emBonus + c.LCrsReactBonus(ai))

	ai.FlatDmg *= (1 + c.ElevationBonus(ai))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3)
	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai, snap, ap, 27)
	}, n4AdditionalHitmark)
}

// radianceGainCB callback for gaining Phase Shift Radiance on hit
func (c *char) radianceGainCB(a combat.AttackCB) {
	if !c.lunarPhaseShiftActive {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(radianceNormalICDKey) {
		return
	}
	c.AddStatus(radianceNormalICDKey, radianceNormalICD, false)
	c.addPhaseShiftRadiance(radianceNormalGain)
}
