package itto

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstDuration  = 660 + 90 + 45 // 基本コンボをカバーする程度
	burstBuffKey   = "itto-q"
	burstAtkSpdKey = "itto-q-atkspd"
)

func init() {
	burstFrames = frames.InitAbilSlice(91) // Q -> N1/CA0/CA1/CAF/E
	burstFrames[action.ActionDash] = 84    // Q -> D
	burstFrames[action.ActionJump] = 85    // Q -> J
	burstFrames[action.ActionSwap] = 90    // Q -> Swap
}

// Noelleを参考に実装
// 元素爆発:
// 荒瀧一斗が鬼王の覇気を解き放ち、鬼王の金砕棒を振るって戦う。
// この状態には以下の特殊効果がある:
// - 一斗の通常攻撃、重撃、落下攻撃を岩元素ダメージに変換する。この変換は上書きされない。
// - 一斗の通常攻撃速度が増加する。また、防御力に基づいて攻撃力が増加する。
// - 命中時、攻撃コンボの1段目と3段目で荒瀧一斗が怒髪衝天スタックを1つ獲得する。
// - 一斗の元素耐性と物理耐性が20%減少する。
// 鬼王状態は一斗がフィールドを離れると解除される。
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// N1プレスタック技法。一斗がN1 -> Qを行った場合、Q防御→攻撃変換前にスタックを追加
	// https://library.keqingmains.com/evidence/characters/geo/itto#itto-n1-burst-cancel-ss-stack
	if p["prestack"] != 0 && c.Core.Player.CurrentState() == action.NormalAttackState && c.NormalCounter == 1 {
		c.addStrStack("n1-burst-cancel", 1)
	}

	// デバッグで適用されたModの一覧を表示するための「仮の」スナップショットを生成
	aiSnapshot := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Royal Descent: Behold, Itto the Evil! (Stat Snapshot)",
	}
	c.Snapshot(&aiSnapshot)
	burstDefSnapshot := c.TotalDef(true)
	mult := defconv[c.TalentLvlBurst()]

	// TODO: バフの正確なタイミングを確認
	// Q 防御力→攻撃力変換
	mATK := make([]float64, attributes.EndStatType)
	mATK[attributes.ATK] = mult * burstDefSnapshot
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(burstBuffKey, burstDuration),
		AffectedStat: attributes.ATK,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return mATK, true
		},
	})

	// Q 攻撃速度バフ
	mAtkSpd := make([]float64, attributes.EndStatType)
	mAtkSpd[attributes.AtkSpd] = 0.10
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(burstAtkSpdKey, burstDuration),
		AffectedStat: attributes.AtkSpd,
		Amount: func() ([]float64, bool) {
			if c.Core.Player.CurrentState() != action.NormalAttackState {
				return nil, false
			}
			return mAtkSpd, true
		},
	})

	c.Core.Log.NewEvent("itto burst", glog.LogSnapshotEvent, c.Index).
		Write("total def", burstDefSnapshot).
		Write("atk added", mATK[attributes.ATK]).
		Write("mult", mult)

	if c.Base.Cons >= 1 {
		// TODO: itto-c1-mechanics TCLエントリへのリンクを後で追加
		// Qのアニメーション完了前のため、キャラキューは不要
		// 75は一斗がC1から2スタックを得るタイミングの概算
		c.Core.Tasks.Add(c.c1, 75)
	}

	if c.Base.Cons >= 2 {
		// TODO: C2の遅延を確認、ただし影響は小さい
		// CD/エネルギー遅延後に適用すべき
		// Qのアニメーション完了前のため、キャラキューは不要
		c.Core.Tasks.Add(c.c2, 9)
	}

	// 元素爆発終了時に適用
	c.burstCastF = c.Core.F
	if c.Base.Cons >= 4 {
		c.applyC4 = true
		src := c.burstCastF
		c.QueueCharTask(func() {
			if src == c.burstCastF {
				c.c4()
			}
		}, burstDuration)
	}

	c.SetCD(action.ActionBurst, 1080) // 18s * 60
	c.ConsumeEnergy(1)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		if prev == c.Index && c.StatModIsActive(burstBuffKey) {
			c.DeleteStatMod(burstBuffKey)
			c.DeleteStatMod(burstAtkSpdKey)
			c.c4()
		}
		return false
	}, "itto-exit")
}
