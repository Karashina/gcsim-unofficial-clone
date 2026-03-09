package chiori

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
	// 固有天賦1
	// スキル使用後に固有天賢1を発動可能な時間枠
	a1WindowKey = "chiori-a1-window"
	// Tapestry - 連携攻撃
	a1SeizeTheMomentKey         = "chiori-seize-the-moment"
	a1SeizeTheMomentDuration    = 8 * 60
	a1SeizeTheMomentICDKey      = "chiori-seize-the-moment-icd"
	a1SeizeTheMomentICD         = 2 * 60
	a1SeizeTheMomentAttackLimit = 2
	// Tailoring - 岩元素付与
	a1GeoInfusionKey      = "chiori-tailoring"
	a1GeoInfusionDuration = 5 * 60
	// 固有天賦2
	a4BuffKey  = "chiori-a4"
	a4Duration = 20 * 60
)

// 羽ばたきの別れの上昇攻撃使用後、短時間内に取る次のアクションに応じて異なる効果を得る。
// 元素スキルを押すと「Tapestry」効果が発動。
// 通常攻撃を押す/タップすると「Tailoring」効果が発動。
//
// Tapestry
// - ロスターの次のキャラクターに切り替え。
// - 全パーティメンバーに「好機を見て」を付与: アクティブキャラの通常攻撃、
// 重撃、落下攻撃が付近の敵に命中すると、「Tamoto」が連携攻撃を実行。
// 羽ばたきの別れの上昇攻撃ダメージの100%の岩元素範囲ダメージ。
// このダメージは元素スキルダメージ扱い。
// - 「好機を見て」は8秒間持続、2秒ごとに1回の連携攻撃が可能。
// 効果時間中に最大2回の連携攻撃が発生可能。
//
// Tailoring
// - 千織が5秒間岩元素付与を得る。
//
// フィールド上で、羽ばたきの別れの上昇攻撃後短時間内に
// 元素スキル押しまたは通常攻撃を行わなかった場合、
// デフォルトでTailoring効果が発動する。
func (c *char) a1TapestrySetup() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// 好機を見てがアクティブでない
		if !c.StatusIsActive(a1SeizeTheMomentKey) {
			return false
		}
		// 好機を見てがICD中
		if c.StatusIsActive(a1SeizeTheMomentICDKey) {
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
		// アクティブキャラによる攻撃でない
		if atk.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		// 攻撃がプレイヤーから30m以内でない
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if !t.IsWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 30)) {
			return false
		}

		// ICDを適用
		c.AddStatus(a1SeizeTheMomentICDKey, a1SeizeTheMomentICD, true)

		// ダメージを与える
		ai := combat.AttackInfo{
			Abil:       "Fluttering Hasode (Seize the Moment)",
			ActorIndex: c.Index,
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagChioriSkill,
			ICDGroup:   attacks.ICDGroupChioriSkill,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Geo,
			Durability: 25,
			Mult:       thrustAtkScaling[c.TalentLvlSkill()],
		}
		snap := c.Snapshot(&ai)
		ai.FlatDmg = snap.Stats.TotalDEF()
		ai.FlatDmg *= thrustDefScaling[c.TalentLvlSkill()]
		c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(t, nil, 2.5), 1)

		// 攻撃カウントを増加し、上限に達した場合は好機を見てを削除
		c.a1AttackCount++
		if c.a1AttackCount == a1SeizeTheMomentAttackLimit {
			c.DeleteStatus(a1SeizeTheMomentKey)
		}

		return false
	}, a1SeizeTheMomentKey)
}

// 単押しと長押しで固有天賦1ウィンドウの開始と持続時間が異なるため、
// 固有天賦1の発動は指定された値に基づく
func (c *char) activateA1Window(start, duration int) {
	if c.Base.Ascension < 1 {
		return
	}
	c.QueueCharTask(func() {
		c.AddStatus(a1WindowKey, duration, true)
		// フィールド上で、羽ばたきの別れの上昇攻撃後短時間内に
		// 元素スキル押しまたは通常攻撃を行わなかった場合、
		// デフォルトでTailoring効果が発動する。
		c.a1Triggered = false
		c.QueueCharTask(func() {
			if c.a1Triggered {
				return
			}
			c.a1Tailoring()
		}, duration)
	}, start)
}

func (c *char) commonA1Trigger() {
	c.a1Triggered = true
	// これを発動するとスキルを再使用できなくなる
	c.DeleteStatus(a1WindowKey)
	c.c4Activation()
	c.c6CooldownReduction()
}

func (c *char) a1Tapestry() {
	c.commonA1Trigger()

	c.Core.Log.NewEvent("a1 tapestry triggered", glog.LogCharacterEvent, c.Index)
	c.AddStatus(a1SeizeTheMomentKey, a1SeizeTheMomentDuration, true)
	c.a1AttackCount = 0
}

func (c *char) a1Tailoring() {
	c.commonA1Trigger()

	c.Core.Log.NewEvent("a1 tailoring triggered", glog.LogCharacterEvent, c.Index)
	c.Core.Player.AddWeaponInfuse(
		c.Index,
		a1GeoInfusionKey,
		attributes.Geo,
		a1GeoInfusionDuration,
		true,
		attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
	)
}

// 通常攻撃経由のTailoring発動は、固有天賦1ウィンドウ期限切れで既に発動済みの場合失敗する可能性がある
func (c *char) tryTriggerA1TailoringNA() {
	if c.Base.Ascension < 1 {
		return
	}
	if !c.StatusIsActive(a1WindowKey) {
		return
	}
	c.a1Tailoring()
}

// パーティメンバーが岩元素設置物を生成した時、千織は20秒間岩元素ダメージボーナス+20%を得る。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4Buff = make([]float64, attributes.EndStatType)
	c.a4Buff[attributes.GeoP] = 0.20
	c.Core.Events.Subscribe(event.OnConstructSpawned, func(args ...interface{}) bool {
		c.applyA4Buff()
		return false
	}, a4BuffKey)
}

// 1凸のrock dollが固有天賦4を発動するため、単独で呼び出せる必要がある
func (c *char) applyA4Buff() {
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(a4BuffKey, a4Duration),
		AffectedStat: attributes.GeoP,
		Amount: func() ([]float64, bool) {
			return c.a4Buff, true
		},
	})
}
