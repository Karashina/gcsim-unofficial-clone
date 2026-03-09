package lynette

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) a1Setup() {
	if c.Base.Ascension < 1 {
		return
	}

	// 元素タイプを計算
	partyEleTypes := make(map[attributes.Element]bool)
	for _, char := range c.Core.Player.Chars() {
		partyEleTypes[char.Base.Element] = true
	}
	count := len(partyEleTypes)

	// 攻撃力%バフのセットアップ
	c.a1Buff = make([]float64, attributes.EndStatType)
	c.a1Buff[attributes.ATKP] = 0.08 + float64(count-1)*0.04
}

// 「魔術・アストニッシングシフト」使用後10秒間、
// パーティ内の元素タイプが1/2/3/4種類の時、
// 全パーティメンバーの攻撃力がそれぞれ8%/12%/16%/20%増加する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	for _, this := range c.Core.Player.Chars() {
		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("lynette-a1", 10*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return c.a1Buff, true
			},
		})
	}
}

// 「魔術・アストニッシングシフト」で召喚された「ビックリボンボックス」が元素変換を行った後、
// リネットの元素爆発のダメージが15%増加する。
// この効果は「ビックリボンボックス」の持続時間終了まで続く。
func (c *char) a4(duration int) {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("lynette-a4", duration),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return m, true
		},
	})
}
