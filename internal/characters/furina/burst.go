package furina

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstHitmark            = 98
	burstDur                = 18.2 * 60
	burstKey                = "furina-burst"
	fanfareDebounceKey      = "furina-fanfare-debounce"
	fanfareDrainToGainDelay = 6                                // 推定値
	fanfareDebounceDur      = 30 + fanfareDrainToGainDelay - 1 // 推定値
)

func init() {
	burstFrames = frames.InitAbilSlice(121)
	burstFrames[action.ActionAttack] = 113 // Q -> N1
	burstFrames[action.ActionCharge] = 113 // Q -> CA
	burstFrames[action.ActionSkill] = 114  // Q -> E
	burstFrames[action.ActionDash] = 115   // Q -> D
	burstFrames[action.ActionJump] = 115   // Q -> J
	burstFrames[action.ActionSwap] = 111   // Q -> Swap
}

func (c *char) addFanfareFunc(amt float64) func() {
	return func() {
		if c.Base.Cons >= 2 {
			amt *= 3.5
		}
		prevFanfare := c.curFanfare
		c.curFanfare = min(c.maxC2Fanfare, c.curFanfare+amt)
		c.Core.Log.NewEvent("Gained Fanfare", glog.LogCharacterEvent, c.Index).
			Write("previous fanfare", prevFanfare).
			Write("current fanfare", c.curFanfare)
	}
}

func (c *char) queueFanfareGain(amt float64) {
	// デバウンスステータスの有無に基づいてHP消費とファンファーレ変更の遅延を決定
	var delay int
	if !c.StatusIsActive(fanfareDebounceKey) {
		// デバウンスステータスが追加されるまで、同フレームの他キャラのHP消費がファンファーレ獲得をキューできるよ1フレームの猶予を残す
		// 一度に1つのデバウンスステータス追加タスクのみキューされるようにcharのboolを使用
		if !c.fanfareDebounceTaskQueued {
			c.fanfareDebounceTaskQueued = true
			c.Core.Tasks.Add(func() {
				c.AddStatus(fanfareDebounceKey, fanfareDebounceDur, false) // TODO: ヒットラグについて不明
				c.fanfareDebounceTaskQueued = false
			}, 1)
		}
		// HP消費からのファンファーレ獲得はサーバーで消費が確認された後でも遅延がある
		delay = fanfareDrainToGainDelay
	} else {
		// HP消費からのファンファーレ獲得はデバウンスステータスが終了するまで遅延する必要がある
		delay = c.StatusDuration(fanfareDebounceKey)
	}
	// ファンファーレ変更をキュー
	c.Core.Tasks.Add(c.addFanfareFunc(amt), delay)
}

func (c *char) burstInit() {
	c.maxC2Fanfare = 300
	c.maxQFanfare = 300
	if c.Base.Cons >= 1 {
		c.maxQFanfare = 400
		c.maxC2Fanfare = 400
	}
	if c.Base.Cons >= 2 {
		// 400 + 140/0.35 = 800
		c.maxC2Fanfare = 800
	}
	c.burstBuff = make([]float64, attributes.EndStatType)

	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		if !c.StatusIsActive(burstKey) {
			return false
		}

		di := args[0].(*info.DrainInfo)

		if di.Amount <= 0 {
			return false
		}

		char := c.Core.Player.ByIndex(di.ActorIndex)
		amt := di.Amount / char.MaxHP() * 100
		c.queueFanfareGain(amt)

		return false
	}, "furina-fanfare-on-hp-drain")

	c.Core.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		if !c.StatusIsActive(burstKey) {
			return false
		}

		target := args[1].(int)
		amount := args[2].(float64)
		overheal := args[3].(float64)

		if amount <= 0 {
			return false
		}

		if math.Abs(amount-overheal) <= 1e-9 {
			return false
		}

		char := c.Core.Player.ByIndex(target)
		amt := (amount - overheal) / char.MaxHP() * 100

		c.queueFanfareGain(amt)

		return false
	}, "furina-fanfare-on-heal")

	burstDMGRatio := burstFanfareDMGRatio[c.TalentLvlBurst()]
	burstHealRatio := burstFanfareHBRatio[c.TalentLvlBurst()]
	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("furina-burst-damage-buff", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if !c.StatusIsActive(burstKey) {
					return nil, false
				}
				c.burstBuff[attributes.DmgP] = min(c.curFanfare, c.maxQFanfare) * burstDMGRatio
				return c.burstBuff, true
			},
		})

		char.AddHealBonusMod(character.HealBonusMod{
			Base: modifier.NewBase("furina-burst-heal-buff", -1),
			Amount: func() (float64, bool) {
				if c.StatusIsActive(burstKey) {
					return min(c.curFanfare, c.maxQFanfare) * burstHealRatio, false
				}
				return 0, false
			},
		})
	}
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Let the People Rejoice",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    c.MaxHP() * burstDMG[c.TalentLvlBurst()],
	}

	c.curFanfare = 0
	c.DeleteStatus(burstKey)

	c.QueueCharTask(func() {
		if c.Base.Cons >= 1 {
			c.curFanfare = 150
		}
		c.AddStatus(burstKey, burstDur, true)
	}, 95) // これはヒットマーク前なので、フリーナの元素爆発ダメージは1凸の恩恵を受ける。テスト済み・確認済み

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), burstHitmark, burstHitmark)

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
