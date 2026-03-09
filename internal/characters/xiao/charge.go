package xiao

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var chargeFrames []int

const chargeHitmark = 16

func init() {
	chargeFrames = frames.InitAbilSlice(45)
	chargeFrames[action.ActionSkill] = 38
	chargeFrames[action.ActionBurst] = 37
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = 43
}

// 重撃のダメージキュー生成
// 非常に標準的 - 香菱などの他キャラクターと同様
// Note: 元素爆発中の重撃は通常攻撃とICDを共有する
// TODO: 通常状態の重撃も同様のパターンかは不明
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charge",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               charge[c.TalentLvlAttack()],
		HitlagHaltFrames:   0.02 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	radius := 3.0
	if c.StatusIsActive(burstBuffKey) {
		radius = 3.2
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, radius, 200),
		chargeHitmark,
		chargeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}
