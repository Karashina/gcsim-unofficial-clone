package yanfei

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 2凸フック：
// HPが50%以下の敵に対して煙绯の重撃会心率を20%増加させる。
func (c *char) c2() {
	if c.Core.Combat.DamageMode {
		m := make([]float64, attributes.EndStatType)
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("yanfei-c2", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagExtra {
					return nil, false
				}
				x, ok := t.(*enemy.Enemy)
				if !ok {
					return nil, false
				}
				if x.HP()/x.MaxHP() >= .5 {
					return nil, false
				}
				m[attributes.CR] = 0.20
				return m, true
			},
		})
	}
}

// 4凸シールド生成を処理
// 「丹書契約」使用時：
// 15秒間、煙绯のHP上限の45%を吸収するシールドを生成
// このシールドは炎元素ダメージを250%多く吸収する
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	c.Core.Player.Shields.Add(&shield.Tmpl{
		ActorIndex: c.Index,
		Target:     -1,
		Src:        c.Core.F,
		ShieldType: shield.YanfeiC4,
		Name:       "Yanfei C4",
		HP:         c.MaxHP() * .45,
		Ele:        attributes.Pyro,
		Expires:    c.Core.F + 15*60,
	})
}
