package hutao

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1BuffKey = "hutao-a1"
)

// 彼岸よって招かれし Paramita Papilio 状態が終了した時、
// パーティーの全味方（胡桃自身を除く）の会心率12%が8秒間アップする。
func (c *char) a1() {
	if c.Base.Ascension < 1 || !c.applyA1 {
		return
	}
	c.applyA1 = false

	for i, char := range c.Core.Player.Chars() {
		// 胡桃には適用されない
		if c.Index == i {
			continue
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a1BuffKey, 480),
			AffectedStat: attributes.CR,
			Amount: func() ([]float64, bool) {
				return c.a1buff, true
			},
		})
	}
}

// 胡桃のHPが50%以下の時、炎元素ダメージボーナスが33%アップする。
//
// - TODO: ゲーム内では実際には0.3秒ごとのチェック。HPが50%未満なら次のチェックまでバフ有効
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4buff = make([]float64, attributes.EndStatType)
	c.a4buff[attributes.PyroP] = 0.33
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("hutao-a4", -1),
		AffectedStat: attributes.PyroP,
		Amount: func() ([]float64, bool) {
			if c.CurrentHPRatio() <= 0.5 {
				return c.a4buff, true
			}
			return nil, false
		},
	})
}
