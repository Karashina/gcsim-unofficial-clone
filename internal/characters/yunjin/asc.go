package yunjin

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// 固有天賦1は未実装:
// TODO: 雲菫が攻撃された正確な瞬間に開花を使用すると、レベル2チャージ（長押し）形式を発動する。

// 元素タイプを計算
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	partyElementalTypes := make(map[attributes.Element]int)
	for _, char := range c.Core.Player.Chars() {
		partyElementalTypes[char.Base.Element]++
	}
	for range partyElementalTypes {
		c.partyElementalTypes += 1
	}
	c.Core.Log.NewEvent("Yun Jin Party Elemental Types (A4)", glog.LogCharacterEvent, c.Index).
		Write("party_elements", c.partyElementalTypes)
}

// 飛雲旗陣の通常攻撃ダメージバフが、パーティに含まれる元素タイプ
// 1/2/3/4種類ごとに雲菫の防御力の2.5%/5%/7.5%/11.5%さらに増加する。
func (c *char) a4() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	if c.partyElementalTypes == 4 {
		return 0.115
	}
	return 0.025 * float64(c.partyElementalTypes)
}
