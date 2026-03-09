package heizou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 鹿野院平蔵がフィールド上で拡散反応を発動した時、
// 勉ぎの心の譲り重ねスタックを1層獲得する。
// この効果は0.1秒に1回発動可能。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	const a1IcdKey = "heizou-a1-icd"
	swirlCB := func() func(args ...interface{}) bool {
		return func(args ...interface{}) bool {
			if c.StatusIsActive(a1IcdKey) {
				return false
			}
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != c.Index {
				return false
			}
			if c.Core.Player.Active() != c.Index {
				return false
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagSwirlPyro:
			case attacks.AttackTagSwirlHydro:
			case attacks.AttackTagSwirlElectro:
			case attacks.AttackTagSwirlCryo:
			default:
				return false
			}
			// スタックが上限であってもICDは発動する
			c.AddStatus(a1IcdKey, 6, true)
			c.addDecStack()
			return false
		}
	}

	c.Core.Events.Subscribe(event.OnEnemyDamage, swirlCB(), "heizou-a1")
}

// 鹿野院平蔵の勉ぎの心が敵に命中した後、
// パーティー全員（平蔵を除く）の元素熔煙が10秒間80アップする。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	dur := 60 * 10
	for i, char := range c.Core.Player.Chars() {
		if i == c.Index {
			continue // 平蔵自身には適用しない
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("heizou-a4", dur),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return c.a4Buff, true
			},
		})
	}
	c.Core.Log.NewEvent("heizou a4 triggered", glog.LogCharacterEvent, c.Index).Write("em snapshot", c.a4Buff[attributes.EM]).Write("expiry", c.Core.F+dur)
}
