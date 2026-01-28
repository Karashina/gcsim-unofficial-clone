package ifa

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1IcdKey = "ifa-c1-icd"
)

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	if c.StatusIsActive(c1IcdKey) {
		return
	}
	c.AddStatus(c1IcdKey, 8*60, true)
	c.AddEnergy("ifa-c1", 6)
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 100
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag("ifa-c4", 15*60),
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	if c.Core.Rand.Float64() < 0.5 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Tonicshot (Hold)(C6)",
		AttackTag:      attacks.AttackTagNormal,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNormalAttack,
		ICDGroup:       attacks.ICDGroupIfaShots,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           1.2,
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2.0)
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			ap,
			0,
			0,
			c.particleCB,
		)
	}, 15)
}
