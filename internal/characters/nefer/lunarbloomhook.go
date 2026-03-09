package nefer

import (
	_ "embed"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// onLunarBloomNeferSpecialはPPの影ダミーと6凸ダミーをLunar-Bloom攻撃に解決する。
func (c *char) onLunarBloomNeferSpecial(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	switch ae.Info.Abil {
	case "Nefer PP1Shade Dummy (C)":
		// 幻影公演 1-Hitダメージ（影）
		c.queueLunarBloomAttack(n, skillpps1[c.TalentLvlSkill()], "Phantasm Performance 1-Hit (Shades / C)", 0)
		return false

	case "Nefer PP2Shade Dummy (C)":
		// 幻影公演 2-Hitダメージ（影）
		c.queueLunarBloomAttack(n, skillpps2[c.TalentLvlSkill()], "Phantasm Performance 2-Hit (Shades / C)", 0)
		return false

	case "Nefer PP3Shade Dummy (C)":
		// 幻影公演 3-Hitダメージ（影）
		c.queueLunarBloomAttack(n, skillpps3[c.TalentLvlSkill()], "Phantasm Performance 3-Hit (Shades / C)", 0)
		return false

	case "Nefer C6 2nd Dummy (C)":
		// 6凸：第2ヒットをEMスケーリングのLunar-Bloomダメージに変換
		mult := 0.85
		c.queueLunarBloomAttack(n, mult, "Phantasm Performance C6 2nd Hit (C/C6)", 0)
		return false

	case "Nefer C6 Extra Dummy (C)":
		// 6凸：幻影公演後の追加インスタンス
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
	if c.Base.Cons >= 1 { // 1凸
		c1mult = 0.6
	}
	baseDmg := em * (mult + c1mult) * (1 + c.LBBaseReactBonus(atk))
	emBonus := (6 * em) / (2000 + em)
	lbReactBonus := c.LBReactBonus(atk)
	c.Core.Log.NewEvent("[Nefer Debug] LBReactBonus", 3, c.Index).
		Write("ability", abilName).
		Write("lb_react_bonus", lbReactBonus).
		Write("em", em).
		Write("base_dmg", baseDmg).
		Write("em_bonus", emBonus)
	atk.FlatDmg = baseDmg * (1 + emBonus + lbReactBonus) * (1 + c.ElevationBonus(atk))
	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)
	trg := combat.NewCircleHitOnTarget(target.Pos(), nil, 5)
	c.Core.QueueAttackWithSnap(atk, snap, trg, delay, c.makePhantasmBonus())
}

// Nefer専用のLunar Bloomコールバックを登録
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onLunarBloomNeferSpecial, "lb-nefer-special")
}
