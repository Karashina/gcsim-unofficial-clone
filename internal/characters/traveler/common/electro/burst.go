package electro

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var burstFrames [][]int

const burstHitmark = 37

func init() {
	burstFrames = make([][]int, 2)

	// 男性
	burstFrames[0] = frames.InitAbilSlice(63) // Q -> E
	burstFrames[0][action.ActionAttack] = 62  // Q -> N1
	burstFrames[0][action.ActionDash] = 62    // Q -> D
	burstFrames[0][action.ActionJump] = 61    // Q -> J
	burstFrames[0][action.ActionSwap] = 60    // Q -> Swap

	// 女性
	burstFrames[1] = frames.InitAbilSlice(62) // Q -> E/D
	burstFrames[1][action.ActionAttack] = 61  // Q -> N1
	burstFrames[1][action.ActionJump] = 61    // Q -> J
	burstFrames[1][action.ActionSwap] = 61    // Q -> Swap
}

/*
*
[12:01 PM] pai: 計測したことはないけど、雷旅人の元素爆発のレンジはだいたい深境螺旋の1～1.5タイル分に見える。元素スキルはもう少し遠くに届くと思う
[12:01 PM] pai: 元素スキルの3ヒットも分散してある程度オートターゲットする。役に立つ情報かもしれない
*
*/
func (c *Traveler) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Bellowing Thunder",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   150,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 0, burstHitmark)

	c.SetCDWithDelay(action.ActionBurst, 1200, 35)
	c.ConsumeEnergy(37)

	// 雷旅人の元素爆発はヒットラグで延長されない
	c.Core.Status.Add("travelerelectroburst", 720) // 12秒、キャスト時に開始

	procAI := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Falling Thunder Proc (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burstTick[c.TalentLvlBurst()],
	}
	c.burstSnap = c.Snapshot(&procAI)
	c.burstAtk = &combat.AttackEvent{
		Info:     procAI,
		Snapshot: c.burstSnap,
	}
	c.burstSrc = c.Core.F

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames[c.gender]),
		AnimationLength: burstFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   burstFrames[c.gender][action.ActionJump], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *Traveler) burstProc() {
	icd := 0

	// 雷の衣
	//  アクティブキャラクターの通常攻撃または重撃が敵に命中すると、落雷を発生させ雷元素ダメージを与える。
	//  落雷が敵に命中すると、そのキャラクターのエネルギーが回復する。
	//  落雷は0.5秒ごとに1回のみ発生可能。
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		t := args[0].(combat.Target)

		// 通常攻撃/重撃のみ適用
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// 攻撃をトリガーしたキャラクターがまだフィールドにいることを確認
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		// 元素爆発がアクティブな場合のみ適用
		if c.Core.Status.Duration("travelerelectroburst") == 0 {
			return false
		}
		// 落雷は0.5秒ごとに1回のみ発生可能。
		if icd > c.Core.F {
			c.Core.Log.NewEvent("travelerelectro Q (active) on icd", glog.LogCharacterEvent, c.Index)
			return false
		}

		// 元素爆発のスナップショットを使用、ターゲットとソースフレームを更新
		atk := *c.burstAtk
		atk.SourceFrame = c.Core.F
		radius := 2.0
		if c.Base.Cons >= 6 && c.burstC6WillGiveEnergy {
			radius = 2.5
		}
		atk.Pattern = combat.NewCircleHitOnTarget(t, nil, radius)

		// 2凸 - Violet Vehemence
		// 轟く雷で生じた落雷が敵に命中すると、8秒間雷元素耐性を15%減少させる。
		// 6凸 - World-Shaker
		//  轟く雷でトリガーされた落雷2回ごとにダメージが大幅に増加し、
		//  次の落雷は元のダメージの200%を与え、さらに
		//  現在のキャラクターに追加で1エネルギーを回復する。
		c.c6Damage(&atk)
		for _, cb := range []combat.AttackCBFunc{c.fallingThunderEnergy(), c.c2(), c.c6Energy()} {
			if cb != nil {
				atk.Callbacks = append(atk.Callbacks, cb)
			}
		}

		c.Core.QueueAttackEvent(&atk, 1)

		c.Core.Log.NewEvent("travelerelectro Q proc'd", glog.LogCharacterEvent, c.Index).
			Write("char", ae.Info.ActorIndex).
			Write("attack tag", ae.Info.AttackTag)

		icd = c.Core.F + 30 // 0.5s
		return false
	}, "travelerelectro-bellowingthunder")
}

func (c *Traveler) fallingThunderEnergy() combat.AttackCBFunc {
	return func(_ combat.AttackCB) {
		// アクティブキャラクターに固定エネルギーを回復
		active := c.Core.Player.ActiveChar()
		active.AddEnergy("travelerelectro-fallingthunder", burstRegen[c.TalentLvlBurst()])
	}
}
