package illuga

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1ICD    = "illuga-c1-icd"
	c1ICDDur = 15 * 60 // 15s
	c4Key    = "illuga-c4-def"
	c4Def    = 200.0
)

// C1: Restores 12 Energy when triggering a Geo-related Elemental Reaction
// Can only trigger once every 15s. Only triggers when Illuga is on-field.
func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}

	// Hook into Crystallize reactions
	cb := func(args ...interface{}) bool {
		// I-12 fix: Only trigger when Illuga is on-field
		if c.Core.Player.Active() != c.Index {
			return false
		}
		if c.StatusIsActive(c1ICD) {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}

		c.AddStatus(c1ICD, c1ICDDur, true)
		c.AddEnergy("illuga-c1", 12)

		c.Core.Log.NewEvent("Illuga C1: Energy restored from Geo reaction", glog.LogCharacterEvent, c.Index).
			Write("energy_gained", 12)

		return false
	}

	// Subscribe to all Crystallize events
	c.Core.Events.Subscribe(event.OnCrystallizePyro, cb, "illuga-c1-pyro")
	c.Core.Events.Subscribe(event.OnCrystallizeHydro, cb, "illuga-c1-hydro")
	c.Core.Events.Subscribe(event.OnCrystallizeCryo, cb, "illuga-c1-cryo")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, cb, "illuga-c1-electro")
	c.Core.Events.Subscribe(event.OnLunarCrystallize, cb, "illuga-c1-lcrs")
}

// C2: During Oriole-Song, for every 7 Nightingale's Song stacks consumed,
// perform a lamp attack dealing Geo DMG based on 400% EM + 200% DEF
func (c *char) c2LampAttack() {
	if c.Base.Cons < 2 {
		return
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "C2 Lamp Attack",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
	}

	// Calculate FlatDmg: 400% EM + 200% DEF
	em := c.Stat(attributes.EM)
	def := c.TotalDef(false)
	ai.FlatDmg = em*4.0 + def*2.0

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4)
	c.Core.QueueAttack(ai, ap, 0, 10)

	c.Core.Log.NewEvent("Illuga C2: Lamp attack triggered", glog.LogCharacterEvent, c.Index).
		Write("em", em).
		Write("def", def).
		Write("flat_dmg", ai.FlatDmg)
}

// C4: During Oriole-Song, all party members gain DEF +200 (I-7 fix: was self-only)
func (c *char) c4Init() {
	if c.Base.Cons < 4 {
		return
	}

	// Apply DEF bonus to ALL party members when Oriole-Song is active
	for _, char := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DEF] = c4Def

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(c4Key, -1),
			AffectedStat: attributes.DEF,
			Amount: func() ([]float64, bool) {
				if !c.orioleSongActive {
					return nil, false
				}
				return m, true
			},
		})
	}
}

// C6: Lightkeeper's Oath bonuses are doubled when Moonsign is Ascendant Gleam
// This is implemented in asc.go applyLightkeeperOath()

func (c *char) consInit() {
	c.c1Init()
	c.c4Init()
}
