package lyney

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1ICDKey = "lyney-c1-icd"
	c1ICD    = 15 * 60
)

// リネは同時に2つのグリンマルキンハットを存在させることができる。
// さらにマジック弾は2つのグリンマルキンハットを召喚し、リネにマジック余剰スタックを1つ追加で付与する。
// この効果は15秒に1回発動可能。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.maxHatCount = 2
}

// 第1命ノ星座のマジック余剰スタックはHP消費に関係なく付与される
func (c *char) addC1PropStack() func() {
	return func() {
		if c.Base.Cons < 1 || c.StatusIsActive(c1ICDKey) {
			return
		}
		c.increasePropSurplusStacks()
	}
}

func (c *char) c1HatIncrease() int {
	addCount := 0
	if c.Base.Cons >= 1 && !c.StatusIsActive(c1ICDKey) {
		addCount = 1
		c.AddStatus(c1ICDKey, c1ICD, true)
	}
	return addCount
}

// TODO: 正確なフレーム数?
const c2Interval = 2 * 60

// リネがフィールド上にいる間、2秒ごとに鮮明のスタックを1つ獲得する。
// 会心ダメージが20%増加する。最大3スタック。
// リネがフィールドを離れるとこの効果は解除される。
func (c *char) c2Setup() {
	if c.Base.Cons < 2 {
		return
	}

	// キャラ交代時に第2命ノ星座のクリア/適用をリッスン
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		// リネから交代する場合は第2命ノ星座をクリア
		prev := args[0].(int)
		if prev == c.Index {
			c.c2Src = -1
			c.c2Stacks = 0
			return false
		}
		// リネに交代する場合は第2命ノ星座を適用
		next := args[1].(int)
		if next == c.Index {
			c.c2Src = c.Core.F
			c.QueueCharTask(c.c2StackCheck(c.Core.F), c2Interval)
			c.Core.Log.NewEvent("Lyney C2 started", glog.LogCharacterEvent, c.Index).Write("c2_stacks", c.c2Stacks)
			return false
		}
		return false
	}, "lyney-c2-swap")

	// バフを追加
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("lyney-c2", -1),
		AffectedStat: attributes.CD,
		Amount: func() ([]float64, bool) {
			m[attributes.CD] = float64(c.c2Stacks) * 0.2
			return m, true
		},
	})
}

func (c *char) c2StackCheck(src int) func() {
	return func() {
		// 交代によりソースが変更された場合はスタックを追加しない
		if src != c.c2Src {
			return
		}
		// フィールド上にいない場合はスタックを追加しない
		// 安全チェック（前のチェック+イベント購読により保証されるはず）
		if c.Index != c.Core.Player.Active() {
			return
		}
		// 既に最大スタックの場合は追加しない
		// 交代以外でスタックを失う方法はないため、最大時はスタックチェックのキューは不要
		if c.c2Stacks == 3 {
			return
		}
		// スタックを追加
		c.c2Stacks++
		c.Core.Log.NewEvent("Lyney C2 stack added", glog.LogCharacterEvent, c.Index).Write("c2_stacks", c.c2Stacks)
		// スタックチェックをキューに追加
		c.QueueCharTask(c.c2StackCheck(src), c2Interval)
	}
}

// リネの炎元素重撃が敵に命中した後、その敵の炎元素耐性が6秒間20%減少する。
func (c *char) makeC4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("lyney-c4", 6*60),
			Ele:   attributes.Pyro,
			Value: -0.20,
		})
	}
}

// リネがマジック弾を発射する際、パイロテクニックストライク・再演を発射し、パイロテクニックストライクのダメージの80%を与える。
// このダメージは重撃ダメージとみなされる。
func (c *char) c6(c6Travel int) {
	if c.Base.Cons < 6 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Pyrotechnic Strike: Reprised",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagLyneyEndBoom,
		ICDGroup:   attacks.ICDGroupLyneyExtra,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       propPyrotechnic[c.TalentLvlAttack()] * 0.8,
	}
	// TODO: スナップショット
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			1,
		),
		0,
		c6Travel,
		c.makeC4CB(),
	)
}
