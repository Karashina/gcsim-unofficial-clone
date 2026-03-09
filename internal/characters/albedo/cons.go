package albedo

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// 創生術・擬似陽花(Solar Isotoma)が生成するTransient BlossomはアルベドにFatal Reckoningを30秒間付与:
	// • 誕生式・大地の潮を発動すると全Fatal Reckoningスタックを消費。消費したスタック毎にFatal Blossomと元素爆発ダメージがアルベドの防御力の30%分増加。
	// • 最大4スタックまで累積可能。
	c2key = "albedo-c2"

	// 1凸防御力バフキー
	c1DEFKey      = "albedo-c1-def"
	c1DEFDuration = 20 * 60 // 20s

	// 6凸追加バフキー
	c6BlossomBuffKey      = "albedo-c6-blossom"
	c6BlossomBuffDuration = 20 * 60 // 20s
)

// 1凸追加バフ:
// アルベドの元素スキル使用時、防御力が20秒間50%増加する。
func (c *char) c1DEFBuff() {
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DEFP] = 0.50
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c1DEFKey, c1DEFDuration),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
	c.Core.Log.NewEvent("Albedo C1: DEF +50% applied", glog.LogCharacterEvent, c.Index)
}

// 2凸追加バフ:
// アルベドがフィールド外でFatal Reckoningのスタックが4に達した時、
// 全スタックを消費し、キャラクター付近に3個のFatal Blossomを生成。
// アルベドの防御力の300%に相当する岩元素範囲ダメージを与える（元素爆発ダメージ扱い）。
// 固有天賦「Homuncular Nature」が解放済みの場合、
// 周囲のパーティメンバーの元素熟知もた10秒間125増加する。
func (c *char) c2AutoBlossom() {
	if c.Base.Cons < 2 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// アルベドがフィールド外かチェック
		if c.Core.Player.Active() == c.Index {
			return false
		}
		// 4スタックあるかチェック
		if !c.StatusIsActive(c2key) || c.c2stacks < 4 {
			return false
		}

		// 全スタックを消費
		stacks := c.c2stacks
		c.c2stacks = 0
		c.DeleteStatus(c2key)

		// 3個のFatal Blossomを生成
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "C2 Fatal Blossom",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   30,
			Element:    attributes.Geo,
			Durability: 25,
			FlatDmg:    c.TotalDef(false) * 3.0, // 防御力300%
		}

		activePos := c.Core.Combat.Player().Pos()
		for i := 0; i < 3; i++ {
			center := geometry.CalcRandomPointFromCenter(activePos, 0.5, 3.0, c.Core.Rand)
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(center, nil, 3),
				30+i*10, // 遅延
				30+i*10,
			)
		}

		// Homuncular Nature (A4)が解放済みならEMバフを適用
		if c.Base.Ascension >= 4 {
			m := make([]float64, attributes.EndStatType)
			m[attributes.EM] = 125
			for _, char := range c.Core.Player.Chars() {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("albedo-c2-em", 600), // 10s
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}
		}

		c.Core.Log.NewEvent("Albedo C2: Auto Fatal Blossom triggered", glog.LogCharacterEvent, c.Index).
			Write("stacks_consumed", stacks)

		return false
	}, "albedo-c2-auto-blossom")
}

// 4凸:
// 陽華のフィールド内のアクティブパーティメンバーの落下攻撃ダメージが30%増加。
// 追加（Hexerei）：近くのアクティブキャラクターがSilver Isotomaの近くでジャンプすると、Silver Isotomaは
// 破壊されジャンプの高さが大幅に増加する。その後3秒間、そのキャラクターの
// 落下攻撃の着地ダメージが30%増加する。
func (c *char) c4(lastConstruct int) func() {
	return func() {
		// スキルが再使用された場合は再適用/再チェックしない
		if c.lastConstruct != lastConstruct {
			return
		}
		// スキルがもうアクティブでない場合は再適用/再チェックしない
		if !c.skillActive {
			return
		}

		// スキル範囲内なら1秒間アクティブキャラにC4バフを適用
		inSkillArea := c.Core.Combat.Player().IsWithinArea(c.skillArea)
		if inSkillArea {
			active := c.Core.Player.ActiveChar()
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = 0.3
			active.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("albedo-c4", 60), // 1s
				Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
					if atk.Info.AttackTag != attacks.AttackTagPlunge {
						return nil, false
					}
					return m, true
				},
			})
		}

		// 0.3秒後に再チェック
		c.Core.Tasks.Add(c.c4(lastConstruct), 18)
	}
}

