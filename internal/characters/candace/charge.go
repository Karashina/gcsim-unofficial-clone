package candace

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var chargeFrames []int

const chargeHitmark = 25

func init() {
	chargeFrames = frames.InitAbilSlice(53)
	chargeFrames[action.ActionSkill] = 52
	chargeFrames[action.ActionBurst] = 52
	chargeFrames[action.ActionDash] = 29
	chargeFrames[action.ActionJump] = 28
	chargeFrames[action.ActionSwap] = 51
}

// 重撃のダメージキュー生成
// 非常に標準的 - 香菱などの他キャラクターと同様
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charge",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupPoleExtraAttack,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               charge[c.TalentLvlAttack()],
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			0.8,
		),
		0,
		chargeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeFrames[action.ActionJump], // 最速キャンセル
		State:           action.ChargeAttackState,
	}, nil
}
