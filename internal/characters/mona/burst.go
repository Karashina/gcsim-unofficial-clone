package mona

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const burstHitmark = 107

func init() {
	burstFrames = frames.InitAbilSlice(127) // Q -> Swap
	burstFrames[action.ActionAttack] = 121  // Q -> N1
	burstFrames[action.ActionCharge] = 118  // Q -> CA
	burstFrames[action.ActionSkill] = 115   // Q -> E
	burstFrames[action.ActionDash] = 115    // Q -> D
	burstFrames[action.ActionJump] = 104    // Q -> J
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 泡は0ダメージの水元素付着を行う
	// 泡ステータスを付与し、泡ステータスが消えた次のフレームで星異ダメージを発動
	// 泡ステータスは次のいずれかで破裂 → 凍結なしでダメージを受ける、または凍結が解除される

	// 1.7秒後に最初のダメージなし攻撃を適用
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Illusory Bubble (Initial)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       0,
	}
	cb := func(a combat.AttackCB) {
		t, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		// 泡は各ターゲットに個別に適用される
		// 通常破裂しない場合8秒間持続
		t.AddStatus(bubbleKey, 481, true) // 破裂処理の問題を避けるため1フレーム余分に追加
		c.Core.Log.NewEvent("mona bubble on target", glog.LogCharacterEvent, c.Index).
			Write("char", c.Index)
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), -1, burstHitmark, cb)

	// 8秒後にまだ泡が割れていなければ0ダメージ攻撃をキューに追加して泡を割る
	aiBreak := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Illusory Bubble (Break)",
		AttackTag:  attacks.AttackTagMonaBubbleBreak,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Physical,
		Durability: 0,
		Mult:       0,
	}
	c.Core.QueueAttack(aiBreak, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), -1, burstHitmark+480)

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionJump], // 最早キャンセルは元素爆発ヒットマーク前
		State:           action.BurstState,
	}, nil
}

func (c *char) burstDamageBonus() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = dmgBonus[c.TalentLvlBurst()]
	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("mona-omen", -1),
			Amount: func(_ *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				x, ok := t.(*enemy.Enemy)
				if !ok {
					return nil, false
				}
				// 泡または星異のどちらかが存在する場合のみ有効
				if x.StatusIsActive(bubbleKey) || x.StatusIsActive(omenKey) {
					return m, true
				}
				return nil, false
			},
		})
	}
}

// 泡は凍結されていない状態で攻撃を受けるか、攻撃が凍結を解除したときに破裂する
// すなわち impulse > 0
func (c *char) burstHook() {
	// OnDamage にフックする。常にアクティブのまま維持
	// 凍結が攻撃をトリガーするため、これで問題ない
	// TODO: この実装では現在、最初の感電Tickで泡がすぐに割れてしまう。
	// 参照: https://docs.google.com/document/d/1pXlgCaYEpoizMIP9-QKlSkQbmRicWfrEoxb9USWD1Ro/edit#
	// 2回目の感電Tickでのみ割れるべき
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// ターゲットにデバフがなければ無視
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if !t.StatusIsActive(bubbleKey) {
			return false
		}
		// 時間切れの場合は常に破裂
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagMonaBubbleBreak {
			c.triggerBubbleBurst(t)
			return false
		}
		// impulseがなければ割らない
		if atk.Info.NoImpulse {
			return false
		}
		// それ以外はダメージで破裂
		c.triggerBubbleBurst(t)

		return false
	}, "mona-bubble-check")
}

func (c *char) triggerBubbleBurst(t *enemy.Enemy) {
	// 泡タグを削除
	t.DeleteStatus(bubbleKey)
	// 星異デバフを付与
	dur := int(omenDuration[c.TalentLvlBurst()] * 60)
	t.AddStatus(omenKey, dur, true)
	// ダメージを発動
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Illusory Bubble (Explosion)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 50,
		Mult:       explosion[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(ai, combat.NewSingleTargetHit(t.Key()), 1, 1)
}
