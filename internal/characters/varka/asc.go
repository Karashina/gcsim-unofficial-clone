package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// a1Init はA1パッシブを設定する: 攻撃力1000ごとに風+他元素ダメージボーナス10%（最大25%）
// TotalAtk()のStatMod内での無限再帰を避けるためAttackModを使用
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	otherP := attributes.NoStat
	if c.hasOtherEle {
		otherP = eleToStatP(c.otherElement)
	}

	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a1Key, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			totalAtk := c.TotalAtk()
			bonus := totalAtk / 1000.0 * 0.10
			if bonus > 0.25 {
				bonus = 0.25
			}
			// リセット
			for i := range m {
				m[i] = 0
			}
			m[attributes.AnemoP] = bonus
			if otherP != attributes.NoStat {
				m[otherP] = bonus
			}
			return m, true
		},
	})
}

// a4Init は拡散イベントを購読してAzure Fang's Oathスタックを管理する
func (c *char) a4Init() {
	swirlEvents := []event.Event{
		event.OnSwirlHydro,
		event.OnSwirlPyro,
		event.OnSwirlCryo,
		event.OnSwirlElectro,
	}
	for _, ev := range swirlEvents {
		c.Core.Events.Subscribe(ev, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			// ICD: 各キャラクターは1秒に1スタックのみ付与可能
			charIdx := atk.Info.ActorIndex
			icdKey := fmt.Sprintf("%s%d", a4ICDPrefix, charIdx)
			if c.StatusIsActive(icdKey) {
				return false
			}
			c.AddStatus(icdKey, 60, true) // キャラクターごとに1秒ICD

			c.a4Stacks++
			if c.a4Stacks > 4 {
				c.a4Stacks = 4
			}
			c.a4Expiry = c.Core.F + 8*60 // 8s duration, refreshed on new stack

			c.a4Apply()
			return false
		}, fmt.Sprintf("varka-a4-%v", ev))
	}
}

// a4Apply はA4のダメージボーナスバフを適用する
func (c *char) a4Apply() {
	// 有効期限切れか確認
	stacks := c.a4Stacks
	if c.Core.F >= c.a4Expiry {
		stacks = 0
		c.a4Stacks = 0
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = float64(stacks) * 0.075

	// C6: 各スタックで会心ダメージ20%も付与
	if c.Base.Cons >= 6 {
		m[attributes.CD] = float64(stacks) * 0.20
	}

	dur := c.a4Expiry - c.Core.F
	if dur <= 0 {
		return
	}

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a4Key, dur),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			// 通常攻撃、重撃、Azure Devour、FWAにのみ適用
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra:
				return m, true
			case attacks.AttackTagElementalArt:
				// FWAのみ、Windbound Executionには適用しない
				if atk.Info.Abil == "Four Winds' Ascension (Other)" ||
					atk.Info.Abil == "Four Winds' Ascension (Anemo)" {
					return m, true
				}
				return nil, false
			default:
				return nil, false
			}
		},
	})
}

// eleToStatP は元素を対応するダメージ%ステータスに変換する
func eleToStatP(ele attributes.Element) attributes.Stat {
	switch ele {
	case attributes.Pyro:
		return attributes.PyroP
	case attributes.Hydro:
		return attributes.HydroP
	case attributes.Electro:
		return attributes.ElectroP
	case attributes.Cryo:
		return attributes.CryoP
	case attributes.Anemo:
		return attributes.AnemoP
	default:
		return attributes.NoStat
	}
}
