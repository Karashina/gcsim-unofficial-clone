package eula

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var burstFrames []int

const burstHitmark = 100
const lightfallHitmark = 35

func init() {
	burstFrames = frames.InitAbilSlice(123) // Q -> E
	burstFrames[action.ActionAttack] = 120  // Q -> N1
	burstFrames[action.ActionDash] = 122    // Q -> D
	burstFrames[action.ActionJump] = 121    // Q -> J
	burstFrames[action.ActionWalk] = 117    // Q -> Walk
	burstFrames[action.ActionSwap] = 120    // Q -> Swap
}

const (
	burstKey         = "eula-q"
	burstStackICDKey = "eula-q-stack-icd"
)

// 元素爆発 365～415フレーム、60fps = 120
// 元素爆発のチャージ時間は約8秒
func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.burstCounter = 0
	if c.Base.Cons >= 6 {
		c.burstCounter = 5
	}

	// 初撃ダメージを追加
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Glacial Illumination",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   0,
		Element:    attributes.Cryo,
		Durability: 50,
		Mult:       burstInitial[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8),
		burstHitmark,
		burstHitmark,
	)
	c.a4()

	// ユーラ元素爆発ステータス開始処理
	// 光落の剣は発動から約9.5秒後に点灯
	// 設置物：ヒットラグの影響を受けない
	c.Core.Tasks.Add(func() {
		c.Core.Status.Add(burstKey, 600-lightfallHitmark-burstFrames[action.ActionWalk]+1)
		c.Core.Log.NewEvent("eula burst started", glog.LogCharacterEvent, c.Index).
			Write("stacks", c.burstCounter).
			Write("expiry", c.Core.F+600-lightfallHitmark-burstFrames[action.ActionWalk]+1)
	}, burstFrames[action.ActionWalk]) // 最も早いタイミングで元素爆発ステータスを開始

	// ユーラ元素爆発ダメージ処理
	// 光落の剣のヒットマークは発動から600フレーム
	c.Core.Tasks.Add(func() {
		// フィールド退場による早期爆発がまだ発生していないことを確認
		if c.Core.Status.Duration(burstKey) > 0 {
			c.triggerBurst()
		}
	}, 600-lightfallHitmark) // 元素爆発ステータスが通常期限切れする直前にダメージをトリガーできるか確認

	// エネルギーはアニメーション後まで消費されない
	c.ConsumeEnergy(107)
	c.SetCDWithDelay(action.ActionBurst, 20*60, 97)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionWalk], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) triggerBurst() {
	if c.burstCounter > 30 {
		c.burstCounter = 30
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Glacial Illumination (Lightfall)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   400,
		Element:    attributes.Physical,
		Durability: 50,
		Mult:       burstExplodeBase[c.TalentLvlBurst()] + burstExplodeStack[c.TalentLvlBurst()]*float64(c.burstCounter),
	}

	c.Core.Log.NewEvent("eula burst triggering", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.burstCounter).
		Write("mult", ai.Mult)

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6.5),
		lightfallHitmark,
		lightfallHitmark,
	)
	c.Core.Status.Delete(burstKey)
	c.burstCounter = 0
}

// ユーラ自身の通常攻撃、元素スキル、元素爆発が敵にダメージを与えた時、
// 光落の剣がチャージされ、0.1秒ごとにエネルギースタックを1つ獲得できる。
func (c *char) burstStackCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.Core.Player.Active() != c.Index {
		return
	}
	if c.Core.Status.Duration(burstKey) == 0 {
		return
	}
	if a.Damage == 0 {
		return
	}
	if c.StatusIsActive(burstStackICDKey) {
		return
	}
	//TODO: ICDはガジェットタイマーに依存しているようだ。要再確認
	c.AddStatus(burstStackICDKey, 0.1*60, false)

	// カウンターに追加
	c.burstCounter++
	c.Core.Log.NewEvent("eula burst add stack", glog.LogCharacterEvent, c.Index).
		Write("stack count", c.burstCounter)
	// 6命ノ星座をチェック
	if c.Base.Cons == 6 && c.Core.Rand.Float64() < 0.5 {
		c.burstCounter++
		c.Core.Log.NewEvent("eula c6 add additional stack", glog.LogCharacterEvent, c.Index).
			Write("stack count", c.burstCounter)
	}
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.Core.Status.Duration(burstKey) > 0 {
			c.triggerBurst()
		}
		return false
	}, "eula-exit")
}
