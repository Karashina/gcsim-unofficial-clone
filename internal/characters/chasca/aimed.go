package chasca

import (
	"strconv"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

var (
	aimedFrames  [][]int
	loadhitmarks = []int{21, 38, 56, 73, 90, 108}
)

func init() {
	aimedFrames = make([][]int, 8)

	// Multitarget Fire
	aimedFrames[0] = frames.InitAbilSlice(41)
	aimedFrames[0][action.ActionAim] = 41

	aimedFrames[1] = frames.InitAbilSlice(58)
	aimedFrames[1][action.ActionAim] = 58

	aimedFrames[2] = frames.InitAbilSlice(76)
	aimedFrames[2][action.ActionAim] = 76

	aimedFrames[3] = frames.InitAbilSlice(93)
	aimedFrames[3][action.ActionAim] = 93

	aimedFrames[4] = frames.InitAbilSlice(110)
	aimedFrames[4][action.ActionAim] = 110

	aimedFrames[5] = frames.InitAbilSlice(132)
	aimedFrames[5][action.ActionAim] = 132

	// Multitarget Fire (C6 Active)
	aimedFrames[6] = frames.InitAbilSlice(44)
	aimedFrames[6][action.ActionAim] = 44

	// Fully-Charged Aimed Shot
	aimedFrames[7] = frames.InitAbilSlice(25)
	aimedFrames[7][action.ActionDash] = 86
	aimedFrames[7][action.ActionJump] = 86

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
		Frames:          frames.NewAbilFunc(aimedFrames[7]),
		AnimationLength: aimedFrames[7][action.InvalidAction],
		CanQueueAfter:   86,
		State:           action.AimState,
	}, nil
}

func (c *char) Shadowhunt(p map[string]int) action.Info {

	c.DeleteStatus(c2ICDKey)

	framekey := 0

	c.anemoremaining = c.anemoCount

	usedElements := make(map[attributes.Element]bool)

	for k := 0; k < 6; k++ {
		c.Shells[k] = c.selectElement(k, usedElements)
		if c.Base.Cons >= 1 && k == 2 && c.Shells[k] != attributes.Anemo {
			c.Shells[1] = c.ElementSlot[c.Core.Rand.Intn(c.typeCount)]
		}
	}

	firekey := false
	c.loadednum = 0
	if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
		framekey = 6
		c.loadednum = 6
		c.fireShells(loadhitmarks[0], c.loadednum)
		c.QueueCharTask(c.removec6, aimedFrames[2][action.ActionAim])
	}
	for l := 0; l < 6; l++ {
		c.QueueCharTask(func() {
			if firekey {
				return
			}
			if !c.nightsoulState.HasBlessing() {
				c.fireShells(loadhitmarks[c.loadednum-1], c.loadednum)
				firekey = true
				return
			}
			c.loadednum++
			c.Core.Log.NewEvent("Chasca Shells Load : "+strconv.Itoa(c.loadednum), glog.LogCharacterEvent, c.Index)
			framekey = c.loadednum - 1
			if c.loadednum == 6 {
				c.fireShells(loadhitmarks[c.loadednum-1], c.loadednum)
				firekey = true
			}
		}, loadhitmarks[l])
	}

	return action.Info{
		Frames: func(next action.Action) int {
			return aimedFrames[framekey][next]
		},
		AnimationLength: 1200, // there is no upper limit on the duration of the CA
		CanQueueAfter:   aimedFrames[1][action.ActionAim],
		State:           action.AimState,
		OnRemoved: func(next action.AnimationState) {
			// need to calculate correct swap cd in case of early cancel
			switch next {
			case action.SkillState, action.BurstState, action.DashState, action.JumpState:
				c.Core.Player.SwapCD = max(player.SwapCDFrames-(c.Core.F-c.lastSwap), 0)
			}
		},
	}
}

func (c *char) fireShells(loadtime int, shellsNum int) {
	aimArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
	target := combat.NewSingleTargetHit(c.Core.Combat.RandomEnemyWithinArea(aimArea, nil).Key())
	if shellsNum == 0 {
		return
	}
	for j := 0; j < shellsNum; j++ {
		i := shellsNum - 1 - j
		c.QueueCharTask(func() {
			c.Core.Log.NewEvent("Chasca Shells Fire : "+strconv.Itoa(shellsNum), glog.LogCharacterEvent, c.Index)
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

		}, loadtime+3*(1+j))
	}
}

func (c *char) selectElement(idx int, usedElements map[attributes.Element]bool) attributes.Element {
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
	case 3, 4, 5:
		if c.anemoremaining > 0 {
			c.anemoremaining--
			return attributes.Anemo
		} else {
			var availableElements []attributes.Element
			for _, elem := range c.ElementSlot {
				if !usedElements[elem] {
					availableElements = append(availableElements, elem)
				}
			}
			if len(availableElements) > 0 {
				selectedElement := availableElements[c.Core.Rand.Intn(len(availableElements))]
				usedElements[selectedElement] = true
				return selectedElement
			} else {
				return attributes.Anemo
			}
		}
	default:
		return attributes.Anemo
	}
}
