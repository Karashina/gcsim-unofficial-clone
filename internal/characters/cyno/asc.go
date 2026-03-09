package cyno

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a1Key = "cyno-a1"

// セノが「圣儀・狼駆」で発動した「契約の导砂者」状態の際、
// 一定間隔で「末途見」の構えに入る。この構え中に「秘儀・裂置の将」を発動すると、
// 「裁定」効果が発動し、その「秘儀・裂置の将」のダメージが35%増加し、
// セノの攻撃力の50%の雷元素ダメージを与える「砂の矢」を3本発射する。
// 「砂の矢」のダメージは元素スキルダメージとみなされる。
//
// - 突破レベルチェックは burst.go で行い、突破レベルチェックが失敗するだけのキューを避ける
//
// - 固有天賦1の他の部分はこのタスクが適用するステータスに依存するため、追加の突破チェックは不要
func (c *char) a1() {
	if !c.StatusIsActive(burstKey) {
		return
	}
	c.a1Extended = false
	c.AddStatus(a1Key, 84, true)
	c.QueueCharTask(c.a1, 234)
}

func (c *char) a1Buff() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.35
	// ゲームでも1秒の modifier でダメージバフを実装している
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("cyno-a1-dmg", 60),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			// 実際のゲームでは固有天賦1に AttackTagElementalArtExtra を使用、これは適切な
			// 回避策
			if atk.Info.Abil != skillBName {
				return nil, false
			}
			return m, true
		},
	})
}

// 固有天賦1の modifier がアクティブの状態でダッシュすると、
// modifier の耐久が20増加する。0.28秒の延長に相当。
func (c *char) a1Extension() {
	c.Core.Events.Subscribe(event.OnDash, func(_ ...interface{}) bool {
		if c.a1Extended {
			return false
		}
		active := c.Core.Player.ActiveChar()
		if !(active.Index == c.Index && active.StatusIsActive(a1Key)) {
			return false
		}
		c.ExtendStatus(a1Key, 17)
		c.a1Extended = true
		c.Core.Log.NewEvent("a1 dash pp slide", glog.LogCharacterEvent, c.Index).
			Write("expiry", c.StatusExpiry(a1Key))
		return false
	}, "cyno-a1-dash")
}

// セノのダメージ値は元素熔化に基づき以下の通り増加する:
//
// - 「契約の导砂者」の通常攻撃ダメージが元素熔化の150%分増加。
func (c *char) a4NormalAttack() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return c.Stat(attributes.EM) * 1.5
}

// セノのダメージ値は元素熔化に基づき以下の通り増加する:
//
// - 固有天賦「羽落ちの裁定」の「砂の矢」ダメージが元素熔化の250%分増加。
func (c *char) a4Bolt() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return c.Stat(attributes.EM) * 2.5
}
