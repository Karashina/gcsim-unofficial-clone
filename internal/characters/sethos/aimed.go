package sethos

import (
	"errors"
	"fmt"
	"math"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var aimedFrames [][]int
var aimedHitmarks = []int{14, 86}
var aimedWreathFrames []int

const ShadowpiercingHitmark = 358

func init() {
	aimedFrames = make([][]int, 2)

	// Aimed Shot
	aimedFrames[0] = frames.InitAbilSlice(23)
	aimedFrames[0][action.ActionDash] = aimedHitmarks[0]
	aimedFrames[0][action.ActionJump] = aimedHitmarks[0]

	// Fully-Charged Aimed Shot
	aimedFrames[1] = frames.InitAbilSlice(94)
	aimedFrames[1][action.ActionDash] = aimedHitmarks[1]
	aimedFrames[1][action.ActionJump] = aimedHitmarks[1]

	// Fully-Charged Aimed Shot (Wreath Arrow)
	aimedWreathFrames = frames.InitAbilSlice(362)
	aimedWreathFrames[action.ActionDash] = ShadowpiercingHitmark
	aimedWreathFrames[action.ActionJump] = ShadowpiercingHitmark
}

func (c *char) Aimed(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(burstkey) {
		return action.Info{}, errors.New("aim can only be used while burst is not active")
	}

	hold, ok := p["hold"]
	if !ok {
		hold = attacks.AimParamLv2
	}
	switch hold {
	case attacks.AimParamPhys:
	case attacks.AimParamLv1:
	case attacks.AimParamLv2:
		return c.ShadowpiercingAimed(p)
	default:
		return action.Info{}, fmt.Errorf("invalid hold param supplied, got %v", hold)
	}
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}
	weakspot := p["weakspot"]

	ai := combat.AttackInfo{
		ActorIndex:           c.Index,
		Abil:                 "Fully-Charged Aimed Shot",
		AttackTag:            attacks.AttackTagExtra,
		ICDTag:               attacks.ICDTagNone,
		ICDGroup:             attacks.ICDGroupDefault,
		StrikeType:           attacks.StrikeTypePierce,
		Element:              attributes.Electro,
		Durability:           25,
		Mult:                 fullaim[c.TalentLvlAttack()],
		HitWeakPoint:         weakspot == 1,
		HitlagHaltFrames:     0.12 * 60,
		HitlagFactor:         0.01,
		HitlagOnHeadshotOnly: true,
		IsDeployable:         true,
	}
	if hold < attacks.AimParamLv1 {
		ai.Abil = "Aimed Shot"
		ai.Element = attributes.Physical
		ai.Mult = aim[c.TalentLvlAttack()]
	}
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			geometry.Point{Y: -0.5},
			0.1,
			1,
		),
		aimedHitmarks[hold],
		aimedHitmarks[hold]+travel,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(aimedFrames[hold]),
		AnimationLength: aimedFrames[hold][action.InvalidAction],
		CanQueueAfter:   aimedHitmarks[hold],
		State:           action.AimState,
	}, nil
}

func (c *char) ShadowpiercingAimed(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	weakspot := p["weakspot"]

	skip := 0
	energyconsume := 0
	if c.Base.Ascension >= 1 {
		for i := 0; i < 20; i++ {
			energyconsume++
			skip += 17
			if energyconsume == int(math.Ceil(c.Energy)) {
				break
			}
		}
		c.AddEnergy("sethos-aim-drain", float64(-1*energyconsume))
		if !c.StatusIsActive(c6icdkey) && c.Base.Cons >= 6 {
			c.AddEnergy("sethos-c6-refund", float64(energyconsume))
			c.AddStatus(c6icdkey, 15*60, true)
		}
		if energyconsume > 0 && c.Base.Cons >= 2 {
			c.AddStatus(c2aimkey, 600, true)
		}
	}

	if skip > ShadowpiercingHitmark {
		skip = ShadowpiercingHitmark
	}

	if c.a4stacks >= 4 {
		c.QueueCharTask(c.removea4, 5*60)
		c.QueueCharTask(c.reseta4, 15*60)
	}
	if c.a4stacks >= 1 {
		c.a4buff = 7 * c.Stat(attributes.EM)
		c.a4stacks--
	}

	ai := combat.AttackInfo{
		ActorIndex:           c.Index,
		Abil:                 "Shadowpiercing Shot DMG",
		AttackTag:            attacks.AttackTagExtra,
		ICDTag:               attacks.ICDTagNone,
		ICDGroup:             attacks.ICDGroupDefault,
		StrikeType:           attacks.StrikeTypePierce,
		Element:              attributes.Electro,
		Durability:           50,
		Mult:                 shadowpiercing[c.TalentLvlAttack()],
		FlatDmg:              shadowpiercingem[c.TalentLvlAttack()]*c.Stat(attributes.EM) + c.a4buff,
		HitWeakPoint:         weakspot == 0,
		HitlagHaltFrames:     0.12 * 60,
		HitlagFactor:         0.01,
		HitlagOnHeadshotOnly: false,
		IsDeployable:         true,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			geometry.Point{Y: -0.5},
			0.1,
			1,
		),
		ShadowpiercingHitmark-skip,
		ShadowpiercingHitmark+travel-travel-skip,
		c.c4cb(),
	)

	return action.Info{
		Frames:          func(next action.Action) int { return aimedWreathFrames[next] - skip },
		AnimationLength: aimedWreathFrames[action.InvalidAction] - skip,
		CanQueueAfter:   ShadowpiercingHitmark - skip,
		State:           action.AimState,
	}, nil
}
