package xiao

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a1Key = "xiao-a1"

// 「妞降」の効果中、魉の与える全ダメージが5%増加する。
// 効果が3秒継続するごとにさらに5%増加。
// 最大ダメージボーナスは25%。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(a1Key, 900+burstStart),
		AffectedStat: attributes.DmgP,
		Amount: func() ([]float64, bool) {
			stacks := 1 + (c.Core.F-c.qStarted)/180
			if stacks > 5 {
				stacks = 5
			}
			m[attributes.DmgP] = float64(stacks) * 0.05
			return m, true
		},
	})
}

// 風輪両立を使用すると、次の風輪両立のダメージが15%増加する。
// この効果は7秒間持続し、最大3スタック。新しいスタック取得時に持続時間がリフレッシュされる。
//
// - スキル発動失敗を避けるため、突破レベルチェックは skill.go で実施
func (c *char) a4() {
	// バフが切れていたらスタックをリセット
	if !c.StatModIsActive(a4BuffKey) {
		c.a4stacks = 0
	}
	// テキストには明記されていないが、最大スタック時でも持続時間はリフレッシュされると仮定
	c.a4stacks++
	if c.a4stacks > 3 {
		c.a4stacks = 3
	}
	c.a4buff[attributes.DmgP] = float64(c.a4stacks) * 0.15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag(a4BuffKey, 420),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return c.a4buff, atk.Info.AttackTag == attacks.AttackTagElementalArt
		},
	})
}
