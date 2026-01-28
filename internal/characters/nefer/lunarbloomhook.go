package nefer

import (
	_ "embed"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// onLunarBloomNeferSpecial resolves PP shade dummies and C6 dummies into Lunar-Bloom attacks.
func (c *char) onLunarBloomNeferSpecial(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	switch ae.Info.Abil {
	case "Nefer PP1Shade Dummy (C)":
		// Phantasm Performance 1-Hit DMG (Shades)
		c.queueLunarBloomAttack(n, skillpps1[c.TalentLvlSkill()], "Phantasm Performance 1-Hit (Shades / C)", 0)
		return false

	case "Nefer PP2Shade Dummy (C)":
		// Phantasm Performance 2-Hit DMG (Shades)
		c.queueLunarBloomAttack(n, skillpps2[c.TalentLvlSkill()], "Phantasm Performance 2-Hit (Shades / C)", 0)
		return false

	case "Nefer PP3Shade Dummy (C)":
		// Phantasm Performance 3-Hit DMG (Shades)
		c.queueLunarBloomAttack(n, skillpps3[c.TalentLvlSkill()], "Phantasm Performance 3-Hit (Shades / C)", 0)
		return false

	case "Nefer C6 2nd Dummy (C)":
		// C6: second hit converted to EM scaling Lunar-Bloom DMG
		mult := 0.85
		c.queueLunarBloomAttack(n, mult, "Phantasm Performance C6 2nd Hit (C/C6)", 0)
		return false

	case "Nefer C6 Extra Dummy (C)":
		// C6: extra instance after Phantasm Performance
		mult := 1.2
		c.queueLunarBloomAttack(n, mult, "Phantasm Performance C6 Extra (C/C6)", 0)
		return false
	}
	return false
}

func (c *char) queueLunarBloomAttack(target combat.Target, mult float64, abilName string, delay int) {
	atk := combat.AttackInfo{ActorIndex: c.Index, Abil: abilName, AttackTag: attacks.AttackTagLBDamage, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro, IgnoreDefPercent: 1}
	em := c.Stat(attributes.EM)
	c1mult := 0.0
	if c.Base.Cons >= 1 { // C1
		c1mult = 0.6
	}
	baseDmg := em * (mult + c1mult) * (1 + c.LBBaseReactBonus(atk))
	emBonus := (6 * em) / (2000 + em)
	atk.FlatDmg = baseDmg * (1 + emBonus + c.LBReactBonus(atk)) * (1 + c.ElevationBonus(atk))
	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)
	trg := combat.NewCircleHitOnTarget(target.Pos(), nil, 5)
	c.Core.QueueAttackWithSnap(atk, snap, trg, delay, c.makePhantasmBonus())
}

// Register Nefer's special Lunar Bloom callback
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onLunarBloomNeferSpecial, "lb-nefer-special")
}
