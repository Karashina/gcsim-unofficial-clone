package flins

import (
	_ "embed"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

// Lunar Charged damage ("DMG considered as Lunar-Charged DMG") handler for Flins
func (c *char) onSpecialLunarChargedFlins(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	switch ae.Info.Abil {

	// Ancient Ritual: Cometh the Night: Middle Phase Lunar-Charged DMG
	case "Flins QMid Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Middle Phase Lunar-Charged DMG (Q)",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * burstlcmid[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
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

		// Ancient Ritual: Cometh the Night: Final Phase Lunar-Charged DMG
	case "Flins QFin Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Final Phase Lunar-Charged DMG (Q)",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * burstlcfin[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
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

	// Ancient Ritual: Cometh the Night: Thunderous Symphony DMG
	case "Flins TS Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Thunderous Symphony DMG (Q)",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * burstlcts[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
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

	// Ancient Ritual: Cometh the Night: Thunderous Symphony Additional DMG
	case "Flins TSADD Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Thunderous Symphony Additional DMG (C6)",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * burstlctsadd[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
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

	// Ancient Ritual: Cometh the Night: Thunderous Symphony Additional DMG
	case "Flins C2 Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "The Devil's Wall (C2)",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * 0.5 * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3
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

// Register Flins's special Lunar Charged callback
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onSpecialLunarChargedFlins, "lc-flins-special")
}
