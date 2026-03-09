package xianyun

import (
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var burstFrames []int

const (
	burstHeal    = 4 // 最初の回復はヒットマーク後約4フレーム
	burstHitmark = 75
	burstKey     = "xianyun-burst"
	// 16秒間の持続時間
	burstDuration  = 16 * 60
	burstRadius    = 7
	burstDoTRadius = 4.8
	burstDoTDelay  = 5
	// TODO: 脆弱な実装
	// 1回の落下攻撃ダメージで仙助スタックが1つだけ消費されるようにするための処理
	// 現在は万葉の固有天賦1のためだけに必要
	lossKey = "xianyun-burst-loss-icd"
	lossIcd = 3
)

// TODO: 申鶴のフレームデータを仮使用
func init() {
	burstFrames = frames.InitAbilSlice(103) // Q -> J
	burstFrames[action.ActionAttack] = 101  // Q -> N1
	burstFrames[action.ActionCharge] = 102  // Q -> CA
	burstFrames[action.ActionSkill] = 101   // Q -> E
	burstFrames[action.ActionDash] = 101    // Q -> D
	burstFrames[action.ActionWalk] = 101    // Q -> Walk
	burstFrames[action.ActionSwap] = 99     // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.SetCD(action.ActionBurst, 18*60)
	c.ConsumeEnergy(18)
	c.burstCast()
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) burstCast() {
	// 初期回復
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Stars Gather at Dusk (Initial)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
		}

		burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7)
		c.Core.QueueAttack(ai, burstArea, 0, 0)

		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Stars Gather at Dusk Heal (Initial)",
			Src:     healInstantP[c.TalentLvlBurst()]*c.TotalAtk() + healInstantFlat[c.TalentLvlBurst()],
			Bonus:   c.Stat(attributes.Heal),
		})

		c.AddStatus(burstKey, burstDuration, false)

		c.c6()

		for _, char := range c.Core.Player.Chars() {
			// 他のキャラクターが高いジャンプが可能かチェックする仕組みのため、
			// 他のキャラクター自身にバフステータスを付与する必要がある。
			char.AddStatus(player.XianyunAirborneBuff, burstDuration, false)
		}

		c.adeptalAssistStacks = 8

		// TODO: フレームシートによると回復タイミングがかなりばらついている
		for i := burstHeal; i <= burstHeal+burstDuration; i += 2.5 * 60 {
			// ヒットラグの影響を受けない
			c.Core.Tasks.Add(c.burstHealDoT, i)
		}

		c.a4StartUpdate()
	}, burstHitmark)
}

func (c *char) burstPlungeDoTTrigger() {
	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...interface{}) bool {
		// ApplyAttack は攻撃ごとに1回しか発生しないため、ICDステータスを追加する必要はない
		atk := args[0].(*combat.AttackEvent)

		// TODO: 脆弱な実装
		// 雷電の元素爆発中の落下攻撃が落下攻撃ダメージではなく元素爆発ダメージになるため、この形にする必要がある
		if atk.Info.AttackTag != attacks.AttackTagPlunge &&
			!strings.Contains(atk.Info.Abil, "Low Plunge") &&
			!strings.Contains(atk.Info.Abil, "High Plunge") {
			return false
		}

		if atk.Info.Durability == 0 {
			// 落下攻撃の衝突は元素量0
			return false
		}

		active := c.Core.Player.ActiveChar()
		if active.Index != atk.Info.ActorIndex {
			return false
		}
		if !active.StatusIsActive(player.XianyunAirborneBuff) {
			return false
		}

		if c.adeptalAssistStacks <= 0 {
			return false
		}

		if c.StatusIsActive(lossKey) {
			return false
		}
		c.AddStatus(lossKey, lossIcd, false)

		aoe := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, burstDoTRadius)
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Starwicker",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       burstDot[c.TalentLvlBurst()],
		}
		c.Core.QueueAttack(
			ai,
			aoe,
			burstDoTDelay,
			burstDoTDelay,
		)
		c.adeptalAssistStacks--
		c.Core.Log.NewEvent("Xianyun Adeptal Assistance stack consumed", glog.LogPreDamageMod, c.Core.Player.Active()).
			Write("effect_ends_at", c.StatusExpiry(player.XianyunAirborneBuff)).
			Write("stacks_left", c.adeptalAssistStacks)
		if c.adeptalAssistStacks == 0 {
			for _, char := range c.Core.Player.Chars() {
				char.DeleteStatus(player.XianyunAirborneBuff)
			}
		}
		// 固有天賦4が適用できるようにウィンドウを開いたままにする
		c.AddStatus(a4WindowKey, 1, false)
		return false
	}, "xianyun-starwicker-plunge-hook")
}

func (c *char) burstHealDoT() {
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  -1,
		Message: "Starwicker Heal",
		Src:     healDotP[c.TalentLvlBurst()]*c.TotalAtk() + healDotFlat[c.TalentLvlBurst()],
		Bonus:   c.Stat(attributes.Heal),
	})
}
