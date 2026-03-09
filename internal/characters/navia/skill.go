package navia

import (
	"fmt"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

var (
	skillFrames     [][][]int
	skillMultiplier = []float64{
		0,                   // 0 hits
		1,                   // 1 hits
		1.05000000074505806, // 2ヒット
		1.10000000149011612, // 3 hits etc
		1.15000000596046448,
		1.20000000298023224,
		1.36000001430511475,
		1.4000000059604645,
		1.6000000238418579,
		1.6660000085830688,
		1.8999999761581421,
		2,
	}
	hitscans = [][]float64{
		// 幅、角度、Xオフセット、Yオフセット
		{0.9, 0, 0, 0},
		{0.25, 7.911, 0, 0},
		{0.25, -1.826, 0, 0},
		{0.25, -4.325, 0, 0},
		{0.25, 0.773, 0, 0},
		{0.25, 6.209, 0, 0},
		{0.25, -2.752, 0, 0},
		{0.25, 7.845, 0.01, 0.01},
		{0.25, -7.933, -0.01, -0.01},
		{0.25, 2.626, 0, 0},
		{0.25, -9.993, 0, 0},
	}
)

const (
	travelDelay = 9

	skillPressCDStart = 11

	skillHoldCDStart     = 41
	skillMaxHoldDuration = 241 // 長押し有効化時に1が設定されることを考慮して1fを追加

	bulletBoxLength = 11.5

	particleICDKey = "navia-particle-icd"
	arkheICDKey    = "navia-arkhe-icd"
)

func init() {
	skillFrames = make([][][]int, 2)

	// 単押し
	skillFrames[0] = make([][]int, 2)

	// 通常
	skillFrames[0][0] = frames.InitAbilSlice(40) // E -> E/Q
	skillFrames[0][0][action.ActionAttack] = 38
	skillFrames[0][0][action.ActionDash] = 24
	skillFrames[0][0][action.ActionJump] = 24
	skillFrames[0][0][action.ActionWalk] = 35
	skillFrames[0][0][action.ActionSwap] = 38

	// 3個以上の榴弾
	skillFrames[0][1] = frames.InitAbilSlice(41) // E -> E/Q
	skillFrames[0][1][action.ActionAttack] = 40
	skillFrames[0][1][action.ActionDash] = 26
	skillFrames[0][1][action.ActionJump] = 24
	skillFrames[0][1][action.ActionWalk] = 39
	skillFrames[0][1][action.ActionSwap] = 40

	// 長押し
	skillFrames[1] = make([][]int, 2)

	// 通常
	skillFrames[1][0] = frames.InitAbilSlice(71) // E -> E/Q
	skillFrames[1][0][action.ActionAttack] = 70
	skillFrames[1][0][action.ActionDash] = 54
	skillFrames[1][0][action.ActionJump] = 54
	skillFrames[1][0][action.ActionWalk] = 65
	skillFrames[1][0][action.ActionSwap] = 69

	// 3個以上の榴弾
	skillFrames[1][1] = frames.InitAbilSlice(73) // E -> E/Q
	skillFrames[1][1][action.ActionDash] = 56
	skillFrames[1][1][action.ActionJump] = 56
	skillFrames[1][1][action.ActionWalk] = 70
	skillFrames[1][1][action.ActionSwap] = 71
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// フレーム関連の設定
	holdIndex := 0
	shrapnelIndex := 0
	firingTime := skillPressCDStart // デフォルトで単押しと仮定

	// 長押しをチェック
	// 長押し時間は最小長押し時間に加算される
	hold := max(p["hold"], 0)
	if hold > 0 {
		// 長押しスキルフレームを使用
		holdIndex = 1
		// 最小長押しを超えた最大長押し時間でキャップ
		if hold > skillMaxHoldDuration {
			hold = skillMaxHoldDuration
		}
		// 長押しを示すために>0を供給する必要があるため1を引く
		hold -= 1
		// firingTimeを計算
		firingTime = skillHoldCDStart + hold

		// 結晶引き寄せ関連
		// 0.2で単押しが長押しに変換され、吸引が開始
		firingTimeF := c.Core.F + firingTime
		for i := 12; i < firingTime; i += 30 {
			c.pullCrystals(firingTimeF, i)
		}
	}
	c.SetCDWithDelay(action.ActionSkill, 9*60, firingTime)

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Rosula Shardshot",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   150,
		Element:    attributes.Geo,
		Durability: 25,
	}

	c.QueueCharTask(
		func() {
			if c.shrapnel >= 3 {
				shrapnelIndex = 1
			}
			c.Core.Log.NewEvent(fmt.Sprintf("firing %v crystal shrapnel", c.shrapnel), glog.LogCharacterEvent, c.Index)

			// スナップショットしてバフを追加
			snap := c.Snapshot(&ai)
			c.addShrapnelBuffs(&snap, c.shrapnel)

			// 各弾丸の軌道上の敵を検索
			// まず命中ゾーンのみをスキャンしてチェック対象を絞り込み
			shots := 5 + min(c.shrapnel, 3)*2
			for _, t := range c.Core.Combat.EnemiesWithinArea(
				combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{X: -0.20568, Y: -0.043841}, 4.0722, 11.5461),
				nil,
			) {
				// ヒット数を集計
				hits := 0
				for i := 0; i < shots; i++ {
					if ok, _ := t.AttackWillLand(
						combat.NewBoxHitOnTarget(
							c.Core.Combat.Player(),
							geometry.Point{X: hitscans[i][2], Y: hitscans[i][3]}.Rotate(geometry.DegreesToDirection(hitscans[i][1])),
							hitscans[i][0],
							bulletBoxLength,
						)); ok {
						hits++
					}
				}
				c.Core.Log.NewEvent(fmt.Sprintf("target %v hit %v times", t.Key(), hits), glog.LogCharacterEvent, c.Index)
				// ヒット数に基づいてダメージを適用
				ai.Mult = skillshotgun[c.TalentLvlSkill()] * skillMultiplier[hits]
				c.Core.QueueAttackWithSnap(
					ai,
					snap,
					combat.NewSingleTargetHit(t.Key()),
					travelDelay,
					c.particleCB,
					c.c2(),
				)
			}
			c.surgingBlade(c.shrapnel)

			// 発射時に固有天賦1と1凸を発動
			c.a1()
			c.c1(c.shrapnel)

			// 発射後にShrapnelを除去
			if c.Base.Cons < 6 {
				c.shrapnel = 0
			} else {
				c.shrapnel = max(c.shrapnel-3, 0) // 6凸は3を超える分を保持
			}
		},
		firingTime,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return skillFrames[holdIndex][shrapnelIndex][next] + hold },
		AnimationLength: skillFrames[holdIndex][1][action.InvalidAction] + hold,
		CanQueueAfter:   skillFrames[holdIndex][0][action.ActionJump] + hold,
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	e := a.Target.(*enemy.Enemy)
	if e.Type() != targets.TargettableEnemy {
		return
	}

	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.2*60, true)

	count := 3.0
	if c.Core.Rand.Float64() < 0.5 {
		count = 4
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Geo, c.ParticleDelay)
}

