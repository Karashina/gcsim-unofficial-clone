package emilie

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

var (
	skillFrames []int
)

const (
	skillLCSpawn      = 32
	skillLCHitmark    = 40
	skillarkheHitmark = 64
	skillLCFirstTick  = 109
	LCTickInterval    = 89
	particleICDKey    = "emilie-particle-icd"
	scentsICDKey      = "scents-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(32)
	skillFrames[action.ActionDash] = 32
	skillFrames[action.ActionJump] = 32
	skillFrames[action.ActionSwap] = 32
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	c.RemoveLumidouceCase(c.LCSource)
	c.LCLevel = 0

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Fragrance Extraction (E)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	radius := 5.0
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: 1}, radius),
		skillLCSpawn,
		skillLCHitmark,
	)

	c.SetCD(action.ActionSkill, 25*60+15)
	if c.Base.Cons >= 6 {
		c.c6init()
	}

	if !c.StatusIsActive(burstKey) {
		c.LCActive = true
		c.queueLumidouceCase("Skill", skillLCSpawn, skillLCFirstTick)
	}

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
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 2.5*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Dendro, c.ParticleDelay)
}

func (c *char) queueLumidouceCase(src string, LCSpawn, firstTick int) {
	// calculate duration
	dur := 22 * 60
	spawnFn := func() {
		// setup variables for tracking
		c.LCActive = true
		c.LCSource = c.Core.F
		c.LCTickSrc = c.Core.F
		c.LCActiveUntil = c.Core.F + dur
		// queue up removal at the end of the duration for gcsl conditional
		c.Core.Tasks.Add(c.RemoveLumidouceCase(c.Core.F), dur)
		c.L1ai = combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       fmt.Sprintf("Level 1 Lumidouce Case Attack DMG (%v) (E)", src),
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupEmilieSkill,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillTickL1[c.TalentLvlSkill()],
		}
		c.L2ai = combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       fmt.Sprintf("Level 2 Lumidouce Case Attack DMG (%v) (E)", src),
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupEmilieSkill,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillTickL2[c.TalentLvlSkill()],
		}
		player := c.Core.Combat.Player()
		c.LCPos = geometry.CalcOffsetPoint(player.Pos(), geometry.Point{Y: 1.5}, player.Direction())

		c.Core.Tasks.Add(c.LumidouceCaseTick(c.Core.F), firstTick)
		c.Core.Log.NewEvent("Lumidouce Case activated", glog.LogCharacterEvent, c.Index).
			Write("source", src).
			Write("expected end", c.LCActiveUntil).
			Write("next expected tick", c.Core.F+89)
	}
	if LCSpawn > 0 {
		c.Core.Tasks.Add(spawnFn, LCSpawn)
		return
	}
	spawnFn()
}

func (c *char) LumidouceCaseTick(src int) func() {
	return func() {
		if src != c.LCTickSrc {
			return
		}
		c.Core.Log.NewEvent("Lumidouce Case ticked", glog.LogCharacterEvent, c.Index).
			Write("next expected tick", c.Core.F+89).
			Write("active", c.LCActiveUntil).
			Write("src", src)
		// trigger damage
		switch c.LCLevel {
		case 0:
			c.Core.QueueAttack(
				c.L1ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 0.5),
				0,
				c.LCTravel,
				c.particleCB,
			)
		case 1:
			c.Core.QueueAttack(
				c.L2ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 0.5),
				0,
				c.LCTravel,
				c.particleCB,
			)
			c.Core.QueueAttack(
				c.L2ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 0.5),
				10,
				c.LCTravel,
				c.particleCB,
			)
		default:
			c.Core.Log.NewEvent("ERROR: Invalid LC Level", glog.LogCharacterEvent, c.Index)
		}

		// queue up next hit only if next hit LC is still active
		if c.Core.F+LCTickInterval <= c.LCActiveUntil {
			c.Core.Tasks.Add(c.LumidouceCaseTick(src), LCTickInterval)
		}
	}
}

func (c *char) RemoveLumidouceCase(src int) func() {
	return func() {
		if c.LCSource != src {
			c.Core.Log.NewEvent("Lumidouce Case not removed, src changed", glog.LogCharacterEvent, c.Index).
				Write("src", src)
			return
		}
		c.Core.Log.NewEvent("Lumidouce Case removed", glog.LogCharacterEvent, c.Index).
			Write("src", src)
		c.LCActive = false
	}
}

func (c *char) ScentsCheck() {
	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {

		if c.StatusIsActive(scentsICDKey) {
			return false
		}

		enemies := c.Core.Combat.EnemiesWithinArea(
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10),
			nil,
		)

		for _, x := range enemies {
			t, ok := x.(*enemy.Enemy)
			if !ok {
				continue
			}
			if t.IsBurning() {
				if !c.LCActive {
					c.Scents = 0
					c.LCLevel = 0
				}
				if c.Scents < 2 {
					c.QueueCharTask(c.AddScents, 90)
					c.AddStatus(scentsICDKey, 120, false)
				}
				if c.Scents >= 2 && c.LCLevel != 1 {
					c.LCLevel = 1
					c.Scents = 0
				}
				if c.Scents >= 2 {
					c.a1()
					c.Scents = 0
				}
			}
		}
		return false
	}, "emilie-scents-check")
}

func (c *char) AddScents() {
	c.Scents++
}
