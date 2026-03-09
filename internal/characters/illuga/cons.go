package illuga

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
	c1ICD    = "illuga-c1-icd"
	c1ICDDur = 15 * 60 // 15s
	c4Key    = "illuga-c4-def"
	c4Def    = 200.0
)

// 1命: 岩元素関連の元素反応をトリガーした際にエネルギーを12回復
// 15秒毎に1回のみ発動可能。イルーガがフィールド上の時のみ発動。
func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}

	// 結晶反応をフックする
	cb := func(args ...interface{}) bool {
		// I-12修正: イルーガがフィールド上の時のみ発動
		if c.Core.Player.Active() != c.Index {
			return false
		}
		if c.StatusIsActive(c1ICD) {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}

		c.AddStatus(c1ICD, c1ICDDur, true)
		c.AddEnergy("illuga-c1", 12)

		c.Core.Log.NewEvent("Illuga C1: Energy restored from Geo reaction", glog.LogCharacterEvent, c.Index).
			Write("energy_gained", 12)

		return false
	}

	// 全ての結晶イベントを購読
	c.Core.Events.Subscribe(event.OnCrystallizePyro, cb, "illuga-c1-pyro")
	c.Core.Events.Subscribe(event.OnCrystallizeHydro, cb, "illuga-c1-hydro")
	c.Core.Events.Subscribe(event.OnCrystallizeCryo, cb, "illuga-c1-cryo")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, cb, "illuga-c1-electro")
	c.Core.Events.Subscribe(event.OnLunarCrystallize, cb, "illuga-c1-lcrs")
}

// 2命: Oriole-Song中、ナイチンゲールの歌のスタックが7消費される毎に、
// 元素熟知400% + 防御力200%に基づく岩元素ダメージのランプ攻撃を行う
func (c *char) c2LampAttack() {
	if c.Base.Cons < 2 {
		return
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "C2 Lamp Attack",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
	}

	// FlatDmgを計算: 元素熟知400% + 防御力200%
	em := c.Stat(attributes.EM)
	def := c.TotalDef(false)
	ai.FlatDmg = em*4.0 + def*2.0

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4)
	c.Core.QueueAttack(ai, ap, 0, 10)

	c.Core.Log.NewEvent("Illuga C2: Lamp attack triggered", glog.LogCharacterEvent, c.Index).
		Write("em", em).
		Write("def", def).
		Write("flat_dmg", ai.FlatDmg)
}

// 4命: Oriole-Song中、全パーティメンバーの防御力+200（I-7修正: 以前は自身のみ）
func (c *char) c4Init() {
	if c.Base.Cons < 4 {
		return
	}

	// Oriole-Songアクティブ時に全パーティメンバーに防御力ボーナスを適用
	for _, char := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DEF] = c4Def

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(c4Key, -1),
			AffectedStat: attributes.DEF,
			Amount: func() ([]float64, bool) {
				if !c.orioleSongActive {
					return nil, false
				}
				return m, true
			},
		})
	}
}

// 6命: 月相がAscendant Gleamの時、灯守の誓いのボーナスが2倍になる
// asc.goのapplyLightkeeperOath()で実装済み

func (c *char) consInit() {
	c.c1Init()
	c.c4Init()
}
