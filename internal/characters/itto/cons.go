package itto

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 1凸:
// 「荒瀧の王」使用後、荒瀧一斗は怒髪衝天スタックを2つ獲得する。
// 1秒後、一斗は0.5秒ごとに怒髪衝天スタックを1つ獲得する（1.5秒間）。
// TODO: itto-c1-mechanics TCLエントリへのリンクを後で追加
func (c *char) c1() {
	// Q入力後約75fで初期2スタックを獲得
	c.addStrStack("itto-c1-cast", 2)
	// 「1秒後」は初期2スタック獲得から1秒後を指すため、適切にキューに追加
	for i := 60; i <= 120; i += 30 {
		c.QueueCharTask(func() { c.addStrStack("itto-c1-timer", 1) }, i)
	}
}

// 2凸:
// 「荒瀧の王」使用後、
// パーティ内の岩元素キャラクターごとにそのスキルのCDが1.5秒減少し、
// 荒瀧一斗のエネルギーが6回復する。
// この方法でCDを最大4.5秒減少できる。
// この方法で最大18エネルギー回復できる。
func (c *char) c2() {
	c.AddEnergy("itto-c2", float64(c.c2GeoMemberCount)*6)
	c.ReduceActionCooldown(action.ActionBurst, c.c2GeoMemberCount*(1.5*60))
}

// 4凸:
// 「荒瀧の王」による鬼王状態が終了すると、
// 周囲のパーティメンバー全員が10秒間、防御力+20%と攻撃力+20%を獲得する。
func (c *char) c4() {
	if !c.applyC4 {
		return
	}
	c.applyC4 = false

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.2
	m[attributes.DEFP] = 0.2
	for _, x := range c.Core.Player.Chars() {
		x.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("itto-c4", 10*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

// 6凸の前半部分:
// 荒瀧一斗の重撃の会心ダメージが+70%。
func (c *char) c6() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CD] = 0.7
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("itto-c6", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			return m, true
		},
	})
}
