package ineffa

import (
	_ "embed"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

// Special Lunar Charged damage handler for Ineffa
func (c *char) onLunarChargedIneffaSpecial(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	switch ae.Info.Abil {
	case "Ineffa A1 Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Overclocking Circuit (A1)",
			AttackTag:        attacks.AttackTagLCDamage,
			ICDTag:           attacks.ICDTagLCDamage,
			ICDGroup:         attacks.ICDGroupReactionB,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * 0.65 * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		c.Core.QueueAttackWithSnap(
			atk,
			snap,
			combat.NewCircleHitOnTarget(n.Pos(), nil, 6),
			9,
		)
		return false

	case "Ineffa C2 Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Support Cleaning Module (C2)",
			AttackTag:        attacks.AttackTagLCDamage,
			ICDTag:           attacks.ICDTagLCDamage,
			ICDGroup:         attacks.ICDGroupReactionB,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * 3 * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		c.Core.QueueAttackWithSnap(
			atk,
			snap,
			combat.NewCircleHitOnTarget(n.Pos(), nil, 6),
			180,
		)
		return false

	case "Ineffa C6 Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "A Dawning Morn for You (C6)",
			AttackTag:        attacks.AttackTagLCDamage,
			ICDTag:           attacks.ICDTagLCDamage,
			ICDGroup:         attacks.ICDGroupReactionB,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * 1.35 * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		c.Core.QueueAttackWithSnap(
			atk,
			snap,
			combat.NewCircleHitOnTarget(n.Pos(), nil, 6),
			0,
		)
		return false
	}
	return false
}

// Register Ineffa's special Lunar Charged callback
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onLunarChargedIneffaSpecial, "lc-ineffa-special")
}
