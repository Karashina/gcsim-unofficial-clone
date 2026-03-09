package ganyu

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstStart   = 122
	burstMarkKey = "ganyu-burst-mark"
)

func init() {
	burstFrames = frames.InitAbilSlice(125) // Q -> D/J
	burstFrames[action.ActionAttack] = 124  // Q -> N1
	burstFrames[action.ActionAim] = 124     // Q -> CA, assumed
	burstFrames[action.ActionSkill] = 124   // Q -> E
	burstFrames[action.ActionSwap] = 122    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Celestial Shower",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       shower[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)

	c.Core.Status.Add("ganyuburst", 15*60+burstStart)

	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	// 固有天賦4関連
	m := make([]float64, attributes.EndStatType)
	m[attributes.CryoP] = 0.2
	// burstStartから0.3秒ごとにTick
	for i := 0; i < 15*60; i += 18 {
		// 4凸関連
		c.Core.Tasks.Add(func() {
			// 元素爆発のTick
			enemy := c.Core.Combat.RandomEnemyWithinArea(
				burstArea,
				func(e combat.Enemy) bool {
					return !e.StatusIsActive(burstMarkKey)
				},
			)
			var pos geometry.Point
			if enemy != nil {
				pos = enemy.Pos()
				enemy.AddStatus(burstMarkKey, 1.45*60, true) // 同じ敵は1.45秒間再ターゲットされない
			} else {
				pos = geometry.CalcRandomPointFromCenter(burstArea.Shape.Pos(), 0.5, 9.5, c.Core.Rand)
			}
			// 一定の遅延後にダメージを与える
			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(pos, nil, 2.5), 8)

			// 固有天賦2:
			// 降璃天華はAoE内のアクティブパーティメンバーに氷元素ダメージ+20%を付与。
			if c.Base.Ascension >= 4 && c.Core.Combat.Player().IsWithinArea(burstArea) {
				active := c.Core.Player.ActiveChar()
				active.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("ganyu-field", 60),
					AffectedStat: attributes.CryoP,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}
			// 4凸デバフTick
			if c.Base.Cons >= 4 {
				enemies := c.Core.Combat.EnemiesWithinArea(burstArea, nil)
				// 3秒ごとにスタック増加するが4凸ステータスは毎Tick適用
				// 4凸は3秒間持続
				increase := i%180 == 0
				for _, e := range enemies {
					e.AddStatus(c4Key, c4Dur, true)
					if increase {
						c4Stacks := e.GetTag(c4Key) + 1
						if c4Stacks > 5 {
							c4Stacks = 5
						}
						e.SetTag(c4Key, c4Stacks)
						c.Core.Log.NewEvent(c4Key+" tick on enemy", glog.LogCharacterEvent, c.Index).
							Write("stacks", c4Stacks).
							Write("enemy key", e.Key())
					}
				}
			}
		}, i+burstStart)
	}

	// シミュレーションにクールダウンを追加
	c.SetCD(action.ActionBurst, 15*60)
	// エネルギーを消費
	c.ConsumeEnergy(3)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
