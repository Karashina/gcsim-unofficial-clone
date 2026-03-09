package raiden

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var burstFrames []int

const (
	burstHitmark = 98
	BurstKey     = "raidenburst"
)

func init() {
	burstFrames = frames.InitAbilSlice(112) // Q -> J
	burstFrames[action.ActionAttack] = 111  // Q -> N1
	burstFrames[action.ActionCharge] = 500  // TODO: このアクションは無効
	burstFrames[action.ActionSkill] = 111   // Q -> E
	burstFrames[action.ActionDash] = 111    // Q -> D
	burstFrames[action.ActionSwap] = 110    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 元素爆発を発動、愿力スタックをリセット
	c.burstCastF = c.Core.F
	c.restoreCount = 0
	c.restoreICD = 0
	c.c6Count = 0
	c.c6ICD = 0

	// 特殊の修飾子で元素爆発状態を追跡
	c.AddStatus(BurstKey, 420+burstHitmark, true)

	// 元素爆発終了時に適用
	if c.Base.Cons >= 4 {
		c.applyC4 = true
		src := c.burstCastF
		c.QueueCharTask(func() {
			if src == c.burstCastF && c.applyC4 {
				c.applyC4 = false
				c.c4()
			}
		}, 420+burstHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Musou Shinsetsu",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Electro,
		Durability: 50,
		Mult:       burstBase[c.TalentLvlBurst()],
	}

	if c.Base.Cons >= 2 {
		ai.IgnoreDefPercent = 0.6
	}

	c.Core.Tasks.Add(func() {
		c.stacksConsumed = c.stacks
		c.stacks = 0
		ai.Mult += resolveBaseBonus[c.TalentLvlBurst()] * c.stacksConsumed
		c.Core.Log.NewEvent("resolve stacks", glog.LogCharacterEvent, c.Index).
			Write("stacks", c.stacksConsumed)
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.1}, 13, 8),
			0,
			0,
		)
	}, burstHitmark)

	c.SetCD(action.ActionBurst, 18*60)
	c.ConsumeEnergy(8)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) burstRestorefunc(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.Core.F > c.restoreICD && c.restoreCount < 5 {
		c.restoreCount++
		c.restoreICD = c.Core.F + 60 // once every 1 second
		energy := burstRestore[c.TalentLvlBurst()] * (1 + c.a4Energy(max(c.NonExtraStat(attributes.ER)-1, 0)))
		for _, char := range c.Core.Player.Chars() {
			char.AddEnergy("raiden-burst", energy)
		}
	}
}

func (c *char) onSwapClearBurst() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		if !c.StatusIsActive(BurstKey) {
			return false
		}
		// 前のキャラが誰かのチェックは不要と思われる
		prev := args[0].(int)
		if prev == c.Index {
			c.DeleteStatus(BurstKey)
			if c.applyC4 {
				c.applyC4 = false
				c.c4()
			}
		}
		return false
	}, "raiden-burst-clear")
}

func (c *char) onBurstStackCount() {
	// TODO: 以前はPostBurstで処理していた。現在も正しく動作するか要確認
	c.Core.Events.Subscribe(event.OnEnergyBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() == c.Index {
			return false
		}
		char := args[0].(*character.CharWrapper)
		amount := args[2].(float64)
		// キャラクターの最大エネルギーに基づきスタックを加算
		stacks := resolveStackGain[c.TalentLvlBurst()] * amount
		if c.Base.Cons > 0 {
			if char.Base.Element == attributes.Electro {
				stacks *= 1.8
			} else {
				stacks *= 1.2
			}
		}
		previous := c.stacks
		c.stacks += stacks
		if c.stacks > 60 {
			c.stacks = 60
		}
		c.Core.Log.NewEvent("resolve stacks gained", glog.LogCharacterEvent, c.Index).
			Write("previous", previous).
			Write("amount", stacks).
			Write("final", c.stacks)
		return false
	}, "raiden-stacks")
}
