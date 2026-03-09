package jean

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 1凸:
// 風圧剣を1秒以上長押し後の引き寄せ速度を増加し、与えるダメージを40%増加させる。
func (c *char) c1(snap *combat.Snapshot) {
	// 40%ダメージを追加
	snap.Stats[attributes.DmgP] += .4
	c.Core.Log.NewEvent("jean c1 adding 40% dmg", glog.LogCharacterEvent, c.Index).
		Write("final dmg%", snap.Stats[attributes.DmgP])
}

// 2凸:
// ジンが元素オーブ/粒子を拾うと、チーム全員の移動速度と攻撃速度が15秒間15%増加する。
func (c *char) c2() {
	c.c2buff = make([]float64, attributes.EndStatType)
	c.c2buff[attributes.AtkSpd] = 0.15
	c.Core.Events.Subscribe(event.OnParticleReceived, func(args ...interface{}) bool {
		// ジンが粒子を拾った場合のみトリガー
		if c.Core.Player.Active() != c.Index {
			return false
		}
		// 全キャラクターに2凸を適用
		for _, this := range c.Core.Player.Chars() {
			this.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("jean-c2", 900),
				AffectedStat: attributes.AtkSpd,
				Amount: func() ([]float64, bool) {
					return c.c2buff, true
				},
			})
		}
		return false
	}, "jean-c2")
}

// 4凸:
// 蒲公英の風が生成したフィールド内の全ての敵の風元素耐性あ40%減少する。
func (c *char) c4() {
	// 元素爆発開始直前に1回、その後回復ティックと同時（1秒ごと）に呼び出される
	// 全ターゲットに1.2秒のデバフを追加
	enemies := c.Core.Combat.EnemiesWithinArea(c.burstArea, nil)
	for _, e := range enemies {
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("jean-c4", 72), // 1.2s
			Ele:   attributes.Anemo,
			Value: -0.4,
		})
	}
}

// 6凸:
// 蒲公英の風が生成したフィールド内で受けるダメージが35%減少する。
// 蒲公英フィールドを離れた後、この効果は3回の攻撃または10秒間持続する。
func (c *char) c6() {
	c.Core.Log.NewEvent("jean-c6 not implemented", glog.LogCharacterEvent, c.Index)
}
