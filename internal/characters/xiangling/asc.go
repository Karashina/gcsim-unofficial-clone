package xiangling

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// グゥオパァーの炎範囲を20%拡大する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.guobaFlameRange *= 1.2
}

// グゥオパァー出撃の効果終了時、グゥオパァーは消えた場所に唐辛子を残す。唐辛子を拾うと攻撃力が10%上昇する（10秒間）。
func (c *char) a4(a4Delay int) {
	if c.Base.Ascension < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.10
	// グゥオパァー消滅後、ユーザー指定の遅延でアクティブキャラが唐辛子を拾う
	c.Core.Tasks.Add(func() {
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("xiangling-a4", 10*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		c.Core.Log.NewEvent(
			fmt.Sprintf("xiangling a4 chili pepper picked up by %v", active.Base.Key.String()),
			glog.LogCharacterEvent,
			c.Index,
		)
	}, a4Delay)
}
