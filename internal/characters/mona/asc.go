package mona

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 虚実流動を2秒間使用した後、付近に敵がいる場合、
// モナは自動的に虚影を生成する。
// この方法で生成された虚影は2秒間持続し、爆発ダメージは水中幻願の50%に等しい。
//
// - 突破レベルチェックに失敗するだけのキューイングを避けるため、突破レベルはdash.goで確認
func (c *char) a1() {
	// モナでなければ何もしない
	if c.Core.Player.Active() != c.Index {
		return
	}
	// ダッシュ中でなければ何もしない
	if c.Core.Player.CurrentState() != action.DashState {
		return
	}
	enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 15), nil)
	if enemies != nil {
		c.Core.Log.NewEvent("mona-a1 phantom added", glog.LogCharacterEvent, c.Index).
			Write("expiry:", c.Core.F+120)
		// 虚影の爆発をキューに追加
		phantomPos := c.Core.Combat.Player()
		c.Core.Tasks.Add(func() {
			aiExplode := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Mirror Reflection of Doom (A1 Explode)",
				AttackTag:  attacks.AttackTagElementalArt,
				ICDTag:     attacks.ICDTagNone,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeDefault,
				Element:    attributes.Hydro,
				Durability: 25,
				Mult:       0.5 * skill[c.TalentLvlSkill()],
			}
			c.Core.QueueAttack(aiExplode, combat.NewCircleHitOnTarget(phantomPos, nil, 5), 0, 0)
		}, 120)
	}
	// モナがまだダッシュ中のため次の固有天賦1チェックをキューに追加
	// 異なる虚影は共存し、互いに上書きしない
	c.Core.Tasks.Add(c.a1, 120) // 2秒後に再チェック
}

// モナの水元素ダメージバフを元素チャージ効率の20%に相当する量だけ増加させる。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	if c.a4Stats == nil {
		c.a4Stats = make([]float64, attributes.EndStatType)
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("mona-a4", -1),
			AffectedStat: attributes.HydroP,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return c.a4Stats, true
			},
		})
	}
	c.a4Stats[attributes.HydroP] = 0.2 * c.NonExtraStat(attributes.ER)
	c.QueueCharTask(c.a4, 60)
}
