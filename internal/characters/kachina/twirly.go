package kachina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var radius = 4.2

const particleICDKey = "kachina-particle-icd"

func (c *char) newTwirly() {
	if c.StatusIsActive(burstKey) {
		radius = 5.2
	}
	c.twirlyDir = c.Core.Combat.Player().Direction()
	c.twirlyPos = geometry.CalcOffsetPoint(c.Core.Combat.Player().Pos(), geometry.Point{Y: 3}, c.twirlyDir)
	c.Core.QueueAttack(c.skillai, combat.NewCircleHitOnTarget(c.twirlyPos, nil, radius), 0, 0, c.particleCB())

	c.Twirlysrc = c.Core.F
	// create a construct
	con := &TurboTwirly{
		src:    c.Twirlysrc,
		expiry: c.Twirlysrc + 9999,
		c:      c,
		dir:    c.twirlyDir,
		pos:    c.twirlyPos,
	}

	c.Core.Constructs.New(con, true)

	c.Core.Log.NewEvent(
		"Turbo Twirly added",
		glog.LogCharacterEvent,
		c.Index,
	).
		Write("next_tick", c.Core.F+150)

	c.skillai = combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Turbo Twirly Independent DMG (Tick)",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalArt,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       75,
		Element:        attributes.Geo,
		Durability:     25,
		UseDef:         true,
		Mult:           skillIndependent[c.TalentLvlSkill()],
		FlatDmg:        c.a4flat,
	}
	c.Core.Tasks.Add(c.Twirlytick(c.Core.F), 150)
}

func (c *char) Twirlytick(src int) func() {
	return func() {
		c.Core.Log.NewEvent("Twirly checking for tick", glog.LogCharacterEvent, c.Index).
			Write("src", src).
			Write("char", c.Index)
		if !c.Core.Constructs.Has(src) {
			return
		}
		if !c.StatusIsActive(skillKey) {
			return
		}
		if c.StatusIsActive(skillRideKey) {
			return // Don't tick if riding
		}
		c.Core.Log.NewEvent("Twirly ticked", glog.LogCharacterEvent, c.Index).
			Write("next expected", c.Core.F+117).
			Write("src", src).
			Write("char", c.Index)

		if c.StatusIsActive(burstKey) {
			radius = 5.2
		}
		c.Core.QueueAttack(c.skillai, combat.NewCircleHitOnTarget(c.twirlyPos, nil, radius), 0, 0, c.particleCB())
		c.depleteNightsoulPoints("attack")
		c.Core.Tasks.Add(c.Twirlytick(src), 117)
	}
}

func (c *char) TwirlyRideAttack() func() {
	return func() {
		c.Core.Log.NewEvent("Twirly checking for Ride attack", glog.LogCharacterEvent, c.Index).
			Write("char", c.Index)
		if !c.StatusIsActive(skillKey) {
			return
		}
		if !c.StatusIsActive(skillRideKey) {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Turbo Twirly Mounted DMG",
			AttackTag:      attacks.AttackTagElementalArt,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagElementalArt,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeBlunt,
			PoiseDMG:       100,
			Element:        attributes.Geo,
			Durability:     25,
			UseDef:         true,
			Mult:           skillRide[c.TalentLvlSkill()],
			FlatDmg:        c.a4flat,
		}
		if c.StatusIsActive(burstKey) {
			radius = 5.2
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.twirlyPos, nil, radius), 0, 0, c.particleCB())
		c.depleteNightsoulPoints("attack")
	}
}

func (c *char) removeTwirly() {
	c.Core.Constructs.Destroy(c.Twirlysrc)
}

func (c *char) particleCB() combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(particleICDKey) {
			return
		}
		c.AddStatus(particleICDKey, 0.2*60, true)
		if c.Core.Rand.Float64() < 0.667 {
			c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Geo, c.ParticleDelay)
		}
	}
}

