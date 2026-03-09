package diluc

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const burstHitmark = 100

func init() {
	burstFrames = frames.InitAbilSlice(140) // Q -> D
	burstFrames[action.ActionAttack] = 139  // Q -> N1
	burstFrames[action.ActionSkill] = 139   // Q -> E
	burstFrames[action.ActionJump] = 139    // Q -> J
	burstFrames[action.ActionSwap] = 138    // Q -> Swap
}

const burstBuffKey = "diluc-q"

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 固有天賦2:
	// 黎明の炎元素付与の持続時間が4秒延長される
	duration := 480
	hasA4 := c.Base.Ascension >= 4
	if hasA4 {
		duration += 240
	}

	// 元素付与は元素爆発開始時に始まりCD終了時に終了する（ディルックの動画で確認可能）
	c.AddStatus(burstBuffKey, duration, true)
	c.Core.Events.Emit(event.OnInfusion, c.Index, attributes.Pyro, duration)

	// 固有天賦2:
	// さらに、効果持続中ディルックの炎元素ダメージ+20%
	if hasA4 {
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(burstBuffKey, duration),
			AffectedStat: attributes.PyroP,
			Amount: func() ([]float64, bool) {
				return c.a4buff, true
			},
		})
	}

	// スナップショットはアニメーション後半、大剣から放たれる時点で行われる
	// ここではダメージ発生時にスナップショットする
	c.Core.Tasks.Add(func() {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               "Dawn (Strike)",
			AttackTag:          attacks.AttackTagElementalBurst,
			ICDTag:             attacks.ICDTagElementalBurst,
			ICDGroup:           attacks.ICDGroupDiluc,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           100,
			Element:            attributes.Pyro,
			Durability:         50,
			Mult:               burstInitial[c.TalentLvlBurst()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.09 * 60,
			CanBeDefenseHalted: true,
		}

		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1}, 16, 6),
			0,
			1,
		)

		ai.StrikeType = attacks.StrikeTypeDefault
		// 初撃、持続ダメージ、爆発すべて元素量50
		ai.Abil = "Dawn (Tick)"
		ai.Mult = burstDOT[c.TalentLvlBurst()]

		// ヒットラグは初撃のみ
		ai.HitlagHaltFrames = 0
		ai.CanBeDefenseHalted = false

		// 持続ダメージと爆発ダメージ
		// - ガジェットはY: 1mに生成され、爆発まで約1.7秒存在
		// - 14 m/sで移動し0.2秒ごとにダメージ、1攻撃あたり2.8m移動
		// - 1.7s / (0.2 s/攻撃) ≒ 爆発前に計8回攻撃
		initialPos := c.Core.Combat.Player().Pos()
		initialDirection := c.Core.Combat.Player().Direction()
		for i := 0; i < 8; i++ {
			nextPos := geometry.CalcOffsetPoint(initialPos, geometry.Point{Y: 1 + 2.8*float64(i)}, initialDirection)
			c.Core.QueueAttack(
				ai,
				combat.NewBoxHit(c.Core.Combat.Player(), nextPos, geometry.Point{Y: -5}, 16, 8),
				0,
				(i+1)*12,
			)
		}

		ai.Abil = "Dawn (Explode)"
		ai.Mult = burstExplode[c.TalentLvlBurst()]
		// 1m + 14 m/s * 1.7s
		finalPos := geometry.CalcOffsetPoint(initialPos, geometry.Point{Y: 1 + 14*1.7}, initialDirection)
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHit(c.Core.Combat.Player(), finalPos, geometry.Point{Y: -6}, 16, 10),
			0,
			1.7*60,
		)
	}, burstHitmark)

	c.ConsumeEnergy(21)
	c.SetCDWithDelay(action.ActionBurst, 720, 14)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
