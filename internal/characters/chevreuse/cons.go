package chevreuse

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1ICDKey    = "chev-c1-icd"
	c2ICDKey    = "chev-c2-icd"
	c4StatusKey = "chev-c4"
)

// 「共同戦術」状態のアクティブキャラクター（シュヴルーズ自身を除く）が
// 過負荷反応を起こすと、元素エネルギー6を回復する。
// この効果は10秒に1回発動可能。
// 固有天賦「先峰隊の協同戦術」の解放が必要。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	if !c.onlyPyroElectro {
		return
	}

	c.Core.Events.Subscribe(event.OnOverload, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		// シュヴルーズ自身は含まない
		if atk.Info.ActorIndex == c.Index {
			return false
		}
		// フィールド外ではトリガーしない
		if atk.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		// CD中はトリガーしない
		if c.StatusIsActive(c1ICDKey) {
			return false
		}
		c.AddStatus(c1ICDKey, 10*60, true)

		active := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		active.AddEnergy("chev-c1", 6)

		return false
	}, "chev-c1")
}

// 近距離キャノン急射の長押しでターゲットに命中した後、
// 命中位置付近で2回の連鎖爆発がトリガーされる。
// 各爆発はシュヴルーズの攻撃力120%分の炎元素ダメージを与える。
// この効果は10秒に1回までトリガー可能で、
// ダメージは元素スキルダメージとみなされる。
func (c *char) c2() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}
	// 敵だけでなく何かに命中した時にトリガー
	return func(a combat.AttackCB) {
		if c.StatusIsActive(c2ICDKey) {
			return
		}
		c.AddStatus(c2ICDKey, 10*60, true)

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Sniper Induced Explosion (C2)",
			AttackTag:  attacks.AttackTagElementalArt,
			// ElementalArtExtra であるべきだが、他の Chevreuse の攻撃はこのタグを共有していないので問題ない
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   25,
			Element:    attributes.Pyro,
			Durability: 25,
			Mult:       1.2,
		}

		// 位置を計算
		player := c.Core.Combat.Player()
		targetPos := a.Target.Pos()
		// TODO: ここでプレイヤーの方向を使用しているのは不正確。被弾方向を使うべきかもしれない
		bomb1Pos := geometry.CalcOffsetPoint(targetPos, geometry.Point{X: -1.5}, player.Direction())
		bomb2Pos := geometry.CalcOffsetPoint(targetPos, geometry.Point{X: 1.5}, player.Direction())

		// 遅延を計算
		// 命中から0.6秒～1秒のランダム
		// 爆弾間で共有されない
		bomb1Delay := int(60 * (0.6 + c.Core.Rand.Float64()*(1-0.6)))
		bomb2Delay := int(60 * (0.6 + c.Core.Rand.Float64()*(1-0.6)))

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(bomb1Pos, nil, 3),
			bomb1Delay,
			bomb1Delay,
		)

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(bomb2Pos, nil, 3),
			bomb2Delay,
			bomb2Delay,
		)
	}
}

// 手榴弾一斉発射使用後、近距離キャノン急射の長押しは
// CDに入らなくなる。
// この効果は長押しで2回発射するか、
// 6秒経過後に解除される。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	c.AddStatus(c4StatusKey, 6*60, true)
	c.c4ShotsLeft = 2
}

// 近距離キャノン急射の回復効果から12秒後、
// 付近の全パーティメンバーはシュヴルーズのHP上限の10%のHPを回復する。
func (c *char) c6TeamHeal() {
	if c.Base.Cons < 6 {
		return
	}
	c.c6HealQueued = false

	for _, char := range c.Core.Player.Chars() {
		c.c6(char)
	}

	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  -1,
		Message: "In Pursuit of Ending Evil (C6)",
		Src:     0.1 * c.MaxHP(),
		Bonus:   c.Stat(attributes.Heal),
	})
}

func (c *char) c6(char *character.CharWrapper) {
	if c.Base.Cons < 6 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.PyroP] = 0.20
	m[attributes.ElectroP] = 0.20

	char.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag(fmt.Sprintf("chev-c6-%v-stack", c.c6StackCounts[char.Index]+1), 8*60),
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
	c.c6StackCounts[char.Index] = (c.c6StackCounts[char.Index] + 1) % 3
}
