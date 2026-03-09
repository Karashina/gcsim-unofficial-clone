package kuki

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 4凸:
// 草の輪の影響下のキャラクターの通常攻撃・重撃・落下攻撃が敵に命中すると、
//
//	敵の位置に雷草の標が落ち、忍の最大HPの9.7%に基づく雷元素範囲ダメージを与える。
//
// 5秒に1回のみ発動。
func (c *char) c4() {
	//TODO: ダメージが即時発生かどうか不明
	const c4IcdKey = "kuki-c4-icd"
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		trg := args[0].(combat.Target)
		// 4凸のICD中なら無視
		if c.StatusIsActive(c4IcdKey) {
			return false
		}
		// 通常攻撃・重撃・落下攻撃のみ発動
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra && ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		// 攻撃をトリガーしたキャラクターがまだフィールドにいることを確認
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if c.Core.Status.Duration(ringKey) == 0 {
			return false
		}
		c.AddStatus(c4IcdKey, 300, true) // 5s * 60

		//TODO: ダメージのフレームとICDタグ要確認
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Thundergrass Mark",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       0,
			FlatDmg:    c.MaxHP() * 0.097,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 2), 5, 5, c.particleCB)

		return false
	}, "kuki-c4")
}

// 6凸:
// 忍が致命的なダメージを受けても戦闘不能にならない。
// HPが1になったとき自動的に発動。60秒に1回のみ。
// 忍のHPが25%以下になると、元素熟知が150、15秒間増加。この効果は60秒に1回のみ発動。
func (c *char) c6() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 150
	const c6IcdKey = "kuki-c6-icd"
	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if di.Amount <= 0 {
			return false
		}
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		// HPが25%以下かチェック
		if c.CurrentHPRatio() > 0.25 {
			return false
		}
		// 死亡している場合はHP1で復活
		if c.CurrentHPRatio() <= 0 {
			c.SetHPByAmount(1)
		}
		c.AddStatus(c6IcdKey, 3600, false) // 60s * 60

		// 元素熟知を150、15秒間増加
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("kuki-c6", 900),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		return false
	}, "kuki-c6")
}
