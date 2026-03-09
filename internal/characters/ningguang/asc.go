package ningguang

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/construct"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 固有天賦1はningguang.goで実装済み:
// 凝光が星璃を所持している場合、重撃はスタミナを消費しない。

// 璇璣屏を通過したキャラクターは10秒間、岩元素ダメージ+12%を獲得する。
//
// - 璇璣屏がフィールド上にあり、キャラクターがダッシュした場合に発動
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.GeoP] = 0.12
	//TODO: 以前はPostDashで発動していた。正常に動作するか要確認
	c.Core.Events.Subscribe(event.OnDash, func(_ ...interface{}) bool {
		// 璇璣屏の存在を確認
		if c.Core.Constructs.CountByType(construct.GeoConstructNingSkill) <= 0 {
			return false
		}
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("ning-screen", 600),
			AffectedStat: attributes.GeoP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		return false
	}, "ningguang-a4")
}
