package itto

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 荒瀧一斗が連続で荒瀧キサギリを使用すると、以下の効果を得る:
//
// - 各斬撃で次の斬撃の攻撃速度が10%増加する。最大攻撃速度増加は30%。
//
// TODO: - 中断耐性が増加する。
//
// これらの効果は連続斬撃を停止すると解除される。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	mAtkSpd := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("itto-a1", -1),
		AffectedStat: attributes.AtkSpd,
		Amount: func() ([]float64, bool) {
			if c.a1Stacks == 0 || c.Core.Player.CurrentState() != action.ChargeAttackState {
				return nil, false
			}
			mAtkSpd[attributes.AtkSpd] = 0.10 * float64(c.a1Stacks)
			return mAtkSpd, true
		},
	})
}

func (c *char) a1Update(curSlash SlashType) {
	if c.Base.Ascension < 1 {
		return
	}
	switch curSlash {
	case SaichiSlash:
		// CA0の場合はA1スタックをリセット
		c.a1Stacks = 0
		c.Core.Log.NewEvent("itto-a1 reset atkspd stacks", glog.LogCharacterEvent, c.Index).
			Write("a1Stacks", c.a1Stacks).
			Write("slash", curSlash.String(false))
	case LeftSlash, RightSlash:
		// CA1/CA2の場合はA1スタックを増加
		// A1スタックを増加、最大3スタック
		c.a1Stacks++
		if c.a1Stacks > 3 {
			c.a1Stacks = 3
		}
		c.Core.Log.NewEvent("itto-a1 atkspd stacks increased", glog.LogCharacterEvent, c.Index).
			Write("a1Stacks", c.a1Stacks).
			Write("slash", curSlash.String(false))
	}
	// CAFの場合は何もしない
}

// 荒瀧キサギリのダメージが荒瀧一斗の防御力の35%分増加する。
func (c *char) a4(ai *combat.AttackInfo) {
	if c.Base.Ascension < 4 {
		return
	}
	ai.FlatDmg = c.TotalDef(false) * 0.35
	c.Core.Log.NewEvent("itto-a4 applied", glog.LogCharacterEvent, c.Index)
}
