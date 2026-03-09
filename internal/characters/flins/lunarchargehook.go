package flins

import (
	_ "embed"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// Flinsのルナチャージダメージ（「ルナチャージダメージとみなされる」）ハンドラー
func (c *char) onSpecialLunarChargedFlins(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	switch ae.Info.Abil {

	// Ancient Ritual: Cometh the Night: 中間フェーズ ルナチャージダメージ
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
		atk.FlatDmg = (c.TotalAtk() * burstlcmid[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3 * (1 + c.ElevationBonus(atk))
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
		c.Core.Events.Emit(event.OnLunarCharged, n, ae)
		return false

		// Ancient Ritual: Cometh the Night: 最終フェーズ ルナチャージダメージ
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
		atk.FlatDmg = (c.TotalAtk() * burstlcfin[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3 * (1 + c.ElevationBonus(atk))
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
		c.Core.Events.Emit(event.OnLunarCharged, n, ae)
		return false

	// Ancient Ritual: Cometh the Night: 雷鸣交響ダメージ
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
		atk.FlatDmg = (c.TotalAtk() * burstlcts[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3 * (1 + c.ElevationBonus(atk))
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
		c.Core.Events.Emit(event.OnLunarCharged, n, ae)
		return false

	// Ancient Ritual: Cometh the Night: 雷鳴交響追加ダメージ
	case "Flins TSADD Dummy":
		atk := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Thunderous Symphony Additional DMG (Q)",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		atk.FlatDmg = (c.TotalAtk() * burstlctsadd[c.TalentLvlBurst()] * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3 * (1 + c.ElevationBonus(atk))
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
		c.Core.Events.Emit(event.OnLunarCharged, n, ae)
		return false

	// Ancient Ritual: Cometh the Night: 雷鳴交響追加ダメージ
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
		atk.FlatDmg = (c.TotalAtk() * 0.5 * (1 + c.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(atk)) * 3 * (1 + c.ElevationBonus(atk))
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
		c.Core.Events.Emit(event.OnLunarCharged, n, ae)
		return false
	}
	return false
}

// Flinsの特殊ルナチャージコールバックを登録
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onSpecialLunarChargedFlins, "lc-flins-special")
}
