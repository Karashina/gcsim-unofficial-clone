package sigewinne

import (
	"fmt"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var endLag []int

const (
	earliestCancel = 122
	chargeBurstDur = 99
)

func init() {
	endLag = frames.InitAbilSlice(271 - 241) // Burst end -> Skill
	endLag[action.ActionAttack] = 269 - 241
	endLag[action.ActionSwap] = 245 - 241
	endLag[action.ActionWalk] = 258 - 241
	endLag[action.ActionDash] = 0
	endLag[action.ActionJump] = 0
}

func (c *char) burstFindDroplets() {
	droplets := c.getSourcewaterDroplets()

	// TODO: 「水滴チェック」前に水滴がタイムアウトした場合はカウントされない
	indices := c.Core.Combat.Rand.Perm(len(droplets))
	orbs := 0
	for _, ind := range indices {
		g := droplets[ind]
		c.consumeDroplet(g)
		orbs += 1
	}
	c.Core.Combat.Log.NewEvent(fmt.Sprint("Picked up ", orbs, " droplets"), glog.LogCharacterEvent, c.Index)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.burstEarlyCancelled {
		return action.Info{}, fmt.Errorf("%v: Cannot early cancel Super Saturated Syringing with Elemental Burst", c.Base.Key)
	}

	ticks, ok := p["ticks"]
	if !ok {
		ticks = -1
	} else {
		ticks = max(ticks, 2)
	}

	c.Core.Player.SwapCD = math.MaxInt16
	c.tickAnimLength = getBurstHitmark(1)

	c.Core.Tasks.Add(func() {
		// TODO: 正確なタイミング？
		c.burstFindDroplets()

		c.burstStartF = c.Core.F
		c.Core.Tasks.Add(c.burstTick(c.burstStartF, 1, ticks, false), getBurstHitmark(1))
	}, chargeBurstDur)

	c.SetCDWithDelay(action.ActionBurst, 18*60, 1)
	c.ConsumeEnergy(5)

	c.addC2Shield()

	return action.Info{
		Frames: func(next action.Action) int {
			return chargeBurstDur + c.tickAnimLength + endLag[next]
		},
		AnimationLength: chargeBurstDur + c.burstMaxDuration,
		CanQueueAfter:   earliestCancel,
		State:           action.BurstState,
		OnRemoved: func(next action.AnimationState) {
			// 早期キャンセル時の正しい交代CDを計算する必要がある
			switch next {
			case action.DashState, action.JumpState:
				c.Core.Player.SwapCD = max(player.SwapCDFrames-(c.Core.F-c.lastSwap), 0)
			}
			c.removeC2Shield()
		},
	}, nil
}

func (c *char) burstTick(src, tick, maxTick int, last bool) func() {
	return func() {
		if c.burstStartF != src {
			return
		}
		// 元素爆発アニメーション外 → ティックなし
		if c.Core.F > c.burstStartF+c.burstMaxDuration {
			return
		}

		if last {
			c.Core.Player.SwapCD = endLag[action.ActionSwap]
			return
		}

		// tickパラメータ指定済みで上限到達 → ウェーブ発動、早期キャンセルフラグを有効化、ティックキュー停止
		if tick == maxTick {
			c.burstWave()
			c.burstEarlyCancelled = true
			c.Core.Player.SwapCD = endLag[action.ActionSwap]
			return
		}

		c.burstWave()

		// 次のTick処理
		if maxTick == -1 || tick < maxTick {
			tickDelay := getBurstHitmark(tick + 1)
			// 次のティックまでの新しいアニメーション長を計算
			nextTickAnimLength := c.Core.F - c.burstStartF + tickDelay

			c.Core.Tasks.Add(c.burstTick(src, tick+1, maxTick, false), tickDelay)

			// 次のティックが元素爆発持続時間終了後になる場合、最終ティックをキューに追加
			if nextTickAnimLength > c.burstMaxDuration {
				// 元素爆発持続時間終了時に最終ティックをキューに追加
				c.Core.Tasks.Add(c.burstTick(src, tick+1, maxTick, true), c.burstMaxDuration-c.tickAnimLength)
				// 最終的にtickAnimLengthを元素爆発全体の持続時間と等しく更新
				c.tickAnimLength = c.burstMaxDuration
			} else {
				// 次のティックが持続時間内 → 通常通りtickAnimLengthを更新
				c.tickAnimLength = nextTickAnimLength
			}
		}
	}
}

func (c *char) burstWave() {
	// TODO: 実際のヒットボックス？
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 4, 10)

	// TODO: 設置型？
	ai := combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         "Super Saturated Syringing",
		AttackTag:    attacks.AttackTagElementalBurst,
		ICDTag:       attacks.ICDTagElementalBurst,
		ICDGroup:     attacks.ICDGroupSigewinneBurst,
		StrikeType:   attacks.StrikeTypeDefault,
		Element:      attributes.Hydro,
		Durability:   25,
		FlatDmg:      burstDMG[c.TalentLvlAttack()] * c.MaxHP(),
		HitlagFactor: 0.01,
	}
	c.Core.QueueAttack(ai, ap, 0, 0, c.c2CB)
}

func getBurstHitmark(tick int) int {
	switch tick {
	case 1:
		return 0
	default:
		return 25
	}
}
