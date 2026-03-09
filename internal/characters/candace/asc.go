package candace

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 固有天賦1は未実装:
// TODO: Candaceが聖典・蒼鵞の康護の長押し中に攻撃を受けた場合、そのスキルは即座にチャージが完了する。

const a4Key = "candace-a4"

// 聖儀・英鵛の潮が与える紅冠の祝福の影響を受けたキャラクターは、キャンディスのHP上限1,000ごとに
// 通常攻撃で元素ダメージを与える時、敵に対するダメージが0.5%増加する。
func (c *char) a4(char *character.CharWrapper) {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a4Key, -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if !c.StatusIsActive(burstKey) {
				return nil, false
			}
			if atk.Info.AttackTag != attacks.AttackTagNormal {
				return nil, false
			}
			if atk.Info.Element == attributes.Physical || atk.Info.Element == attributes.NoElement {
				return nil, false
			}
			m[attributes.DmgP] = 0.005 * c.MaxHP() / 1000
			return m, true
		},
	})
}
