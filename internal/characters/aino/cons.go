package aino

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1Key        = "aino-c1"
	c1Duration   = 15 * 60
	c1EMBonus    = 80
	c6Key        = "aino-c6"
	c6Duration   = 15 * 60
	c6DMGBonus   = 0.15
	c6ExtraBonus = 0.20
)

// 1凸 - アイノが元素スキルまたは元素爆発を使用後、
// 元素熔煙が15秒間80アップ。
// 近くのアクティブパーティーメンバーの元素熔煙も15秒間80アップ。
// この元素熔煙アップ効果は重複しない。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.applyC1Buff()
		return false
	}, "aino-c1-skill")

	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.applyC1Buff()
		return false
	}, "aino-c1-burst")
}

func (c *char) applyC1Buff() {
	// 全パーティメンバーに適用
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c1Key, c1Duration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return c.c1EMBuff(), true
			},
		})
	}
	c.Core.Log.NewEvent("aino c1 em buff applied", glog.LogCharacterEvent, c.Index)
}

func (c *char) c1EMBuff() []float64 {
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.EM] = c1EMBonus
	return buff
}

// 2凸 - アイノの元素爆発「精密ハイドロニッククーラー」のゾーンがアクティブな間、
// アイノがフィールド外の時、アクティブメンバーの攻撃が敵に命中すると、
// アヒルが追加の水球を発射し、水元素範囲ダメージ。
// 倍率なし、FlatDmgはアイノの攻撃力25%+元素熔煙100%。
// AttackTagはAttackTagElementalBurstとして扱う。この効果は5秒に1回発動可能。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {

		// アイノがフィールド外か確認
		if c.Core.Player.Active() == c.Index {
			return false
		}

		// 元素爆発がアクティブか確認
		if !c.StatusIsActive(burstKey) {
			return false
		}

		// ICD中か確認
		if c.StatusIsActive(c.c2IcdKey) {
			return false
		}

		// 追加の水球を発射
		atk := c.Stat(attributes.ATK)
		em := c.Stat(attributes.EM)
		flatDmg := 0.25*atk + 1.0*em

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Precision Hydronic Cooler (C2)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
			Mult:       0,
			FlatDmg:    flatDmg,
		}

		// args[0]はターゲット（敵）、args[1]は攻撃イベント
		tgt := args[0].(combat.Target)
		// OnEnemyHitイベントハンドラ内で同期的にスナップショットを生成すると、
		// イベントハンドラの再帰呼び出しで無限再帰が発生する可能性がある。
		// 再帰チェーンを断つため、1フレーム後にスケジュールする。
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(tgt, nil, 3), 1, 1)

		c.AddStatus(c.c2IcdKey, 5*60, false)
		c.Core.Log.NewEvent("aino c2 proc", glog.LogCharacterEvent, c.Index)

		return false
	}, "aino-c2")
}

// 4凸 - 元素スキルが敵に命中するとアイノの元素エネルギーが10回復する。
// この方法によるエネルギー回復は10秒に1回のみ。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != c.Index {
			return false
		}

		if ae.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}

		if c.StatusIsActive("aino-c4-icd") {
			return false
		}

		c.AddEnergy("aino-c4", 10)
		c.AddStatus("aino-c4-icd", 10*60, false)
		c.Core.Log.NewEvent("aino c4 proc", glog.LogCharacterEvent, c.Index)

		return false
	}, "aino-c4")
}

// 6凸 - 元素爆発使用後15秒間、近くのアクティブキャラの感電、開花、
// Lunar-Charged、Lunar-Bloom反応のダメージ+15%。
// ムーンサインが昇順の時、上記反応のダメージがさらに+20%。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}

		c.AddStatus(c6Key, c6Duration, false)
		c.Core.Log.NewEvent("aino c6 buff applied", glog.LogCharacterEvent, c.Index)

		return false
	}, "aino-c6")

	// 感電の反応ダメージボーナスを適用
	c.Core.Events.Subscribe(event.OnElectroCharged, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(c6Key) {
			return false
		}

		bonus := c6DMGBonus
		if c.MoonsignAscendant {
			bonus += c6ExtraBonus
		}
		atk.Info.FlatDmg += atk.Info.FlatDmg * bonus
		c.Core.Log.NewEvent("aino c6 electro-charged dmg bonus", glog.LogCharacterEvent, c.Index).
			Write("bonus", bonus)

		return false
	}, "aino-c6-ec")

	// 開花の反応ダメージボーナスを適用
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(c6Key) {
			return false
		}

		bonus := c6DMGBonus
		if c.MoonsignAscendant {
			bonus += c6ExtraBonus
		}
		atk.Info.FlatDmg += atk.Info.FlatDmg * bonus
		c.Core.Log.NewEvent("aino c6 bloom dmg bonus", glog.LogCharacterEvent, c.Index).
			Write("bonus", bonus)

		return false
	}, "aino-c6-bloom")
}
