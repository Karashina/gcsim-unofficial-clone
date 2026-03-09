package lisa

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const (
	// ヒットマークフレーム（重撃の溜め時間を含む）
	chargeHitmark = 70
	// TODO: スタックは技術的には15秒しか持たず、各スタックに独自のタイマーがある
	conductiveTag = "lisa-conductive-stacks"
)

func init() {
	chargeFrames = frames.InitAbilSlice(93)
	chargeFrames[action.ActionAttack] = 91
	chargeFrames[action.ActionCharge] = 90
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = 90
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// 通常攻撃またはスキル（長押し）中なら重撃のモーション開始をスキップ
	windup := 0
	switch c.Core.Player.CurrentState() {
	case action.NormalAttackState:
		windup = 14
	case action.SkillState:
		if c.Core.Player.LastAction.Param["hold"] != 0 {
			windup = 14
		}
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			geometry.Point{Y: 1},
			10,
			40,
		),
		chargeHitmark-windup,
		chargeHitmark-windup,
		c.makeA1CB(),
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeHitmark - windup,
		State:           action.ChargeAttackState,
	}, nil
}
