package razor

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 元素オーブまたは元素粒子を拾うと、Razorのダメージが8秒間10%上昇する。
func (c *char) c1() {
	c.c1bonus = make([]float64, attributes.EndStatType)
	c.c1bonus[attributes.DmgP] = 0.1

	c.Core.Events.Subscribe(event.OnParticleReceived, func(_ ...interface{}) bool {
		// キャラクターがフィールドにいなければ無視
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("razor-c1", 8*60),
			AffectedStat: attributes.DmgP,
			Amount: func() ([]float64, bool) {
				return c.c1bonus, true
			},
		})
		return false
	}, "razor-c1")
}

// HPが30%未満の敵に対して会心率が10%上昇する。
func (c *char) c2() {
	if c.Core.Combat.DamageMode {
		c.c2bonus = make([]float64, attributes.EndStatType)
		c.c2bonus[attributes.CR] = 0.1

		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("razor-c2", -1),
			Amount: func(_ *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				x, ok := t.(*enemy.Enemy)
				if !ok {
					return nil, false
				}
				if x.HP()/x.MaxHP() < 0.3 {
					return c.c2bonus, true
				}
				return nil, false
			},
		})
	}
}

// 爪雷（単押し）で命中した敵の防御力が7秒間15%減少する。
func (c *char) c4cb(a combat.AttackCB) {
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddDefMod(combat.DefMod{
		Base:  modifier.NewBaseWithHitlag("razor-c4", 7*60),
		Value: -0.15,
	})
}

const c6ICDKey = "razor-c6-icd"

// 10秒ごとにRazorの剣が充電され、次の通常攻撃がRazorの攻撃力100%分の雷元素ダメージを与える電撃を放つ。
// 雷牙未使用時、敵への電撃は爪雷用の雷印を付与する。
func (c *char) c6cb(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	// 効果は10秒ごとにのみ発動
	if c.StatusIsActive(c6ICDKey) {
		return
	}

	c.AddStatus(c6ICDKey, 600, true)

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lupus Fulguris",
		AttackTag:  attacks.AttackTagNone, // TODO: 別のタグがある？
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       1,
	}

	sigilcb := func(a combat.AttackCB) {
		// 爆発外でのみ雷印を追加
		if c.StatusIsActive(burstBuffKey) {
			return
		}
		c.addSigil(false)(a)
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), a.Target, geometry.Point{Y: 0.7}, 1.5),
		1,
		1,
		sigilcb,
	)
}
