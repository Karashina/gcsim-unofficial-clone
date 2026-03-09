package illuga

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var burstFrames []int

const (
	burstHitmark          = 50
	orioleSongDur         = 1255 // 20s (I-3 fix: was 15s)
	baseNightingaleStacks = 21
	stacksPerGeoConstruct = 5
)

func init() {
	burstFrames = frames.InitAbilSlice(66)
}

// Burst は影なき映し・Oriole-Songを実行する
// Oriole-Song状態に入る。ナイチンゲールの歌はパーティの岩元素ヒットにFlatDmgを追加
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 初期ナイチンゲールスタックを計算
	// 基本21 + 岩元素構築物1体につき5（構築物最大3体）
	geoConstructs := c.countGeoConstructs()
	initialConstructStacks := geoConstructs * stacksPerGeoConstruct
	if initialConstructStacks > 15 {
		initialConstructStacks = 15
	}
	c.geoConstructBonusStacks = initialConstructStacks // 初期構築物スタックを15上限に向けて追跡
	c.nightingaleSongStacks = baseNightingaleStacks + initialConstructStacks

	// 元素爆発初撃
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Shadowless Reflection",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 50,
	}

	// burst_genデータからの元素熟知 + 防御力スケーリング
	em := c.Stat(attributes.EM)
	def := c.TotalDef(false)
	emMult := burstEM[c.TalentLvlBurst()]
	defMult := burstDEF[c.TalentLvlBurst()]

	ai.FlatDmg = em*emMult + def*defMult

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 8)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, burstHitmark)

	// Oriole-Song状態に入る
	c.QueueCharTask(func() {
		c.enterOrioleSong()
	}, burstHitmark)

	// 固有天賦1の強化バフを適用（Ascendant Gleamチェック）
	c.applyLightkeeperOath()

	// I-4修正: CD 18秒 → 15秒
	c.SetCDWithDelay(action.ActionBurst, 15*60, burstHitmark)
	c.ConsumeEnergy(4)

	c.Core.Log.NewEvent("Illuga uses Shadowless Reflection", glog.LogCharacterEvent, c.Index).
		Write("nightingale_stacks", c.nightingaleSongStacks).
		Write("geo_constructs", geoConstructs).
		Write("flat_dmg", ai.FlatDmg)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// enterOrioleSong はOriole-Song状態を発動する
func (c *char) enterOrioleSong() {
	c.orioleSongActive = true
	c.orioleSongSrc = c.Core.F
	c.AddStatus(orioleSongKey, orioleSongDur, true)

	// I-1/I-2リファクタ: ナイチンゲールの歌はOnEnemyHit経由でヒット毎にFlatDmgを使用
	// GeoP% StatModの代わりにburstGeoBonusEMとburstLCrsBonusEMを使用。
	c.subscribeNightingaleStackConsumption()

	// I-GC1: 岩元素構築物生成を購読し、動的にスタックを獲得
	c.subscribeGeoConstructStacks()

	// モード終了をスケジュール
	src := c.orioleSongSrc
	c.QueueCharTask(func() {
		if c.orioleSongSrc != src {
			return
		}
		c.exitOrioleSong()
	}, orioleSongDur)

	c.Core.Log.NewEvent("Illuga enters Oriole-Song", glog.LogCharacterEvent, c.Index).
		Write("nightingale_stacks", c.nightingaleSongStacks).
		Write("duration", orioleSongDur)
}