// snapを変更してバフを追加
// 発射時に計算された余剰分がその発射のsurgingBladeにも使用されるためこの方法が必要
func (c *char) addShrapnelBuffs(snap *combat.Snapshot, count int) {
	// 余剰Shrapnelに基づくバフを計算
	excess := float64(max(count-3, 0))

	dmg := 0.15 * excess
	cr := 0.0
	cd := 0.0
	if c.Base.Cons >= 2 {
		cr = 0.12 * float64(min(count, 3))
	}
	if c.Base.Cons >= 6 {
		cd = 0.45 * excess
	}
	snap.Stats[attributes.DmgP] += dmg
	snap.Stats[attributes.CR] += cr
	snap.Stats[attributes.CD] += cd
	c.Core.Log.NewEvent("adding shrapnel buffs", glog.LogCharacterEvent, c.Index).Write("dmg%", dmg).Write("cr", cr).Write("cd", cd)
}

// Shrapnelスタックを結晶化シールド拾得時に追加。
// スタックは300秒持続だが長すぎるので省略。
// パーティメンバーが結晶化反応で生成された元素の欠片を獲得すると、
// ナヴィアはCrystal Shrapnelを1スタック獲得。最大6スタック保持可能。
// Crystal Shrapnel獲得時、既存の欠片の持続時間がリセットされる。
func (c *char) shrapnelGain() {
	c.Core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		// シールドをチェック
		shd := args[0].(shield.Shield)
		if shd.Type() != shield.Crystallize {
			return false
		}

		if c.shrapnel < 6 {
			c.shrapnel++
			c.Core.Log.NewEvent("Crystal Shrapnel gained from Crystallise", glog.LogCharacterEvent, c.Index).Write("shrapnel", c.shrapnel)
		}
		return false
	}, "shrapnel-gain")
}

func (c *char) surgingBlade(count int) {
	if c.StatusIsActive(arkheICDKey) {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Surging Blade",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Geo,
		Durability: 0,
		Mult:       skillblade[c.TalentLvlSkill()],
	}

	// 攻撃位置を決定
	player := c.Core.Combat.Player()
	// ショットガン範囲
	e := c.Core.Combat.ClosestEnemyWithinArea(combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{X: -0.20568, Y: -0.043841}, 4.0722, 11.5461), nil)
	// デフォルト位置はプレイヤー + Y: 3.6
	pos := geometry.CalcOffsetPoint(player.Pos(), geometry.Point{Y: 3.6}, player.Direction())
	if e != nil {
		// ショットガン範囲に敵がいる: 敵の位置を使用
		pos = e.Pos()
	}

	// アラインドCDトリガーは遅延され、トリガー後にアラインド攻撃タスクをキューすべき
	c.QueueCharTask(func() {
		c.AddStatus(arkheICDKey, 7*60, true)
		c.QueueCharTask(func() {
			snap := c.Snapshot(&ai)
			c.addShrapnelBuffs(&snap, count)
			c.Core.QueueAttackWithSnap(
				ai,
				snap,
				combat.NewCircleHitOnTarget(pos, nil, 3),
				0,
			)
		}, 36)
	}, 28)
}

// ナヴィアに結晶を引き寄せる。データマインによると範囲12mのため、
// 30fごとにチェック。
func (c *char) pullCrystals(firingTimeF, i int) {
	c.Core.Tasks.Add(func() {
		for _, g := range c.Core.Combat.Gadgets() {
			cs, ok := g.(*reactable.CrystallizeShard)
			// 欠片でない場合はスキップ
			if !ok {
				continue
			}
			// 欠片が12m範囲外ならスキップ
			if !cs.IsWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 12)) {
				continue
			}

			// 吸引速度を1フレームあたり約0.4mと見積もり（約8mの距離で五郎に到20fかかった）
			distance := cs.Pos().Distance(c.Core.Combat.Player().Pos())
			travel := int(math.Ceil(distance / 0.4))
			// 結晶が発射前に到着しない場合はスキップ
			if c.Core.F+travel >= firingTimeF {
				continue
			}
			// 結晶が生成直後で拾える前に到着するエッジケースのための特別チェック
			if c.Core.F+travel < cs.EarliestPickup {
				continue
			}

			c.Core.Tasks.Add(func() {
				cs.AddShieldKillShard()
			}, travel)
		}
	}, i)
}
