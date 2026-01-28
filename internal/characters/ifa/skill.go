package ifa

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int
var skillCancelFrames []int

const (
	plungeAvailableKey = "ifa-plunge-available"
)

func init() {
	skillFrames = frames.InitAbilSlice(19) // E -> E
	skillFrames[action.ActionDash] = 19
	skillFrames[action.ActionSwap] = 589 + 45 // wait for nightsoul to run out and fall onto the ground

	skillCancelFrames = frames.InitAbilSlice(45) // E -> Dash/Jump
	skillCancelFrames[action.ActionSwap] = 45
}

func (c *char) reduceNightsoulPoints(val float64) {
	c.nightsoulState.ConsumePoints(val)
	c.checkNS()
}

func (c *char) checkNS() {
	if c.nightsoulState.Points() < 0.001 {
		c.exitNightsoul()
	}
}

func (c *char) enterNightsoul() {
	c.nightsoulState.EnterBlessing(80)
	c.nightsoulSrc = c.Core.F
	c.Core.Tasks.Add(c.nightsoulPointReduceFunc(c.nightsoulSrc), 6)
	c.skillParticleICD = false
}

func (c *char) nigthsoulFallingMsg() {
	c.Core.Log.NewEvent("nightsoul ended, falling", glog.LogCharacterEvent, c.Index)
}
func (c *char) exitNightsoul() {
	if !c.nightsoulState.HasBlessing() {
		return
	}

	switch c.Core.Player.CurrentState() {
	case action.Idle:
		c.Core.Player.SwapCD = 37
		c.nigthsoulFallingMsg()
	case action.DashState, action.NormalAttackState:
		c.nigthsoulFallingMsg()
	}

	c.nightsoulState.ExitBlessing()
	c.nightsoulState.ClearPoints()
	c.nightsoulSrc = -1
	c.SetCD(action.ActionSkill, 7.5*60)
	c.AddStatus(plungeAvailableKey, 26, true)
}

func (c *char) nightsoulPointReduceFunc(src int) func() {
	return func() {
		if c.nightsoulSrc != src {
			return
		}
		c.reduceNightsoulPoints(0.8)
		// reduce 0.8 point per 6, which is 8 per second
		c.Core.Tasks.Add(c.nightsoulPointReduceFunc(src), 6)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.exitNightsoul()
		return action.Info{
			Frames:          frames.NewAbilFunc(skillCancelFrames),
			AnimationLength: skillCancelFrames[action.InvalidAction],
			CanQueueAfter:   skillCancelFrames[action.ActionLowPlunge], // earliest cancel
			State:           action.SkillState,
		}, nil
	}
	c.enterNightsoul()
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  -1,
		Message: "Tonicshot Heal",
		Src:     c.Stat(attributes.EM)*skillHealPct[c.TalentLvlSkill()] + skillHealCst[c.TalentLvlSkill()],
		Bonus:   c.Stat(attributes.Heal),
	})
	c.c1()
	if c.skillParticleICD {
		return
	}
	c.skillParticleICD = true
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Anemo, c.ParticleDelay)
}
