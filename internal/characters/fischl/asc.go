package fischl

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a4IcdKey = "fischl-a4-icd"

// Hexereiパッシブ定数
const (
	hexOverloadAtkKey = "fischl-hex-overload-atk"
	hexECEMKey        = "fischl-hex-ec-em"
	hexC6BoostKey     = "fischl-hex-c6-boost"
	hexDuration       = 10 * 60 // 10秒
	hexOverloadAtkPct = 0.225   // 22.5% ATK
	hexECEM           = 90      // +90 EM
	hexC6Multiplier   = 2.0     // 2x boost when C6 hits
)

// 固有天賦1は未実装:
// TODO: フィッシュルがフルチャージ狙い撃ちでオズに命中した時、オズは矢のダメージの152.7%に等しい範囲雷元素ダメージを与える。

// オズがフィールド上にいる時、アクティブキャラクターが雷元素関連反応を起こすと、
// フィッシュルの攻撃力80%の雷元素ダメージを与える。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// 超開花はガジェットから発生するため、ガジェットを無視しない
	//nolint:unparam // 今は無視。イベントリファクタでbool戻り値は不要になるはず
	a4cb := func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		// オズがフィールド上にいない場合は何もしない
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}
		active := c.Core.Player.ActiveChar()
		if active.StatusIsActive(a4IcdKey) {
			return false
		}
		active.AddStatus(a4IcdKey, 0.5*60, true)

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Fischl A4",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupFischl,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       0.8,
		}

		// 固有天賦4はオズのスナップショットを使用
		// TODO: 「元素反応発生位置」から15m以内の最も近い敵をターゲットすべき
		c.Core.QueueAttackWithSnap(
			ai,
			c.ozSnapshot.Snapshot,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 0.5),
			4)
		return false
	}

	a4cbNoGadget := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		return a4cb(args...)
	}

	c.Core.Events.Subscribe(event.OnOverload, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnElectroCharged, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSuperconduct, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnHyperbloom, a4cb, "fischl-a4")
	c.Core.Events.Subscribe(event.OnQuicken, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnAggravate, a4cbNoGadget, "fischl-a4")
}

// Hexerei パッシブ:
// オズがフィールド上にいる時、チームキャラクターは以下のバフを得る:
// - 過負荷トリガー後: フィッシュルとアクティブキャラが攻撃力+22.5%（10秒）
// - 感電または激化トリガー後: フィッシュルとアクティブキャラが元素熟知+90（10秒）
// - 6凸解放時、6凸攻撃命中で上記効果2倍（10秒）
func (c *char) hexPassive() {
	if !c.isHexerei {
		return
	}

	// ATK%バフを適用するヘルパー
	applyAtkBuff := func(multiplier float64) {
		atkPct := hexOverloadAtkPct * multiplier

		// フィッシュルをバフ
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexOverloadAtkKey, hexDuration),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.ATKP] = atkPct
				return stats[:], true
			},
		})

		// アクティブキャラクターをバフ
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexOverloadAtkKey, hexDuration),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.ATKP] = atkPct
				return stats[:], true
			},
		})
	}

	// EMバフを適用するヘルパー
	applyEMBuff := func(multiplier float64) {
		em := hexECEM * multiplier

		// フィッシュルをバフ
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexECEMKey, hexDuration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.EM] = em
				return stats[:], true
			},
		})

		// アクティブキャラクターをバフ
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexECEMKey, hexDuration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.EM] = em
				return stats[:], true
			},
		})
	}

	// 過負荷を購読
	c.Core.Events.Subscribe(event.OnOverload, func(args ...interface{}) bool {
		// オズが有効かチェック
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}

		// 6凸ブースト状態に基づいて倍率を決定
		multiplier := 1.0
		if c.StatusIsActive(hexC6BoostKey) {
			multiplier = hexC6Multiplier
		}

		applyAtkBuff(multiplier)
		return false
	}, "fischl-hex-overload")

	// 感電と激化を購読
	ecCallback := func(args ...interface{}) bool {
		// オズが有効かチェック
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}

		// 6凸ブースト状態に基づいて倍率を決定
		multiplier := 1.0
		if c.StatusIsActive(hexC6BoostKey) {
			multiplier = hexC6Multiplier
		}

		applyEMBuff(multiplier)
		return false
	}

	c.Core.Events.Subscribe(event.OnElectroCharged, ecCallback, "fischl-hex-ec")
	c.Core.Events.Subscribe(event.OnAggravate, ecCallback, "fischl-hex-aggravate")
}

// 6凸 Hexerei ブースト: OnEnemyHit を購読して6凸攻撃を検出
func (c *char) hexC6Boost() {
	if !c.isHexerei || c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		// フィッシュルの6凸攻撃かチェック
		if ae.Info.ActorIndex != c.Index {
			return false
		}
		if ae.Info.Abil != "Fischl C6" {
			return false
		}

		// 6凸ブーストを10秒間有効化
		c.AddStatus(hexC6BoostKey, hexDuration, true)
		return false
	}, "fischl-hex-c6-boost")
}
