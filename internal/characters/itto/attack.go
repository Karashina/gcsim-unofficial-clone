package itto

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	attackFrames          [][][]int
	attackHitmarks        = []int{23, 25, 16, 48}
	attackPoiseDMG        = []float64{82.9, 77.9, 98.3, 124.5}
	attackHitlagHaltFrame = []float64{0.08, 0.08, 0.10, 0.10}
	attackHitboxes        = [][][]float64{{{2.5}, {2.5}, {2.5}, {3.2, 6}}, {{3.5}, {3.5}, {3.5}, {3.8, 8}}}
	attackOffsets         = [][]float64{{0.8, 0.8, 0.85, -1.5}, {0.8, 0.8, 0.8, -1.7}}
)

const normalHitNum = 4

type ittoAttackState int

const (
	InvalidAttackState ittoAttackState = iota - 1
	attack0Stacks
	attack1PlusStacks
	attackEndState
)

func init() {
	attackFrames = make([][][]int, attackEndState)
	attackFrames[attack0Stacks] = make([][]int, normalHitNum)
	attackFrames[attack0Stacks][0] = frames.InitNormalCancelSlice(attackHitmarks[0], 41) // N1 -> CA0
	attackFrames[attack0Stacks][1] = frames.InitNormalCancelSlice(attackHitmarks[1], 51) // N2 -> CA0
	attackFrames[attack0Stacks][2] = frames.InitNormalCancelSlice(attackHitmarks[2], 57) // N3 -> CA0
	attackFrames[attack0Stacks][3] = frames.InitNormalCancelSlice(attackHitmarks[3], 83) // N4 -> N1

	attackFrames[attack0Stacks][0][action.ActionAttack] = 33  // N1 -> N2
	attackFrames[attack0Stacks][1][action.ActionAttack] = 36  // N2 -> N3
	attackFrames[attack0Stacks][2][action.ActionAttack] = 43  // N3 -> N4
	attackFrames[attack0Stacks][3][action.ActionCharge] = 500 // N4 -> CA0, TODO: このアクションは不正; より良い処理方法が必要

	attackFrames[attack1PlusStacks] = make([][]int, normalHitNum)
	attackFrames[attack1PlusStacks][0] = frames.InitNormalCancelSlice(attackHitmarks[0], 33) // N1 -> N2
	attackFrames[attack1PlusStacks][1] = frames.InitNormalCancelSlice(attackHitmarks[1], 36) // N2 -> N3
	attackFrames[attack1PlusStacks][2] = frames.InitNormalCancelSlice(attackHitmarks[2], 43) // N3 -> N4
	attackFrames[attack1PlusStacks][3] = frames.InitNormalCancelSlice(attackHitmarks[3], 83) // N4 -> N1

	attackFrames[attack1PlusStacks][0][action.ActionCharge] = 23 // N1 -> CA1/CAF
	attackFrames[attack1PlusStacks][1][action.ActionCharge] = 27 // N2 -> CA1/CAF
	attackFrames[attack1PlusStacks][2][action.ActionCharge] = 21 // N3 -> CA1/CAF
	attackFrames[attack1PlusStacks][3][action.ActionCharge] = 52 // N4 -> CA1/CAF
}

func (c *char) attackState() ittoAttackState {
	if c.Tags[strStackKey] == 0 {
		// 0スタック: NX -> CA0 フレームを使用
		return attack0Stacks
	}
	// 1+スタック: NX -> CA1/CAF フレームを使用（ここでは同一）
	return attack1PlusStacks
}

// 通常攻撃:
// 最大4段の連続攻撃を行う。
// 2段目と4段目が敵に命中すると、一斗はそれぞれ怒髪衝天スタックを1と2獲得する。
// 最大5スタック。この効果が発動すると、既存スタックの持続時間がリセットされる。
// さらに、一斗の通常攻撃コンボはダッシュや元素スキル「魔殺絶技・岩牛発破!」使用後にリセットされない。
func (c *char) Attack(p map[string]int) (action.Info, error) {
	// さらに、一斗の通常攻撃コンボはダッシュや元素スキル使用後にリセットされない
	switch c.Core.Player.CurrentState() {
	case action.DashState, action.SkillState:
		c.NormalCounter = c.savedNormalCounter
	}

	// 攻撃
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
		Mult:               attack[c.NormalCounter][c.TalentLvlAttack()],
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           attackPoiseDMG[c.NormalCounter],
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter] * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	// 元素爆発状態をヒットボックスのために確認
	attackIndex := 0
	if c.StatModIsActive(burstBuffKey) {
		attackIndex = 1
	}
	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: attackOffsets[attackIndex][c.NormalCounter]},
		attackHitboxes[attackIndex][c.NormalCounter][0],
	)
	if c.NormalCounter == 3 {
		ap = combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[attackIndex][c.NormalCounter]},
			attackHitboxes[attackIndex][c.NormalCounter][0],
			attackHitboxes[attackIndex][c.NormalCounter][1],
		)
	}
	// TODO: ヒットマークが攻撃速度で調整されていない
	c.Core.QueueAttack(ai, ap, attackHitmarks[c.NormalCounter], attackHitmarks[c.NormalCounter])

	// TODO: 通常攻撃は常に命中すると仮定。重撃フレームを決定する際に次の重撃がCA0かCA1/CAFか判別不可能なため。
	// ダメージ時に怒髪衝天スタックを追加
	n := c.NormalCounter
	if n == 1 {
		c.addStrStack("attack", 1)
	} else if n == 3 {
		c.addStrStack("attack", 2)
	}
	if c.StatModIsActive(burstBuffKey) && (n == 0 || n == 2) {
		c.addStrStack("q-attack", 1)
	}

	// NX -> CA0/CA1/CAF フレームを処理
	state := c.attackState()

	defer func() {
		c.AdvanceNormalIndex()
		c.savedNormalCounter = c.NormalCounter
	}()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames[state]),
		AnimationLength: attackFrames[state][c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
