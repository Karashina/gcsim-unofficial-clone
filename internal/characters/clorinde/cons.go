package clorinde

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1Icd              int     = 1.2 * 60
	c1AtkP             float64 = 0.3
	c1IcdKey                   = "clorinde-c1-IcdKey"
	c2A1FlatDmg        float64 = 2700
	c2A1PercentBuff    float64 = 0.3
	c6Icd              int     = 12 * 60
	c6IcdKey                   = "clorinde-c6-icd"
	c6Mitigate                 = 0.8
	c6GlimbrightIcdKey         = "glimbrightIcdKey"
	c6GlimbrightAtkP           = 2
)

var c1Hitmarks = []int{1, 1} // TODO: 各１凸ヒットのヒットマーク

// 狩人の律の夜巡状態中、クロリンデの通常攻撃の電元素ダメージが敵に命中した時、
// 命中した敵の近くに召喚された夜巡の影から2回の追撃を行い、
// それぞれクロリンデの攻撃力の30%の電元素ダメージを与える。
// この効果は1.2秒ごとに1回のみ発動。このダメージは
// 通常攻撃ダメージとみなされる。

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if !c.StatusIsActive(skillStateKey) {
			return false
		}
		if c.StatusIsActive(c1IcdKey) {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		if atk.Info.Element != attributes.Electro {
			return false
		}
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		c.AddStatus(c1IcdKey, c1Icd, false)
		c1AI := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Nightwatch Shade (C1)",
			AttackTag:        attacks.AttackTagNormal,
			ICDTag:           attacks.ICDTagClorindeCons,
			ICDGroup:         attacks.ICDGroupClorindeElementalArt,
			StrikeType:       attacks.StrikeTypeSlash,
			Element:          attributes.Electro,
			Durability:       25,
			Mult:             c1AtkP,
			HitlagHaltFrames: 0.01,
			IgnoreInfusion:   true,
		}
		for _, hitmark := range c1Hitmarks {
			c.Core.QueueAttack(
				c1AI,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -3}, 4),
				hitmark,
				hitmark,
				c.particleCB,
			)
		}
		return false
	}, "clorinde-c1")
}

// ラストライトフォールが敵にダメージを与えた時、
// クロリンデの命の契約の割合に応じてダメージが増加する。
// 現在の命の契約1%ごとにラストライトフォールのダメージ2%増加。
// 最大増加は200%。

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("clorinde-c4-burst-bonus", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			m[attributes.DmgP] = min(c.CurrentHPDebtRatio()*100*0.02, 2)
			return m, true
		},
	})
}

// 狩人の律使用後12秒間、
// クロリンデの会心率10%アップ、会心ダメージ70%アップ。
func (c *char) c6skill() {
	if c.Base.Cons < 6 {
		return
	}
	c.c6Stacks = 6
	if !c.StatusIsActive(skillStateKey) {
		return
	}
	if c.StatusIsActive(c6IcdKey) {
		return
	}
	c.AddStatus(c6IcdKey, c6Icd, true)

	mCR := make([]float64, attributes.EndStatType)
	mCR[attributes.CR] = 0.1
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("clorinde-c6-cr-bonus", c6Icd),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return mCR, true
		},
	})

	mCD := make([]float64, attributes.EndStatType)
	mCD[attributes.CD] = 0.7
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("clorinde-c6-cd-bonus", c6Icd),
		AffectedStat: attributes.CD,
		Amount: func() ([]float64, bool) {
			return mCD, true
		},
	})
}

// また、夜巡中に特定の条件でグリムブライトシェードが出現し、
// クロリンデの攻撃力の200%の電元素ダメージを与える攻撃を実行する。
// このダメージは通常攻撃ダメージとみなされる。
func (c *char) c6() {
	if c.StatusIsActive(c6GlimbrightIcdKey) {
		return
	}

	c.c6Stacks--
	c.AddStatus(c6GlimbrightIcdKey, 1*60, true)

	c6AI := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Glimbright Shade (C6)",
		AttackTag:      attacks.AttackTagNormal,
		ICDTag:         attacks.ICDTagClorindeCons,
		ICDGroup:       attacks.ICDGroupClorindeElementalArt,
		StrikeType:     attacks.StrikeTypeSlash,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           c6GlimbrightAtkP,
		IgnoreInfusion: true,
	}
	c.Core.QueueAttack(c6AI, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 8), 0, 0)
}
