package skirk

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var (
	chargeFrames     []int
	spchargeFrames   []int
	chargeHitmarks   = []int{30, 40}
	spchargeHitmarks = []int{29, 37, 47}
)

func init() {
	chargeFrames = frames.InitAbilSlice(47)
	chargeFrames[action.ActionDash] = 37

	spchargeFrames = frames.InitAbilSlice(62)
	spchargeFrames[action.ActionDash] = 37
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.onSevenPhaseFlash {
		return c.ChargeAttackSP(p)
	}
	ai := combat.AttackInfo{
		Abil:       "Charge",
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupAyakaExtraAttack,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	singleCharge := func(pos geometry.Point, hitmark int) {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(
				pos,
				nil,
				4,
			),
			hitmark,
			hitmark,
		)
	}

	for i := 0; i < 2; i++ {
		c.Core.Tasks.Add(func() {
			singleCharge(c.Core.Combat.PrimaryTarget().Pos(), 0)
		}, chargeHitmarks[i])
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[len(chargeHitmarks)-1],
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) ChargeAttackSP(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		Abil:       "Seven-Phase Flash Charge",
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupAyakaExtraAttack,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       spcharge[c.TalentLvlSkill()],
	}

	singleCharge := func(pos geometry.Point, hitmark int) {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(
				pos,
				nil,
				4,
			),
			hitmark,
			hitmark,
		)
	}

	for j := 0; j < 3; j++ {
		c.Core.Tasks.Add(func() {
			singleCharge(c.Core.Combat.PrimaryTarget().Pos(), 0)
		}, spchargeHitmarks[j])
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(spchargeFrames),
		AnimationLength: spchargeFrames[action.InvalidAction],
		CanQueueAfter:   spchargeHitmarks[len(chargeHitmarks)-1],
		State:           action.ChargeAttackState,
	}, nil
}
