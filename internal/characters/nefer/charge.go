package nefer

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var (
	chargeFrames     []int
	phantasmFrames   []int
	phantasmHitmarks = []int{28, 81, 48, 56, 81} // 2 Nefer hits + 3 Shade hits
)

const (
	chargeHitmark = 45
)

func init() {
	chargeFrames = frames.InitAbilSlice(60)
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark

	phantasmFrames = frames.InitAbilSlice(92)
	phantasmFrames[action.ActionDash] = 28
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// Check if we should use Phantasm Performance
	if c.StatusIsActive(skillKey) && c.Core.Player.Verdant.Count() >= 1 {
		return c.phantasmPerformance(p)
	}

	// Normal Charged attack (Slither): forward movement with stamina drain; exit deals Dendro DMG. Shadow Dance lowers max consumption.

	if p["hold"] > 0 {
		// Slither hold - set up stamina drain mechanics
		dur := p["hold"]
		chargeFrames[action.ActionSkill] = 1 // using the Elemental Skill while Nefer is in the Slither state will not cause her to exit the state.
		prevbonus := c.Core.Player.Verdant.GetGainBonus()
		frameremaining := c.Core.Player.Verdant.RemainingFrames()

		// A4 interaction: strengthen Verdant Dew gain in Shadow Dance
		if c.StatusIsActive(skillKey) {
			emBonus := 1 + min(0.001*float64(max(0, c.Stat(attributes.EM)-500)), 0.5)
			c.Core.Player.Verdant.SetGainBonus(prevbonus + emBonus)
			if frameremaining < dur {
				c.Core.Player.Verdant.StartCharge(dur - frameremaining)
			}
			c.Core.Tasks.Add(func() {
				c.Core.Player.Verdant.SetGainBonus(prevbonus)
			}, dur)
		}
	}

	// Deal Charged Attack DMG
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3.5),
		chargeHitmark,
		chargeHitmark,
	)

	// Attempt to absorb Seeds of Deceit within range when using Charged Attack
	c.absorbSeeds(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) phantasmPerformance(_ map[string]int) (action.Info, error) {
	/*
		When nefer is on the Shadow Dance state and party have at least 1 Verdant Dew, Nefer's Charged Attacks will be replaced with the special Charged Attack "Phantasm Performance", which will not consume Stamina.
		Nefer will deal "Phantasm Performance n-Hit DMG (Nefer)" 2 times and "Phantasm Performance n-Hit DMG (Shades)" 3 times. DMG dealt by the shades is considered Lunar-Bloom DMG.
		after Phantasm Performance 1-Hit DMG (Shades) is used, 1 Verdant Dew will be consumed.
	*/

	// Phantasm Performance: 2 Nefer hits (ATK-scaled) and 3 Shade hits (Lunar-Bloom). Consumes 1 Verdant Dew after first Shade hit.

	// Absorb Seeds of Deceit when unleashing Phantasm Performance
	c.absorbSeeds(6)
	c.QueueCharTask(func() {
		aiATK := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Phantasm Performance 1-Hit (Nefer / C)",
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagExtraAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			FlatDmg:    skillppn1atk[c.TalentLvlSkill()]*c.Stat(attributes.ATK) + skillppn1em[c.TalentLvlSkill()]*c.Stat(attributes.EM),
		}
		c.Core.QueueAttack(
			aiATK,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 4),
			0, 0,
			c.makePhantasmBonus(),
		)

	}, phantasmHitmarks[0])

	// 2nd Nefer hit (ATK) or C6 conversion to Lunar-Bloom
	c.QueueCharTask(func() {
		if c.Base.Cons >= 6 {
			// C6: Convert 2nd hit to Lunar-Bloom DMG based on EM
			ai := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Nefer C6 2nd Dummy (C)",
				FlatDmg:    0,
			}
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
				0, 0,
			)
		} else {
			// Normal 2nd hit
			aiATK := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Phantasm Performance 2-Hit (Nefer / C)",
				AttackTag:  attacks.AttackTagExtra,
				ICDTag:     attacks.ICDTagExtraAttack,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeDefault,
				Element:    attributes.Dendro,
				Durability: 25,
				FlatDmg:    skillppn2atk[c.TalentLvlSkill()]*c.Stat(attributes.ATK) + skillppn2em[c.TalentLvlSkill()]*c.Stat(attributes.EM),
			}
			c.Core.QueueAttack(
				aiATK,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 4),
				0, 0,
				c.makePhantasmBonus(),
			)
		}
	}, phantasmHitmarks[1])

	// Shade hits (dummies) -> resolved to Lunar-Bloom via hook; consume Verdant Dew after first Shade hit
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Nefer PP1Shade Dummy (C)",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
			0, 0,
		)
		// Consume 1 Verdant Dew
		c.Core.Player.Verdant.Consume(1)
	}, phantasmHitmarks[2])

	// Shade 2 (dummy)
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Nefer PP2Shade Dummy (C)",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
			0, 0,
		)
	}, phantasmHitmarks[3])

	// Shade 3 (dummy)
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Nefer PP3Shade Dummy (C)",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
			0, 0,
		)
	}, phantasmHitmarks[4])

	// C6: Extra dummy hit after PP
	if c.Base.Cons >= 6 {
		c.QueueCharTask(func() {
			ai := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Nefer C6 Extra Dummy (C)",
				FlatDmg:    0,
			}
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
				0, 0,
			)
		}, phantasmHitmarks[4]+5)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(phantasmFrames),
		AnimationLength: phantasmFrames[action.InvalidAction],
		CanQueueAfter:   phantasmHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}

// Apply Veil of Falsehood bonus to Phantasm Performance
func (c *char) makePhantasmBonus() combat.AttackCBFunc {
	bonus := c.a1count * 0.08 // Each stack increases DMG by 8%
	if bonus == 0 {
		return nil
	}
	return func(a combat.AttackCB) {
		// Apply bonus as percentage increase to total damage
		a.AttackEvent.Info.FlatDmg *= (1 + bonus)
	}
}

// absorbSeeds looks for Seeds of Deceit (DendroCore with IsSeed=true) within radius of player's position and absorbs them
func (c *char) absorbSeeds(radius float64) {
	absorbed := 0
	player := c.Core.Combat.Player()
	for _, g := range c.Core.Combat.Gadgets() {
		if g == nil {
			continue
		}
		if g.GadgetTyp() != combat.GadgetTypDendroCore {
			continue
		}
		if dc, ok := g.(*reactable.DendroCore); ok {
			if !dc.IsSeed {
				continue
			}
			// distance check
			if dc.Pos().Distance(player.Pos()) <= radius {
				absorbed++
				// remove gadget
				dc.Gadget.Kill()
			}
		}
	}
	if absorbed == 0 {
		return
	}

	prev := c.a1count
	maxStacks := 3.0
	if c.Base.Cons >= 2 {
		maxStacks = 5.0
	}
	c.a1count = math.Min(maxStacks, c.a1count+float64(absorbed))

	// add per-stack duration (9s) and refresh durations
	c.AddStatus("veil-of-falsehood", 9*60, true)

	// If reaching 3 stacks or refreshing 3rd stack's duration, grant EM bonus
	if c.a1count >= 3 || (prev >= 3 && absorbed > 0) {
		emBonus := 100.0
		if c.Base.Cons >= 2 && c.a1count >= 5 {
			emBonus = 200.0
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("veil-em-bonus", 8*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				m := make([]float64, attributes.EndStatType)
				m[attributes.EM] = emBonus
				return m, true
			},
		})
	}
}

