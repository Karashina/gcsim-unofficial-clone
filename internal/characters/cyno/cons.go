package cyno

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1Key    = "cyno-c1"
	c6Key    = "cyno-c6"
	c6ICDKey = "cyno-c6-icd"
)

// 「圣儀・狼駆」使用後、セノの通常攻撃速度が10秒間20%増加する。
// 「秘儀・裂置の将」中に固有天賦「羽落ちの裁定」の「裁定」効果が発動した場合、
// この増加の持続時間がリフレッシュされる。
//
// 固有天賦「羽落ちの裁定」を先に解放する必要がある。
func (c *char) c1() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.AtkSpd] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c1Key, 600), // 10s
		AffectedStat: attributes.AtkSpd,
		Amount: func() ([]float64, bool) {
			if c.Core.Player.CurrentState() != action.NormalAttackState {
				return nil, false
			}
			return m, true
		},
	})
}

const c2Key = "cyno-c2"
const c2ICD = "cyno-c2-icd"

// セノの通常攻撃が敵に命中すると、雷元素ダメージボーナスが
// 4秒間10%増加する。0.1秒に1回発動可能。最大5スタック。
func (c *char) makeC2CB() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(c2ICD) {
			return
		}
		c.AddStatus(c2ICD, 0.1*60, true)

		if !c.StatModIsActive(c2Key) {
			c.c2Stacks = 0
		}
		c.c2Stacks++
		if c.c2Stacks > 5 {
			c.c2Stacks = 5
		}

		m := make([]float64, attributes.EndStatType)
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2Key, 4*60),
			AffectedStat: attributes.ElectroP,
			Amount: func() ([]float64, bool) {
				m[attributes.ElectroP] = 0.1 * float64(c.c2Stacks)
				return m, true
			},
		})
	}
}

// 「圣儀・狼駆」で発動した「契約の导砂者」状態のセノが、
// 感電・超伝導・過負荷・激化・超激化・超開花・雷拡散反応を
// 発動した際、周囲のパーティメンバー（自身を除く）の元素エネルギーを3回復。
// 「圣儀・狼駆」1回につき5回まで発動可能。
func (c *char) c4() {
	//nolint:unparam // 今は無視、イベントリファクタで bool 戻り値が解決されるはず
	restore := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if c.c4Counter > 4 { // 0〜4のカウント、最大5回
			return false
		}
		c.c4Counter++
		for _, this := range c.Core.Player.Chars() {
			// サイノ以外
			if this.Index != c.Index {
				this.AddEnergy("cyno-c4", 3)
			}
		}

		return false
	}

	restoreNoGadget := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		return restore(args...)
	}
	c.Core.Events.Subscribe(event.OnOverload, restoreNoGadget, "cyno-c4")
	c.Core.Events.Subscribe(event.OnElectroCharged, restoreNoGadget, "cyno-c4")
	c.Core.Events.Subscribe(event.OnSuperconduct, restoreNoGadget, "cyno-c4")
	c.Core.Events.Subscribe(event.OnQuicken, restoreNoGadget, "cyno-c4")
	c.Core.Events.Subscribe(event.OnAggravate, restoreNoGadget, "cyno-c4")
	c.Core.Events.Subscribe(event.OnHyperbloom, restore, "cyno-c4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, restoreNoGadget, "cyno-c4")
}

// 「圣儀・狼駆」または固有天賦「羽落ちの裁定」の「裁定」効果発動後、
// 「ジャッカルの日」効果を4スタック獲得する。
func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}
	c.AddStatus(c6Key, 8*60, true)
	c.c6Stacks += 4
	if c.c6Stacks > 8 {
		c.c6Stacks = 8
	}
}

// 通常攻撃が敵に命中すると「ジャッカルの日」スタックを1消費し、「砂の矢」を1本発射する。
// 「ジャッカルの日」は8秒持続。最大8スタック。「契約の导砂者」終了時に解除される。
// 0.4秒に1本の「砂の矢」が発射可能。
// 固有天賦「羽落ちの裁定」を先に解放する必要がある。
func (c *char) makeC6CB() combat.AttackCBFunc {
	if c.Base.Cons < 6 || c.c6Stacks == 0 || !c.StatusIsActive(c6Key) {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.c6Stacks == 0 {
			return
		}
		if !c.StatusIsActive(c6Key) {
			return
		}
		if c.StatusIsActive(c6ICDKey) {
			return
		}
		c.AddStatus(c6ICDKey, 0.4*60, true)
		c.c6Stacks--

		// 技術的には ICDGroupCynoC6 を使うべきだが、実質的には標準ICDと同じ
		ai := combat.AttackInfo{
			ActorIndex:   c.Index,
			Abil:         "Raiment: Just Scales (C6)",
			AttackTag:    attacks.AttackTagElementalArtHold,
			ICDTag:       attacks.ICDTagElementalArt,
			ICDGroup:     attacks.ICDGroupDefault,
			StrikeType:   attacks.StrikeTypeSlash,
			Element:      attributes.Electro,
			Durability:   25,
			IsDeployable: true,
			Mult:         1.0,
			FlatDmg:      c.a4Bolt(),
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				nil,
				0.3,
			),
			0,
			0,
		)
	}
}
