package aino

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
	c1Key        = "aino-c1"
	c1Duration   = 15 * 60
	c1EMBonus    = 80
	c6Key        = "aino-c6"
	c6Duration   = 15 * 60
	c6DMGBonus   = 0.15
	c6ExtraBonus = 0.20
)

// C1 - After Aino uses her Elemental Skill or her Elemental Burst,
// her Elemental Mastery will be increased by 80 for 15s.
// Also, he Elemental Mastery of other nearby active party members will be increased by 80 for 15s.
// The Elemental Mastery-increasing effects of this Constellation do not stack.
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.applyC1Buff()
		return false
	}, "aino-c1-skill")

	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.applyC1Buff()
		return false
	}, "aino-c1-burst")
}

func (c *char) applyC1Buff() {
	// Apply to all party members
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c1Key, c1Duration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return c.c1EMBuff(), true
			},
		})
	}
	c.Core.Log.NewEvent("aino c1 em buff applied", glog.LogCharacterEvent, c.Index)
}

func (c *char) c1EMBuff() []float64 {
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.EM] = c1EMBonus
	return buff
}

// C2 - If Aino is off-field while the Focused Hydronic Cooling Zone of her Elemental Burst Precision Hydronic Cooler is active,
// when your active party member hits a nearby opponent with an attack, the Cool Your Jets Ducky will fire an additional water ball at that opponent,
// dealing AoE Hydro DMG.
// It has no Mult but it's FlatDmg is sum of 25% of Aino's ATK and 100% of her Elemental Mastery.
// AttackTag of this DMG is considered as AttackTagElementalBurst. This effect can be triggered once every 5s.
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {

		// Check if Aino is off-field
		if c.Core.Player.Active() == c.Index {
			return false
		}

		// Check if burst is active
		if !c.StatusIsActive(burstKey) {
			return false
		}

		// Check if on ICD
		if c.StatusIsActive(c.c2IcdKey) {
			return false
		}

		// Trigger additional water ball
		atk := c.Stat(attributes.ATK)
		em := c.Stat(attributes.EM)
		flatDmg := 0.25*atk + 1.0*em

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Precision Hydronic Cooler (C2)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
			Mult:       0,
			FlatDmg:    flatDmg,
		}

		// args[0] is the target (enemy) and args[1] is the attack event
		tgt := args[0].(combat.Target)
		// Avoid generating a snapshot synchronously inside an OnEnemyHit event handler.
		// Synchronous snapshots here can re-enter event handlers and cause unbounded recursion
		// (see issue: stack overflow during sim runs). Schedule the snapshot+damage 1 frame
		// later to break the re-entrant call chain.
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(tgt, nil, 3), 1, 1)

		c.AddStatus(c.c2IcdKey, 5*60, false)
		c.Core.Log.NewEvent("aino c2 proc", glog.LogCharacterEvent, c.Index)

		return false
	}, "aino-c2")
}

// C4 - When the Elemental Skill hits an opponent, it will restore 10 Elemental Energy for Aino.
// Energy can be restored to her in this manner once every 10s.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != c.Index {
			return false
		}

		if ae.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}

		if c.StatusIsActive("aino-c4-icd") {
			return false
		}

		c.AddEnergy("aino-c4", 10)
		c.AddStatus("aino-c4-icd", 10*60, false)
		c.Core.Log.NewEvent("aino c4 proc", glog.LogCharacterEvent, c.Index)

		return false
	}, "aino-c4")
}

// C6 - For the next 15s after using the Elemental Burst, DMG from nearby active characters' Electro-Charged, Bloom,
// Lunar-Charged, and Lunar-Bloom reactions is increased by 15%.
// When the Moonsign is Ascendant, DMG from the aforementioned reactions will be further increased by 20%.
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}

		c.AddStatus(c6Key, c6Duration, false)
		c.Core.Log.NewEvent("aino c6 buff applied", glog.LogCharacterEvent, c.Index)

		return false
	}, "aino-c6")

	// Apply reaction damage bonus for Electro-Charged
	c.Core.Events.Subscribe(event.OnElectroCharged, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(c6Key) {
			return false
		}

		bonus := c6DMGBonus
		if c.MoonsignAscendant {
			bonus += c6ExtraBonus
		}
		atk.Info.FlatDmg += atk.Info.FlatDmg * bonus
		c.Core.Log.NewEvent("aino c6 electro-charged dmg bonus", glog.LogCharacterEvent, c.Index).
			Write("bonus", bonus)

		return false
	}, "aino-c6-ec")

	// Apply reaction damage bonus for Bloom
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(c6Key) {
			return false
		}

		bonus := c6DMGBonus
		if c.MoonsignAscendant {
			bonus += c6ExtraBonus
		}
		atk.Info.FlatDmg += atk.Info.FlatDmg * bonus
		c.Core.Log.NewEvent("aino c6 bloom dmg bonus", glog.LogCharacterEvent, c.Index).
			Write("bonus", bonus)

		return false
	}, "aino-c6-bloom")
}
