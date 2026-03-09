package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 17
const chargeHitlagHaltFrame = 0.06

func init() {
	chargeFrames = frames.InitAbilSlice(41) // CA -> N1
	chargeFrames[action.ActionDash] = 25    // CA -> D
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: 重撃 (ICDTagExtraAttack)
	// ICDグループ: 標準 (ICDGroupDefault)
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charged Attack",
		AttackTag:          attacks.AttackTagExtra, // 重撃はAttackTagExtraを使用
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               charge[c.TalentLvlAttack()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   chargeHitlagHaltFrame * 60,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), geometry.Point{Y: 0.3}, 2.2, 270),
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
