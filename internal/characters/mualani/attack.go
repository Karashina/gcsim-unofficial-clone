package mualani

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

var (
	attackFramesNormal  [][]int
	attackReleaseNormal = []int{22, 12, 35}
)

const normalHitNum = 3
const attackFramesE = 46

func init() {
	attackFramesNormal = make([][]int, normalHitNum)

	attackFramesNormal[0] = frames.InitNormalCancelSlice(attackReleaseNormal[0], 32)
	attackFramesNormal[0][action.ActionAttack] = 32
	attackFramesNormal[0][action.ActionCharge] = 32
	attackFramesNormal[0][action.ActionSkill] = 32
	attackFramesNormal[0][action.ActionBurst] = 32
	attackFramesNormal[0][action.ActionDash] = 22

	attackFramesNormal[1] = frames.InitNormalCancelSlice(attackReleaseNormal[1], 24)
	attackFramesNormal[1][action.ActionAttack] = 24
	attackFramesNormal[1][action.ActionCharge] = 24
	attackFramesNormal[1][action.ActionSkill] = 24
	attackFramesNormal[1][action.ActionBurst] = 24
	attackFramesNormal[1][action.ActionDash] = 12
	attackFramesNormal[1][action.ActionJump] = 12
	attackFramesNormal[1][action.ActionSwap] = 12

	attackFramesNormal[2] = frames.InitNormalCancelSlice(attackReleaseNormal[2], 55)
	attackFramesNormal[2][action.ActionAttack] = 55
	attackFramesNormal[2][action.ActionCharge] = 55
	attackFramesNormal[2][action.ActionSkill] = 55
	attackFramesNormal[2][action.ActionBurst] = 55
	attackFramesNormal[2][action.ActionDash] = 35
	attackFramesNormal[2][action.ActionJump] = 35
	attackFramesNormal[2][action.ActionSwap] = 35
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.AttackSharkysBite(p)
	}

	currentNormalCounter := c.NormalCounter

	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
	}

	release := attackReleaseNormal[c.NormalCounter]

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 0.5),
		release,
		release+travel,
	)

	defer c.AdvanceNormalIndex()
	atkspd := c.Stat(attributes.AtkSpd)
	return action.Info{
		Frames: func(next action.Action) int {
			return frames.AtkSpdAdjust(attackFramesNormal[currentNormalCounter][next], atkspd)
		},
		AnimationLength: attackFramesNormal[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackReleaseNormal[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) AttackSharkysBite(p map[string]int) (action.Info, error) {
	c.c1buff = 0
	if c.Base.Cons >= 1 {
		if c.c1count < 1 {
			c.c1buff = 0.66 * c.MaxHP()
			c.c1count++
		}
		if c.Base.Cons >= 6 {
			c.c1buff = 0.66 * c.MaxHP()
		}
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Sharky's Bite DMG",
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    c.MaxHP()*skillBase[c.TalentLvlSkill()] + c.c1buff,
		Alignment:  attacks.AlignmentNightsoul,
	}
	aimissiles := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Shark Missile DMG",
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    c.MaxHP()*skillBase[c.TalentLvlSkill()] + c.c1buff,
		Alignment:  attacks.AlignmentNightsoul,
	}

	enemycount := 0
	area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{X: 0.0, Y: 0.0}, 5)
	enemies := c.Core.Combat.EnemiesWithinArea(area, nil)
	for _, e := range enemies {
		if enemycount < 5 {
			if e.StatusIsActive(skillMarkKey) {
				enemycount++
			}
		}
	}

	hitmark := 9
	SSBframe := 0

	if c.WaveMomentum > 0 {
		ai.FlatDmg += c.MaxHP() * skillWMBonus[c.TalentLvlSkill()] * float64(c.WaveMomentum)
		aimissiles.FlatDmg = ai.FlatDmg
	}

	if c.WaveMomentum >= 3 {
		ai.FlatDmg += c.MaxHP() * skillSSBBonus[c.TalentLvlSkill()]
		aimissiles.FlatDmg = ai.FlatDmg
		ai.Abil = "Sharky's Surging Bite DMG"
		aimissiles.Abil = "Surging Shark Missile DMG"
		hitmark = 45
		SSBframe = 36
	}
	ai.FlatDmg = ai.FlatDmg * max(0.72, 1-(float64(enemycount)-1)*0.14)
	aimissiles.FlatDmg = ai.FlatDmg

	ptarget := c.Core.Combat.PrimaryTarget()

	c.Core.QueueAttack(
		ai,
		combat.NewSingleTargetHit(ptarget.Key()),
		hitmark,
		hitmark,
		c.particleCB,
		c.removemarkCB,
		c.a1cb,
	)

	enemyhit := 0
	for _, e := range enemies {
		if enemyhit < 5 {
			if e.StatusIsActive(skillMarkKey) {
				if e.Key() != ptarget.Key() {
					c.Core.QueueAttack(
						aimissiles,
						combat.NewSingleTargetHit(e.Key()),
						hitmark+60,
						hitmark+60,
						c.removemarkCB,
					)
					enemyhit++
				}
			}
		}
	}
	c.WaveMomentum = 0
	return action.Info{
		Frames:          func(next action.Action) int { return attackFramesE },
		AnimationLength: attackFramesE + SSBframe,
		CanQueueAfter:   attackFramesE + SSBframe,
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) removemarkCB(a combat.AttackCB) {
	e := a.Target.(*enemy.Enemy)
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !e.StatusIsActive(skillMarkKey) {
		return
	}
	e.DeleteStatus(skillMarkKey)
}
