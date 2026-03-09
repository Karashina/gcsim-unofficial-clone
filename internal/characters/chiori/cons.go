package chiori

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

const (
	c2Duration        = 10 * 60
	c2SpawnInterval   = 3 * 60
	c2MinRandom       = 0.8
	c2MaxRandom       = 1.8
	c2CenterOffset    = 0.2
	c4Key             = "chiori-c4"
	c4Duration        = 8 * 60
	c4Lockout         = "chiori-c4-lockout"
	c4LockoutDuration = 15 * 60
	c4ICDKey          = "chiori-c4-icd"
	c4ICD             = 1 * 60
	c4AttackLimit     = 3
	c4MinRandom       = 1.8
	c4MaxRandom       = 2.8
	c4CenterOffset    = 0
)

// オートマトン人形「Tamoto」のAoEが50%増加。
// また、千織以外に岩元素パーティメンバーがいる場合、
// 羽ばたきの別れのダッシュ完了後に以下が発動:
// - 追加のTamotoを召喚。この方法または岩元素設置物により
// 召喚された追加Tamotoは同時に1体まで存在可能。
// - 固有天賦「仕亊げ」を発動。固有天賦「仕亊げ」の解放が必要。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	geoCount := 0
	for _, v := range c.Core.Player.Chars() {
		if v.Base.Element == attributes.Geo {
			geoCount++
		}
		if geoCount >= 2 {
			c.c1Active = true
			break
		}
	}
	// 説明の50%はAoEの体積を指すと思われる
	// 円柱の体積は pi*r^2*h なので、半径に 1.5^2=2.25 を掛ける必要がある
	c.skillSearchAoE *= 2.25
}

// 飛翼の双刃使用後10秒間、3秒ごとに簡略化オートマトン人形「絹」を
// アクティブキャラ付近に召喚。「絹」は付近の敵を攻撃し、
// Tamotoの170%に相当する岩元素範囲ダメージを与える。
// このダメージは元素スキルダメージ扱い。
//
// 「絹」は1回攻撃後または3秒経過後に退場する。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	c.Core.Log.NewEvent("c2 activated", glog.LogCharacterEvent, c.Index)

	// 既存の2凸ティッカーを破棄
	c.kill(c.c2Ticker)

	// 新しい2凸ティッカーを生成
	// 2凸の持続時間と生成間隔はヒットラグの影響を受ける
	t := newTicker(c.Core, c2Duration, c.QueueCharTask)
	t.cb = c.createKinu(c.Core.F, c2CenterOffset, c2MinRandom, c2MaxRandom)
	t.interval = c2SpawnInterval
	c.QueueCharTask(t.tick, c2SpawnInterval)
	c.c2Ticker = t
}

// 固有天賦「仕亊げ」の追加効果発動後8秒間、アクティブキャラの
// 通常攻撃、重撃、落下攻撃が付近の敵に命中すると、「絹」を
// その敵付近に召喚。1秒ごとに1体召喚可能、
// 「仕亊げ」の好機を見て/Tailoring効果毎に最大3体。
// 上記効果は15秒ごとに1回まで発動可能。
//
// 固有天賦「仕亊げ」の解放が必要。
func (c *char) c4Activation() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.Base.Cons < 4 {
		return
	}
	if c.StatusIsActive(c4Lockout) {
		return
	}

	c.Core.Log.NewEvent("c4 activated", glog.LogCharacterEvent, c.Index)

	c.AddStatus(c4Lockout, c4LockoutDuration, true) // 千織に適用

	c.c4AttackCount = 0
	c.DeleteStatus(c4ICDKey)
	c.AddStatus(c4Key, c4Duration, false) // チームに適用、ヒットラグの影響なし
}

func (c *char) c4() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.Base.Cons < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// 4凸ステータスがアクティブでない
		if !c.StatusIsActive(c4Key) {
			return false
		}
		// 4凸攻撃がICD中
		if c.StatusIsActive(c4ICDKey) {
			return false
		}
		// 通常攻撃/重撃/落下攻撃以外の攻撃
		atk := args[1].(*combat.AttackEvent)
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		case attacks.AttackTagPlunge:
		default:
			return false
		}
		// アクティブキャラクター以外の攻撃
		if atk.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}

		// ICDを適用
		c.AddStatus(c4ICDKey, c4ICD, true) // 千織に適用

		c.Core.Log.NewEvent("c4 spawning kinu", glog.LogCharacterEvent, c.Index)

		// 「絹」を召喚
		c.createKinu(c.Core.F, c4CenterOffset, c4MinRandom, c4MaxRandom)()

		// 攻撃カウントを増加し、上限に達したら4凸を削除
		c.c4AttackCount++
		if c.c4AttackCount == c4AttackLimit {
			c.DeleteStatus(c4Key)
		}

		return false
	}, "chiori-c4")
}

// 固有天賦「仕亊げ」の追加効果発動後、
// 千織自身の羽ばたきの別れのCDが12秒短縮される。
// 固有天賦「仕亊げ」の解放が必要。
func (c *char) c6CooldownReduction() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.Base.Cons < 6 {
		return
	}
	c.ReduceActionCooldown(action.ActionSkill, 12*60)
}

// さらに、千織自身の通常攻撃ダメージが自身の防御力の235%分増加する。
func (c *char) c6NAIncrease(ai *combat.AttackInfo, snap *combat.Snapshot) {
	if c.Base.Ascension < 1 {
		return
	}
	if c.Base.Cons < 6 {
		return
	}
	ai.FlatDmg = snap.Stats.TotalDEF()
	ai.FlatDmg *= 2.35
}
