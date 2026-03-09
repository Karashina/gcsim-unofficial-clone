package dori

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// トラブルシューターショットが生成するアフターサービス弾の数が1増加する。
func (c *char) c1() {
	c.afterCount++
}

// 戦闘中、ジニーが接続されたキャラクターを回復すると、
// そのキャラクターの位置からドリーの攻撃力50%のジニー弾を発射する。
func (c *char) c2(travel int) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Special Franchise",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagDoriC2,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       0.5,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			1,
		),
		0,
		travel,
	)
}

// ジニーに接続されたキャラクターは現在のHPとエネルギーに応じて以下のバフを得る：
// ・HPが50%未満の場合、受ける治療効果+50%。
// ・エネルギーが50%未満の場合、元素チャージ効率+30%。
func (c *char) c4() {
	active := c.Core.Player.ActiveChar()
	if active.CurrentHPRatio() < 0.5 {
		active.AddHealBonusMod(character.HealBonusMod{
			Base: modifier.NewBaseWithHitlag("dori-c4-healbonus", 48),
			Amount: func() (float64, bool) {
				return 0.5, false
			},
		})
	}
	// 元素チャージ効率を追加
	if active.Energy/active.EnergyMax < 0.5 {
		erMod := make([]float64, attributes.EndStatType)
		erMod[attributes.ER] = 0.3
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("dori-c4-er-bonus", 48),
			AffectedStat: attributes.ER,
			Amount: func() ([]float64, bool) {
				return erMod, true
			},
		})
	}
}

const c6ICD = "dori-c6-heal-icd"
const c6Key = "dori-c6"

// ドリーは魔除の灯使用後3秒間、以下の効果を得る：
// - 雷元素付与。
// - 通常攻撃が敵に命中すると、付近のパーティ全員がドリーのHP上限4%分のHPを回復する。
// この回復は0.1秒に1回発動可能。
func (c *char) makeC6CB() combat.AttackCBFunc {
	if c.Base.Cons < 6 || !c.Core.Player.WeaponInfuseIsActive(c.Index, c6Key) {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		if !c.Core.Player.WeaponInfuseIsActive(c.Index, c6Key) {
			return
		}
		if c.StatusIsActive(c6ICD) {
			return
		}
		c.AddStatus(c6ICD, 0.1*60, true)

		// パーティメンバーを回復
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "dori-c6-heal",
			Src:     0.04 * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}
}
