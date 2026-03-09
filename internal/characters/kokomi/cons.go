package kokomi

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) c1(f, travel int) {
	if c.Base.Cons < 1 {
		return
	}
	if c.Core.Status.Duration("kokomiburst") == 0 {
		return
	}

	// TODO: 1Aと仮定（ライブラリに指定なし）
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "At Water's Edge (C1)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       0,
	}
	ai.FlatDmg = 0.3 * c.MaxHP()

	// TODO: スナップショット/ダイナミックか不明
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 1.2),
		f,
		f+travel,
	)
}

const c4ICDKey = "kokomi-c4-icd"

// 4凸（エネルギー回復のみ）の処理
// 海人の羽衣中、珊瑚宮心海の通常攻撃速度が10%増加。
// 通常攻撃が敵に命中すると、0.8エネルギーを回復する。この効果は0.2秒に1回のみ発動。
func (c *char) makeC4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.Core.Status.Duration("kokomiburst") == 0 {
			return
		}
		if c.StatusIsActive(c4ICDKey) {
			return
		}
		c.AddStatus(c4ICDKey, 0.2*60, true)
		c.AddEnergy("kokomi-c4", 0.8)
	}
}

// 6凸: 海人の羽衣中、通常攻撃・重撃でHP80%以上のキャラクターを
// 回復すると、4秒間水元素ダメージ+40%。
func (c *char) c6() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.HydroP] = .4
	for _, char := range c.Core.Player.Chars() {
		if char.CurrentHPRatio() < 0.8 {
			continue
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("kokomi-c6", 480),
			AffectedStat: attributes.HydroP,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		// 1人見つかれば確認不要
		break
	}
}
