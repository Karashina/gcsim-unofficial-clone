package geo

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/construct"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

var burstFrames [][]int

const burstStart = 35   // クールダウン開始と一致
const burstHitmark = 51 // Initial Shockwave 1

func init() {
	burstFrames = make([][]int, 2)

	// 男性
	burstFrames[0] = frames.InitAbilSlice(67) // Q -> N1/E
	burstFrames[0][action.ActionDash] = 42    // Q -> D
	burstFrames[0][action.ActionJump] = 42    // Q -> J
	burstFrames[0][action.ActionSwap] = 51    // Q -> Swap

	// 女性
	burstFrames[1] = frames.InitAbilSlice(64) // Q -> E
	burstFrames[1][action.ActionAttack] = 62  // Q -> N1
	burstFrames[1][action.ActionDash] = 42    // Q -> D
	burstFrames[1][action.ActionJump] = 42    // Q -> J
	burstFrames[1][action.ActionSwap] = 49    // Q -> Swap
}

func (c *Traveler) Burst(p map[string]int) (action.Info, error) {
	hits, ok := p["hits"]
	if !ok {
		hits = 4 // 衝撃波ダメージの全4回が敵に命中すると仮定
	}
	maxConstructCount, ok := p["construct_limit"]
	if !ok {
		// 全4枚の壁が実際に出現すると仮定
		// 4未満にすると左上から反時計回りに壁が出現しなくなる
		maxConstructCount = 4
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Wake of Earth",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagTravelerWakeOfEarth,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   100,
		Element:    attributes.Geo,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)

	// 4凸
	// 岩潮の衝撃波は敵に命中するごとに5エネルギーを回復する。
	// 1回で最大25エネルギーまで回復可能。
	src := c.Core.F
	var c4cb combat.AttackCBFunc
	if c.Base.Cons >= 4 {
		energyCount := 0
		c4cb = func(a combat.AttackCB) {
			t, ok := a.Target.(*enemy.Enemy)
			if !ok {
				return
			}
			// TODO: フレーム0キャストに対処するための応急処置。この動作についてもう少し考える必要あり
			if t.GetTag("traveler-c4-src") == src && src > 0 {
				return
			}
			if energyCount >= 5 {
				return
			}
			t.SetTag("traveler-c4-src", src)
			c.AddEnergy("geo-traveler-c4", 5)
			energyCount++
		}
	}
	player := c.Core.Combat.Player()
	c.burstArea = combat.NewCircleHitOnTarget(player, nil, 7)
	// 1.1秒間持続、0.25秒ごとにティック
	for i := 0; i < hits; i++ {
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(c.burstArea.Shape.Pos(), nil, 6),
			burstHitmark+(i+1)*15,
			c4cb,
		)
	}

	// 4枚の壁は+-2.75, +-6.67の位置に出現（デフォルトの視点方向を想定）
	// 出現は右上から開始し、時計回りに進む
	// (2.75, 6.67)を反時計回りに(0, x)になるまで回転すると、角度は約22.5度
	// この角度は壁の視点方向の決定に使用される
	angles := []float64{22.5, 112.5, 202.5, 292.5}
	offsets := []geometry.Point{{X: 2.75, Y: 6.67}, {X: 2.75, Y: -6.67}, {X: -2.75, Y: -6.67}, {X: -2.75, Y: 6.67}}
	c.Core.Tasks.Add(func() {
		// 1凸
		// 岩潮の範囲内のパーティメンバーは会心率10%増加、中断耐性が向上。
		if c.Base.Cons >= 1 {
			c.Tags["wall"] = 1
		}
		if c.Base.Cons >= 1 {
			c.Core.Tasks.Add(c.c1(1), 60) // 1秒後にチェック開始
		}
		// 6凸
		// 岩潮のバリアの持続時間が5秒延長される。
		// 星落としの剣の隕石の持続時間が10秒延長される。
		dur := 15 * 60
		if c.Base.Cons >= 6 {
			dur += 300
		}
		// 指定された上限まで壁を出現させる
		for i := 0; i < maxConstructCount; i++ {
			dir := geometry.DegreesToDirection(angles[i]).Rotate(player.Direction())
			pos := geometry.CalcOffsetPoint(player.Pos(), offsets[i], player.Direction())
			c.Core.Constructs.NewNoLimitCons(c.newWall(dur, dir, pos), false)
		}
	}, burstStart)

	c.SetCDWithDelay(action.ActionBurst, 900, burstStart)
	c.ConsumeEnergy(37)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames[c.gender]),
		AnimationLength: burstFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   burstFrames[c.gender][action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

type wall struct {
	src    int
	expiry int
	char   *Traveler
	dir    geometry.Point
	pos    geometry.Point
}

func (c *Traveler) newWall(dur int, dir, pos geometry.Point) *wall {
	return &wall{
		src:    c.Core.F,
		expiry: c.Core.F + dur,
		char:   c,
		dir:    dir,
		pos:    pos,
	}
}

func (w *wall) OnDestruct() {
	if w.char.Base.Cons >= 1 {
		w.char.Tags["wall"] = 0
	}
}

func (w *wall) Key() int                         { return w.src }
func (w *wall) Type() construct.GeoConstructType { return construct.GeoConstructTravellerBurst }
func (w *wall) Expiry() int                      { return w.expiry }
func (w *wall) IsLimited() bool                  { return true }
func (w *wall) Count() int                       { return 1 }
func (w *wall) Direction() geometry.Point        { return w.dir }
func (w *wall) Pos() geometry.Point              { return w.pos }
