package qiqi

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

var burstFrames []int

const burstHitmark = 82

func init() {
	burstFrames = frames.InitAbilSlice(115) // Q -> D
	burstFrames[action.ActionAttack] = 113  // Q -> N1
	burstFrames[action.ActionSkill] = 113   // Q -> E
	burstFrames[action.ActionJump] = 114    // Q -> J
	burstFrames[action.ActionSwap] = 112    // Q -> Swap
}

// 元素爆発ダメージのみを適用。主な箓機能はqiqi.goで処理
func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Fortune-Preserving Talisman",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 50,
		Mult:       burstDmg[c.TalentLvlBurst()],
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7)
	c.Core.QueueAttack(ai, ap, burstHitmark, burstHitmark)

	// 箓はダメージが与えられる前に0ダメージ攻撃で適用される
	talismanAi := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Fortune-Preserving Talisman (Talisman application)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Physical,
	}
	talismanCB := func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddStatus(talismanKey, 15*60, true)
	}
	c.Core.QueueAttack(talismanAi, ap, 40, 40, talismanCB)

	c.SetCD(action.ActionBurst, 20*60)
	c.ConsumeEnergy(8)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) talismanHealHook() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		e, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}

		// 箓が期限切れなら何もしない
		if !e.StatusIsActive(talismanKey) {
			return false
		}
		// 箓がまだICD中なら何もしない
		if e.GetTag(talismanICDKey) >= c.Core.F {
			return false
		}

		healAmt := c.healDynamic(burstHealPer, burstHealFlat, c.TalentLvlBurst())
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  atk.Info.ActorIndex,
			Message: "Fortune-Preserving Talisman",
			Src:     healAmt,
			Bonus:   c.Stat(attributes.Heal),
		})
		e.SetTag(talismanICDKey, c.Core.F+60)

		return false
	}, "talisman-heal-hook")
}

// 2命ノ星座、固有天賦4、および元素スキルの通常/重撃命中時フックを処理
// また元素爆発の箓フックも処理 - 箓は元素爆発ダメージより前に適用されるため、他の方法では実現できない
func (c *char) onNACAHitHook() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		e, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != c.Index {
			return false
		}

		// 以下はすべて七七の通常攻撃/重撃命中時のみ発動
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		default:
			return false
		}

		// 固有天賤2:
		// 七七が通常攻撃と重撃で敵に命中したとき、
		// 50%の確率で敵に寿命の箓を付与する（6秒間）。
		// この効果は30秒に1回のみ発動。
		if c.Base.Ascension >= 4 && !c.StatusIsActive(a4ICDKey) && (c.Core.Rand.Float64() < 0.5) {
			// より長い元素爆発箓の持続時間を短いもので上書きしたくない
			// TODO: 既に敵に箓がある場合の相互作用は不明
			// TODO: 現状は競合がある場合CDに置かない寛大な処理
			if e.StatusExpiry(talismanKey) < c.Core.F+360 {
				e.AddStatus(talismanKey, 360, true)
				c.AddStatus(a4ICDKey, 1800, true) // 30秒ICD
				c.Core.Log.NewEvent(
					"Qiqi A4 Adding Talisman",
					glog.LogCharacterEvent,
					c.Index,
				).
					Write("target", e.Key()).
					Write("talisman_expiry", e.StatusExpiry(talismanKey))
			}
		}

		// 七七の元素スキル持続中の通常/重撃回復処理
		if c.StatusIsActive(skillBuffKey) {
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  -1,
				Message: "Herald of Frost (Attack)",
				Src:     c.healSnapshot(&c.skillHealSnapshot, skillHealOnHitPer, skillHealOnHitFlat, c.TalentLvlSkill()),
				Bonus:   c.skillHealSnapshot.Stats[attributes.Heal],
			})
		}

		return false
	}, "qiqi-onhit-naca-hook")
}
