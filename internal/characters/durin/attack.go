package durin

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
	attackFrames          [][]int
	attackHitmarks        = []int{11, 10, 7, 35}              // 検証データから更新
	attackHitlagHaltFrame = []float64{0.03, 0.03, 0.00, 0.06} // N1: 0.03, N2: 0.03, N3-1: 0.00, N4: 0.06
	attackHitlagFactor    = []float64{0.01, 0.01, 0.01, 0.01}
	attackDefHalt         = []bool{true, true, true, true}
	attackHitboxes        = [][]float64{{1.8, 2.5}, {1.8}, {2.0}, {1.8, 2.5}}
	attackOffsets         = []float64{0.5, 0.5, 0.5, 0.5}
	attackN3Hitmark2      = 29   // N3 2段目
	attackN3HitlagHalt2   = 0.04 // N3 2段目のヒットラグ
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 26) // N1 -> Walk
	attackFrames[0][action.ActionAttack] = 18                             // N1 -> N2
	attackFrames[0][action.ActionCharge] = 20                             // N1 -> C
	attackFrames[0][action.ActionDash] = 16                               // N1 -> D

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 28) // N2 -> Walk
	attackFrames[1][action.ActionAttack] = 23                             // N2 -> N3
	attackFrames[1][action.ActionDash] = 13                               // N2 -> D

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 37) // N3 -> Walk
	attackFrames[2][action.ActionAttack] = 40                             // N3 -> N4
	attackFrames[2][action.ActionDash] = 29                               // N3 -> D

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3], 65) // N4 -> Walk
	attackFrames[3][action.ActionAttack] = 65                             // N4 -> N1
	attackFrames[3][action.ActionDash] = 37                               // N4 -> D
	attackFrames[3][action.ActionCharge] = 500                            // N4は重撃キャンセル不可
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	// Essential Transmutation中は通常攻撃が上書きされ、
	// "変容: 暗黒の否定"（3ヒット単体）を実行する。
	// これはゲーム内で元素スキルが通常攻撃を変更する挙動を再現している。
	if c.StatusIsActive(essentialTransmutationKey) {
		return c.skillDenialOfDarkness(nil)
	}

	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: 通常攻撃 (ICDTagNormalAttack)
	// ICDグループ: 標準 (ICDGroupDefault)
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               attack[c.NormalCounter][c.TalentLvlAttack()],
		HitlagFactor:       attackHitlagFactor[c.NormalCounter],
		HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter] * 60,
		CanBeDefenseHalted: attackDefHalt[c.NormalCounter],
	}

	var ap combat.AttackPattern
	if c.NormalCounter == 1 {
		// N2は円形判定
		ap = combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
		)
	} else {
		// その他の攻撃は扇形判定
		ap = combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
			attackHitboxes[c.NormalCounter][1],
		)
	}

	// N3は2ヒット
	if c.NormalCounter == 2 {
		c.Core.QueueAttack(ai, ap, attackHitmarks[c.NormalCounter], attackHitmarks[c.NormalCounter])
		ai.Abil = fmt.Sprintf("Normal %v (2nd Hit)", c.NormalCounter)
		ai.HitlagHaltFrames = attackN3HitlagHalt2 * 60
		c.Core.QueueAttack(ai, ap, attackN3Hitmark2, attackN3Hitmark2)
	} else {
		c.Core.QueueAttack(ai, ap, attackHitmarks[c.NormalCounter], attackHitmarks[c.NormalCounter])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
