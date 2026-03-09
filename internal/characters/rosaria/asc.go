package rosaria

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 告解の兄姉で敵の背後から攻撃した時、ロサリアの会心率12%が5秒間アップする。
// TODO: プレイヤー位置が追加された場合、これを変更する必要があるか？
func (c *char) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = 0.12
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("rosaria-a1", 300),
			AffectedStat: attributes.CR,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		c.Core.Log.NewEvent("Rosaria A1 activation", glog.LogCharacterEvent, c.Index).
			Write("ends_on", c.Core.F+300)
	}
}

// 教典のレクイエム発動時、近くの全パーティーメンバー（ロサリア自身を除く）の
// 会心率がロサリアの会心率15%分アップする（10秒間）。
// この方法で得られる会心率ボーナスは15%を超えない。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	critShare := 0.15 * c.NonExtraStat(attributes.CR)
	if critShare > 0.15 {
		critShare = 0.15
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = critShare
	for i, char := range c.Core.Player.Chars() {
		// ロサリアをスキップ
		if i == c.Index {
			continue
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("rosaria-a4", 600),
			AffectedStat: attributes.CR,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	c.Core.Log.NewEvent("Rosaria A4 activation", glog.LogCharacterEvent, c.Index).
		Write("ends_on", c.Core.F+600).
		Write("crit_share", critShare)
}
