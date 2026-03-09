package keqing

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 雷楔が存在する間に星辰帰位を再発動すると、刻晴の武器に5秒間雷元素付与を獲得する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	// スキル再発動時に開始
	// 実際は5.6秒
	// EE N1 Q N1 E 2N1C / 5N1C N1 をぎりぎりカバーする程度
	// 非常に途切れやすい
	dur := 5*60 + 20
	c.AddStatus("keqinginfuse", dur, true)
	c.Core.Player.AddWeaponInfuse(
		c.Index,
		"keqing-a1",
		attributes.Electro,
		dur,
		true,
		attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
	)
	c.Core.Events.Emit(event.OnInfusion, c.Index, attributes.Electro, dur)

}

// 天街巡遊を発動すると、刻晴の会心率が15%、元素チャージ効率が15%上昇する。この効果は8秒間持続する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("keqing-a4", 480),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return c.a4buff, true
		},
	})
}
