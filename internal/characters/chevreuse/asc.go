package chevreuse

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// パーティ全員が炎元素と雷元素のキャラクターで、かつ炎元素と雷元素が
// それぞれ少なくとも1人ずついる場合:
// シュヴルーズは付近のパーティメンバーに「共同戦術」を付与する:
// キャラクターが過負荷反応を起こした後、この過負荷反応の影響を受けた
// 敵の炎耐性と雷耐性が40%減少する（6秒間）。
// パーティ内のキャラクターの元素タイプが固有天賦の基本条件を
// 満たさなくなると「共同戦術」効果は解除される。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	// 炎と雷のみかチェック
	chars := c.Core.Player.Chars()
	count := make(map[attributes.Element]int)
	for _, this := range chars {
		count[this.Base.Element]++
	}
	c.onlyPyroElectro = count[attributes.Pyro] > 0 && count[attributes.Electro] > 0 && count[attributes.Electro]+count[attributes.Pyro] == len(chars)

	if !c.onlyPyroElectro {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		// 過負荷ダメージでなければトリガーしない
		if atk.Info.AttackTag != attacks.AttackTagOverloadDamage {
			return false
		}

		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		t.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("chev-a1-pyro", 6*60),
			Ele:   attributes.Pyro,
			Value: -0.40,
		})
		t.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("chev-a1-electro", 6*60),
			Ele:   attributes.Electro,
			Value: -0.40,
		})

		return false
	}, "cheuv-a1")
}

// シュヴルーズが近距離キャノン急射で過充填弾を発射した後、
// 付近のパーティ内の炎元素・雷元素キャラクターは30秒間、
// シュヴルーズのHP上限1,000ごとに攻撃力+1%を獲得する。
// この方法で攻撃力は最大40%まで増加可能。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = min(c.MaxHP()/1000*0.01, 0.4)
	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element != attributes.Pyro && char.Base.Element != attributes.Electro {
			continue
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("chev-a4", 30*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}
