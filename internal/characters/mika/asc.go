package mika

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1Stacks = "detector-stacks"
	a1Buff   = "detector-buff"
)

func (c *char) addDetectorStack() {
	stacks := c.Tag(a1Stacks)

	if stacks < c.maxDetectorStacks {
		stacks++
		c.Core.Log.NewEvent("add detector stack", glog.LogCharacterEvent, c.Index).
			Write("stacks", stacks).
			Write("maxstacks", c.maxDetectorStacks)
	}
	c.SetTag(a1Stacks, stacks)
}

// 以下の条件で、星霜の旋による極寒の風状態はキャラクターにディテクターを付与し、
// フィールド上にいる時、物理ダメージが10%増加する。
// - 霜流れの矢が複数の敵に命中した場合、追加の敵1体につきディテクタースタックが1つ生成される。
// - 霜星の破片が敵に命中すると、ディテクタースタックが1つ生成される。各破片は1回のみ効果を発動できる。
//
// 極寒の風状態は最大3つのディテクタースタックを持つことができ、この持続中に星霜の旋を再使用すると、
// 既存の極寒の風状態とすべてのディテクタースタックがクリアされる。
func (c *char) a1(char *character.CharWrapper) {
	m := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(a1Buff, skillBuffDuration),
		AffectedStat: attributes.PhyP,
		Amount: func() ([]float64, bool) {
			m[attributes.PhyP] = 0.1 * float64(c.Tag(a1Stacks))
			return m, c.Core.Player.Active() == char.Index
		},
	})
}

// 星翼の歌の鷲羽と星霜の旋の極寒の風の両方の効果を受けたアクティブキャラクターが攻撃で会心した場合、
// 極寒の風は索敵弾幕のディテクタースタックを1つ付与する。1回の極寒の風につき、この方法で1スタックのみ獲得可能。
// さらに、極寒の風のみで獲得できるスタックの最大数が1つ増加する。
// 索敵弾幕のアンロックが必要。
func (c *char) a4() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if c.a4Stack {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		char := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if char.Index != c.Core.Player.Active() {
			return false
		}

		if !char.StatModIsActive(skillBuffKey) || !c.StatusIsActive(healKey) {
			return false
		}

		crit := args[3].(bool)
		if !crit {
			return false
		}

		c.addDetectorStack()
		c.a4Stack = true
		return false
	}, "mika-a4")
}
