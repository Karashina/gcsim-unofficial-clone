package yaemiko

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 殺生櫻の雷が敵にヒットすると、近くの全パーティメンバーの雷元素ダメージバフが5秒間20%増加する。
func (c *char) c4() {
	// TODO: これは八重神子もトリガーされる？されると仮定
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("yaemiko-c4", 5*60),
			AffectedStat: attributes.ElectroP,
			Amount: func() ([]float64, bool) {
				return c.c4buff, true
			},
		})
	}
}
