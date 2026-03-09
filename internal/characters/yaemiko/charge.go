package yaemiko

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 90

func init() {
	chargeFrames = frames.InitAbilSlice(96) // CA -> N1/E/Q
	chargeFrames[action.ActionCharge] = 95  // CA -> CA
	chargeFrames[action.ActionDash] = 46    // CA -> D
	chargeFrames[action.ActionJump] = 47    // CA -> J
	chargeFrames[action.ActionSwap] = 94    // CA -> Swap
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupYaeCharged,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// 通常攻撃アニメーション中の場合は重撃溜め時間をスキップ
	windup := 0
	if c.Core.Player.CurrentState() == action.NormalAttackState {
		windup = 14
	}

	// 記録されたヒットマークから開始し、ターゲット方向に+1.65m
	// 11m/sで移動、0.15秒（9フレーム）ごとに1回攻撃、つまり1攻撃あたり11 * 0.15 = 1.65m移動
	// 特殊ダメージシーケンスにより制限（0.5秒ごとに1回）
	initialPos := c.Core.Combat.Player().Pos()
	initialDirection := c.Core.Combat.Player().Direction()
	for i := 0; i < 5; i++ {
		nextPos := geometry.CalcOffsetPoint(initialPos, geometry.Point{Y: 1.65 * float64(i+1)}, initialDirection)
		c.Core.QueueAttack(
			ai,
			// このループ中にプライマリターゲットの位置は変更できないため方向は同じはず
			combat.NewBoxHit(c.Core.Combat.Player(), nextPos, nil, 2, 2),
			0, // TODO: スナップショット遅延を確認
			chargeHitmark+(i*9)-windup,
		)
	}

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeFrames[action.ActionDash] - windup, // 最速キャンセルはヒットマークより前
		State:           action.ChargeAttackState,
	}, nil
}
