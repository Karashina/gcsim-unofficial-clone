package baizhu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c2ICDKey = "baizhu-c2-icd"

// 木鑰の薬診に追加チャージ1回を付与する。
func (c *char) c1() {
	c.SetNumCharges(action.ActionSkill, 2)
}

// アクティブキャラクターの攻撃が近くの敵に命中すると、白朮は遊糸徴霊・結を放つ。
// 遊糸徴霊・結は1回攻撃してから帰還し、白朮の攻撃力250%の草元素ダメージを与え、
// 木鑰の薬診の遊糸徴霊の通常回復量の20%を回復する。
// この方法で与えたダメージは元素スキルダメージとみなされる。
// この効果は5秒に1回発動可能。
func (c *char) c2() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		t := args[0].(combat.Target)
		// アクティブキャラクターでのみ発動
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if c.StatusIsActive(c2ICDKey) {
			return false
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Gossamer Sprite: Splice. (Baizhu's C2)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupBaizhuC2,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       2.5,
		}
		c.c6done = false
		var c6cb combat.AttackCBFunc
		if c.Base.Cons >= 6 {
			c6cb = c.makeC6CB()
		}
		// TODO: 正確な2凸のヒットマークと帰還移動値
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(t, nil, 0.6),
			0,
			skillFirstHitmark, // 今のところスキルを再利用
			c6cb,
		)

		// 2凸の回復
		c.Core.Tasks.Add(func() {
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  c.Core.Player.Active(),
				Message: "Baizhu's C2: Healing",
				Src:     (skillHealPP[c.TalentLvlSkill()]*c.MaxHP() + skillHealFlat[c.TalentLvlSkill()]) * 0.2,
				Bonus:   c.Stat(attributes.Heal),
			})
		}, skillReturnTravel) // 今のところスキルを再利用

		c.AddStatus(c2ICDKey, 60*5, false) // 5s
		return false
	}, "baizhu-c2")
}

// 癒しの秘法を使用後15秒間、白朮は近くのチーム全員の元素熟知を80上昇させる。
func (c *char) c4() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 80
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("baizhu-c4", 900),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

// 癒しの霊脈のダメージを白朮のHP上限の8%分増加させる。
// また、遊糸徴霊または遊糸徴霊・結が敵に命中すると、100%の確率で
// 癒しの秘法の継ぎ目なきシールドを1つ生成する。この効果は1回の遊糸徴霊につき1回のみ発動。
func (c *char) makeC6CB() combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if done {
			return
		}
		done = true
		c.summonSeamlessShield()
	}
}
