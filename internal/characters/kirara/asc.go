package kirara

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1IcdStatus = "kirara-a1-icd"
)

// 緋球玄縄のUrgent Neko Parcel状態の時、敵に当たるたびに強化梱包のスタックを獲得する。
// この効果は各敵に対して0.5秒に1回発動可能。最大3スタック。Urgent Neko Parcel状態終了時、
// 各スタックにつき1枚の安全輸送シールドを生成する。このシールドは
// 緋球玄縄が生成する安全輸送シールドのダメージ吸収量の20%を持つ。
// 既に緋球玄縄の安全輸送シールドがある場合、
// ダメージ吸収量が累積し持続時間がリセットされる。
func (c *char) a1StackGain(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(a1IcdStatus) {
		return
	}
	if c.a1Stacks >= 3 {
		return
	}
	c.a1Stacks++
	c.AddStatus(a1IcdStatus, 0.5*60, true)
}

func (c *char) a1() {
	shieldamt := c.shieldHP() * 0.2 * float64(c.a1Stacks)
	c.genShield("Shield of Safe Transport", shieldamt)
}

// キララのHP上限1,000ごとに緋球玄縄のダメージが0.4%増加し、秘技・緋球天眼通のダメージが0.3%増加する。
func (c *char) a4() {
	mSkill := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("kirara-a4-skill", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
				return nil, false
			}
			mSkill[attributes.DmgP] = c.MaxHP() * 0.001 * 0.004
			return mSkill, true
		},
	})

	mBurst := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("kirara-a4-burst", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			mBurst[attributes.DmgP] = c.MaxHP() * 0.001 * 0.003
			return mBurst, true
		},
	})
}
