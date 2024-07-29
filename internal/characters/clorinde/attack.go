package clorinde

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	swiftFrames           []int
	impaleFrames          []int
	attackFrames          [][]int
	attackHitmarks        = [][]int{{20}, {11}, {34, 38}, {13, 21, 28}, {25}}
	attackHitlagHaltFrame = [][]float64{{0.03}, {0.03}, {0, 0.03}, {0, 0, 0.03}, {0.08}}
	attackDefHalt         = [][]bool{{true}, {true}, {false, true}, {false, false, true}, {true}}
	attackHitboxes        = [][]float64{{1.7}, {1.7}, {1.6, 2.8}, {2, 2, 2.6}, {6, 2}}
	attackOffsets         = []float64{0.6, 0.8, 0.3, -0.2, 0.6}
)

const normalHitNum = 5
const swiftHitmark = 5

func init() {
	// NA cancels
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 30) // N1 -> CA
	attackFrames[0][action.ActionAttack] = 27                                // N1 -> N2

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 14) // N2 -> N3

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 45) // N3 -> N4

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][2], 34) // N4 -> N5

	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 70) // N5 -> N1
	attackFrames[4][action.ActionCharge] = 500                               // N5 -> CA, TODO: this action is illegal; need better way to handle it

	// NA (in skill) -> x
	swiftFrames = frames.InitNormalCancelSlice(swiftHitmark, 17)
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillBuffKey) {
		return c.SwiftHunt(p)
	}

	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			ActorIndex:         c.Index,
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
		if c.NormalCounter >= 2 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter]},
				attackHitboxes[c.NormalCounter][0],
				attackHitboxes[c.NormalCounter][1],
			)
		}
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	// normal state
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) SwiftHunt(p map[string]int) (action.Info, error) {

	bolswift := c.bollevel - 1
	if bolswift <= 0 {
		bolswift = 0
	}

	ai := combat.AttackInfo{
		Abil:               fmt.Sprintf("Swift Hunt %v", c.NormalCounter),
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypePierce,
		Element:            attributes.Electro,
		Durability:         25,
		Mult:               swift[bolswift][c.TalentLvlSkill()],
		FlatDmg:            c.a1buff,
		HitlagFactor:       0,
		HitlagHaltFrames:   0,
		CanBeDefenseHalted: false,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 8, 7),
		0,
		swiftHitmark,
		c.particleCB,
	)

	if c.CurrentHPDebt() < c.MaxHP() {
		c.ModifyHPDebtByAmount(0.35 * c.MaxHP())
	}

	c.skillAligned()
	defer c.AdvanceNormalIndex()
	atkspd := c.Stat(attributes.AtkSpd)
	c.c1()

	return action.Info{
		Frames: func(next action.Action) int {
			return frames.AtkSpdAdjust(swiftFrames[next], atkspd)
		},
		AnimationLength: swiftFrames[action.InvalidAction],
		CanQueueAfter:   swiftHitmark,
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) ImpaletheNight(p map[string]int) (action.Info, error) {

	bolimpale := c.bollevel

	impaleHitmark := 8
	impaleFrames = frames.InitNormalCancelSlice(impaleHitmark, 24)

	if bolimpale >= 2 {
		impaleHitmark = 7
		impaleFrames = frames.InitNormalCancelSlice(impaleHitmark, 19)
	}

	ai := combat.AttackInfo{
		Abil:               fmt.Sprintf("Impale the Night %v", c.NormalCounter),
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Electro,
		Durability:         25,
		Mult:               impale[bolimpale][c.TalentLvlSkill()],
		FlatDmg:            c.a1buff,
		HitlagFactor:       0,
		HitlagHaltFrames:   0,
		CanBeDefenseHalted: false,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 8, 7),
		0,
		impaleHitmark,
		c.particleCB,
	)
	c.c1()
	if bolimpale >= 2 {
		if c.Base.Cons >= 6 {
			c.c6()
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 8, 7),
			0,
			impaleHitmark+7,
			c.particleCB,
		)
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 8, 7),
			0,
			impaleHitmark+12,
			c.particleCB,
		)
	}

	if c.CurrentHPDebt() < c.MaxHP() && c.CurrentHPDebt() > 0 {
		amt := 1.04 * c.CurrentHPDebt()
		// call the template healing method directly to bypass Heal override
		c.Character.Heal(&info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "clorinde-skill-heal",
			Src:     amt,
			Bonus:   c.Stat(attributes.Heal),
		})
	}
	if c.CurrentHPDebt() >= c.MaxHP() {
		amt := 1.10 * c.CurrentHPDebt()
		c.Character.Heal(&info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "clorinde-skill-heal",
			Src:     amt,
			Bonus:   c.Stat(attributes.Heal),
		})
	}
	c.skillAligned()

	defer c.AdvanceNormalIndex()
	atkspd := c.Stat(attributes.AtkSpd)

	return action.Info{
		Frames: func(next action.Action) int {
			return frames.AtkSpdAdjust(impaleFrames[next], atkspd)
		},
		AnimationLength: impaleFrames[action.InvalidAction],
		CanQueueAfter:   impaleHitmark,
		State:           action.NormalAttackState,
	}, nil
}
