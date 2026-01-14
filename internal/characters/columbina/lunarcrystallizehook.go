package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// Lunar-Crystallize damage ("DMG considered as Lunar-Crystallize DMG") handler for Columbina
func (c *char) onSpecialLunarCrystallizeColumbina(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	switch ae.Info.Abil {

	// Moondew Cleanse Lunar-Crystallize variant (if applicable)
	case "Columbina LCrs Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Moondew Cleanse (Lunar-Crystallize)",
			AttackTag:        attacks.AttackTagLCrsDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Geo,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.MaxHP() * moondewCleanse[c.TalentLvlAttack()] * (1 + c.LCrsBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCrsReactBonus(atk)) * (1 + c.ElevationBonus(atk))
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		trg := combat.NewCircleHitOnTarget(n.Pos(), nil, 6)
		c.Core.QueueAttackWithSnap(
			atk,
			snap,
			trg,
			0,
		)
		c.Core.Events.Emit(event.OnLunarCrystallize, n, ae)
		return false

	// Gravity Interference (Lunar-Crystallize) variant
	case "Columbina GI LCrs Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Gravity Interference (Lunar-Crystallize)",
			AttackTag:        attacks.AttackTagLCrsDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Geo,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.MaxHP() * gravityInterfLCrs[c.TalentLvlSkill()] * (1 + c.LCrsBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCrsReactBonus(atk)) * (1 + c.ElevationBonus(atk))
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		trg := combat.NewCircleHitOnTarget(n.Pos(), nil, 8)
		c.Core.QueueAttackWithSnap(
			atk,
			snap,
			trg,
			0,
		)
		c.Core.Events.Emit(event.OnLunarCrystallize, n, ae)
		return false
	}
	return false
}

// Register Columbina's special Lunar-Crystallize callback
func (c *char) InitLCrsCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onSpecialLunarCrystallizeColumbina, "lcrs-columbina-special")
}