// subscribeNightingaleStackConsumption はパーティメンバーの岩元素ダメージヒットを購読する
// I-1: ナイチンゲールの歌は対象の岩元素ヒット毎にFlatDmgを追加し、1ヒットにつき1スタック消費
// I-13: 全パーティメンバーの岩元素攻撃に適用
// I-14: 通常攻撃/重撃/落下攻撃/元素スキル/元素爆発の攻撃タグのみ対象
func (c *char) subscribeNightingaleStackConsumption() {
	cb := func(args ...interface{}) bool {
		if !c.orioleSongActive {
			return false
		}
		if c.nightingaleSongStacks <= 0 {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		t := args[0].(combat.Target)

		// 敵へのヒットのみ適用
		if t.Type() != targets.TargettableEnemy {
			return false
		}

		// 岩元素ヒットのみ対象
		if atk.Info.Element != attributes.Geo {
			return false
		}

		// I-14: 通常攻撃/重撃/落下攻撃/元素スキル/元素爆発の攻撃タグのみ対象
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal,
			attacks.AttackTagExtra,
			attacks.AttackTagPlunge,
			attacks.AttackTagElementalArt,
			attacks.AttackTagElementalArtHold,
			attacks.AttackTagElementalBurst,
			attacks.AttackTagLCrsDamage:
			// 有効
		default:
			return false
		}

		// I-1/I-2: イルーガの元素熟知と天賦レベルに基づきFlatDmgを追加
		illugaEM := c.Stat(attributes.EM)
		var flatDmg float64
		if atk.Info.AttackTag == attacks.AttackTagLCrsDamage {
			// LCrsヒットはburstLCrsBonusEM乗数を使用
			flatDmg = burstLCrsBonusEM[c.TalentLvlBurst()] * illugaEM
			flatDmg += c.getA4LCrsBonus() // 固有天賦4 LCrs強化
		} else {
			// 岩元素ヒットはburstGeoBonusEM乗数を使用
			flatDmg = burstGeoBonusEM[c.TalentLvlBurst()] * illugaEM
			flatDmg += c.getA4GeoBonus() // 固有天賦4 岩元素強化
		}
		atk.Info.FlatDmg += flatDmg

		// 敵ヒット毎に1スタック消費
		c.nightingaleSongStacks--

		// 2命: ランプ攻撃に十分なスタックが消費されたかチェック
		if c.Base.Cons >= 2 {
			c.c2StackCounter++
			for c.c2StackCounter >= 7 {
				c.c2LampAttack()
				c.c2StackCounter -= 7
			}
		}

		c.Core.Log.NewEvent("Illuga Nightingale stack consumed", glog.LogCharacterEvent, c.Index).
			Write("remaining_stacks", c.nightingaleSongStacks).
			Write("flat_dmg_added", flatDmg)

		if c.nightingaleSongStacks <= 0 {
			c.nightingaleSongStacks = 0
			c.exitOrioleSong()
		}

		return false
	}

	// I-10: 元素爆発再使用時の重複登録を避けるためキーベースの購読を使用
	c.Core.Events.Subscribe(event.OnEnemyHit, cb, "illuga-nightingale-consume")
}

// exitOrioleSong はOriole-Song状態を終了する
func (c *char) exitOrioleSong() {
	c.orioleSongActive = false
	c.orioleSongSrc = -1
	c.DeleteStatus(orioleSongKey)
	c.nightingaleSongStacks = 0
	c.c2StackCounter = 0
	c.geoConstructBonusStacks = 0

	c.Core.Log.NewEvent("Illuga exits Oriole-Song", glog.LogCharacterEvent, c.Index)
}

// subscribeGeoConstructStacks は岩元素構築物生成イベントを購読する
// Oriole-Song中に岩元素構築物が生成されると、現在のフィールド上の
// 構築物1体につき5スタック獲得する（追加スタック上限15）。
func (c *char) subscribeGeoConstructStacks() {
	c.Core.Events.Subscribe(event.OnConstructSpawned, func(args ...interface{}) bool {
		if !c.orioleSongActive {
			return false
		}
		if c.geoConstructBonusStacks >= 15 {
			return false
		}

		constructCount := c.countGeoConstructs()
		gain := constructCount * stacksPerGeoConstruct
		remaining := 15 - c.geoConstructBonusStacks
		if gain > remaining {
			gain = remaining
		}

		c.nightingaleSongStacks += gain
		c.geoConstructBonusStacks += gain

		c.Core.Log.NewEvent("Illuga gains Nightingale stacks from Geo Construct", glog.LogCharacterEvent, c.Index).
			Write("gain", gain).
			Write("total_bonus_stacks", c.geoConstructBonusStacks).
			Write("nightingale_stacks", c.nightingaleSongStacks)

		return false
	}, "illuga-geo-construct-stacks")
}

// countGeoConstructs はアクティブな岩元素構築物の数を返す
func (c *char) countGeoConstructs() int {
	return c.Core.Constructs.Count()
}
