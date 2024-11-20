package chasca

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var (
	aimedFrames    [][]int
	aimedHitmark   = 114
	aimedHitmarkC6 = 20
)

func init() {
	aimedFrames = make([][]int, 3)

	// Fully-Charged Aimed Shot
	aimedFrames[0] = frames.InitAbilSlice(25)
	aimedFrames[0][action.ActionDash] = 86
	aimedFrames[0][action.ActionJump] = 86

	// Multitarget Fire
	aimedFrames[1] = frames.InitAbilSlice(128)
	aimedFrames[1][action.ActionAim] = 128

	// Multitarget Fire (C6 Active)
	aimedFrames[2] = frames.InitAbilSlice(44)
	aimedFrames[2][action.ActionAim] = 44
}

func (c *char) Aimed(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		return c.Shadowhunt(p), nil
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
		Element:              attributes.Anemo,
		Durability:           25,
		Mult:                 fullaim[c.TalentLvlAttack()],
		HitWeakPoint:         weakspot == 1,
		HitlagHaltFrames:     .12 * 60,
		HitlagOnHeadshotOnly: true,
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
		86,
		86+travel,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(aimedFrames[0]),
		AnimationLength: aimedFrames[0][action.InvalidAction],
		CanQueueAfter:   86,
		State:           action.AimState,
	}, nil
}

func (c *char) Shadowhunt(p map[string]int) action.Info {

	c.DeleteStatus(c2ICDKey)

	framekey := 1
	hitmark := aimedHitmark
	if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
		framekey = 2
		hitmark = aimedHitmarkC6
	}

	aimArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
	target := combat.NewSingleTargetHit(c.Core.Combat.RandomEnemyWithinArea(aimArea, nil).Key())

	for k := 0; k < 6; k++ {
		c.Shells[k] = c.selectElement(k)
		if c.Base.Cons >= 1 && k == 2 && c.Shells[k] != attributes.Anemo {
			c.Shells[1] = c.ElementSlot[c.Core.Rand.Intn(c.typeCount)]
		}
	}

	for j := 0; j < 6; j++ {
		i := 5 - j
		c.QueueCharTask(func() {
			ai := combat.AttackInfo{
				ActorIndex:     c.Index,
				Abil:           "Shadowhunt Shell DMG (E)",
				AttackTag:      attacks.AttackTagExtra,
				AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
				ICDTag:         attacks.ICDTagShadowhuntShell,
				ICDGroup:       attacks.ICDGroupShadowhuntShell,
				StrikeType:     attacks.StrikeTypeDefault,
				Mult:           shadowhunt[c.TalentLvlSkill()],
				Element:        c.Shells[i],
				Durability:     25,
			}

			if c.Shells[i] != attributes.Anemo {
				ai.Abil = "Shining Shadowhunt Shell DMG (E)"
				ai.Mult = shiningshadowhunt[c.TalentLvlSkill()]
				ai.ICDTag = attacks.ICDTagShiningShadowhuntShell
				ai.ICDGroup = attacks.ICDGroupChascaConvertedShell
			}

			c.Core.QueueAttack(
				ai,
				target,
				0,
				0,
				c.particleCB,
				c.c2CB,
			)

		}, hitmark+3*(1+j))
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(aimedFrames[1]),
		AnimationLength: aimedFrames[framekey][action.InvalidAction],
		CanQueueAfter:   aimedFrames[framekey][action.ActionAim], // earliest cancel
		State:           action.AimState,
	}
}

func (c *char) selectElement(idx int) attributes.Element {
	c.anemoremaining = c.anemoCount
	switch idx {
	case 0:
		return attributes.Anemo
	case 1:
		return attributes.Anemo
	case 2:
		if c.Core.Rand.Float64() < c.a1Prob {
			if c.Base.Cons >= 6 && !c.StatusIsActive(c6ICDKey) {
				c.AddStatus(c6Key, 3*60, true)
				c.AddStatus(c6ICDKey, 3*60, true)
			}
			return c.ElementSlot[c.Core.Rand.Intn(len(c.ElementSlot))]
		} else {
			return attributes.Anemo
		}
	case 3:
		if c.anemoremaining > 0 {
			c.anemoremaining--
			return attributes.Anemo
		} else {
			return c.ElementSlot[c.Core.Rand.Intn(len(c.ElementSlot))]
		}
	case 4:
		if c.anemoremaining > 0 {
			c.anemoremaining--
			return attributes.Anemo
		} else {
			return c.ElementSlot[c.Core.Rand.Intn(len(c.ElementSlot))]
		}
	case 5:
		if c.anemoremaining > 0 {
			c.anemoremaining--
			return attributes.Anemo
		} else {
			return c.ElementSlot[c.Core.Rand.Intn(len(c.ElementSlot))]
		}
	default:
		return attributes.Anemo
	}
}