// c4HexereiJumpBuff: Hexerei 4凸追加効果
// Silver Isotoma付近でジャンプすると破壊し、落下攻撃ダメージ+30%を3秒間付与
func (c *char) c4HexereiJumpBuff() {
	if c.Base.Cons < 4 {
		return
	}
	if !c.isHexerei {
		return
	}

	c.Core.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		// ジャンプアクションかチェック
		// 注意: gcsimはジャンプを完全にシミュレートしないため簡略化実装
		// 落下攻撃開始時に効果を適用
		return false
	}, "albedo-c4-hexerei-jump")

	// 代替: 落下攻撃時に適用
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		// 落下攻撃かつスキル範囲内かチェック
		if atk.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		if !c.skillActive {
			return false
		}
		if !c.Core.Combat.Player().IsWithinArea(c.skillArea) {
			return false
		}

		// 設置物を破壊
		c.Core.Constructs.Destroy(c.lastConstruct)
		c.skillActive = false

		// 攻撃者に落下攻撃ダメージ+30%を3秒間適用
		attacker := c.Core.Player.Chars()[atk.Info.ActorIndex]
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.3
		attacker.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("albedo-c4-hexerei", 180), // 3s
			Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagPlunge {
					return nil, false
				}
				return m, true
			},
		})

		c.Core.Log.NewEvent("Albedo C4 Hexerei: Silver Isotoma destroyed, Plunge DMG buff applied", glog.LogCharacterEvent, c.Index)

		return false
	}, "albedo-c4-hexerei-plunge")
}

// 6凸:
// Solar Isotomaフィールド内の結晶反応シールドに守られたアクティブパーティメンバーのダメージが17%増加。
func (c *char) c6(lastConstruct int) func() {
	return func() {
		// スキルが再使用された場合は再適用/再チェックしない
		if c.lastConstruct != lastConstruct {
			return
		}
		// スキルがもうアクティブでない場合は再適用/再チェックしない
		if !c.skillActive {
			return
		}

		// 結晶化シールド保有かつスキル範囲内のアクティブキャラに6凸バフを1秒間適用
		crystallizeShield := c.Core.Player.Shields.Get(shield.Crystallize) != nil
		inSkillArea := c.Core.Combat.Player().IsWithinArea(c.skillArea)
		if crystallizeShield && inSkillArea {
			active := c.Core.Player.ActiveChar()
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = 0.17
			active.AddStatMod(character.StatMod{
				Base:         modifier.NewBase("albedo-c6", 60), // 1s
				AffectedStat: attributes.DmgP,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		// 0.3秒後に再チェック
		c.Core.Tasks.Add(c.c6(lastConstruct), 18)
	}
}

// c6BlossomBuffOnBurst: 6凸追加効果 (Hexerei)
// 元素爆発使用時に全Silver Isotomaを破壊し、破壊ごとにFatal Reckoningを4スタック消費、
// Fatal Blossomダメージを防御力の250%分增加（20秒間）。
func (c *char) c6BlossomBuffOnBurst() {
	if c.Base.Cons < 6 {
		return
	}
	if !c.isHexerei {
		return
	}

	// 破壊したSilver Isotomaの数を追跡
	if c.skillActive {
		// Silver Isotomaを破壊
		c.Core.Constructs.Destroy(c.lastConstruct)
		c.skillActive = false

		// 破壊したIsotomaごとにFatal Reckoningを4スタック除去
		if c.c2stacks >= 4 {
			c.c2stacks -= 4
		} else {
			c.c2stacks = 0
		}
		if c.c2stacks == 0 {
			c.DeleteStatus(c2key)
		}

		// Fatal Blossomダメージ+防御力250%を20秒間付与
		c.AddStatus(c6BlossomBuffKey, c6BlossomBuffDuration, true)

		c.Core.Log.NewEvent("Albedo C6 Hexerei: Silver Isotoma destroyed on Burst", glog.LogCharacterEvent, c.Index).
			Write("remaining_c2_stacks", c.c2stacks)
	}
}
