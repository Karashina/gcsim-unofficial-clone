package yoimiya

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var aimedFrames [][]int

var aimedHitmarks = []int{15, 86, 103, 121, 139}

func init() {
	aimedFrames = make([][]int, 5)

	// 狙い撃ち
	aimedFrames[0] = frames.InitAbilSlice(26)
	aimedFrames[0][action.ActionDash] = aimedHitmarks[0]
	aimedFrames[0][action.ActionJump] = aimedHitmarks[0]

	// フルチャージ狙い撃ち
	aimedFrames[1] = frames.InitAbilSlice(97)
	aimedFrames[1][action.ActionDash] = aimedHitmarks[1]
	aimedFrames[1][action.ActionJump] = aimedHitmarks[1]

	// フルチャージ狙い撃ち（火焰矢1本）
	aimedFrames[2] = frames.InitAbilSlice(114)
	aimedFrames[2][action.ActionDash] = aimedHitmarks[2]
	aimedFrames[2][action.ActionJump] = aimedHitmarks[2]

	// フルチャージ狙い撃ち（火焰矢2本）
	aimedFrames[3] = frames.InitAbilSlice(132)
	aimedFrames[3][action.ActionDash] = aimedHitmarks[3]
	aimedFrames[3][action.ActionJump] = aimedHitmarks[3]

	// フルチャージ狙い撃ち（火焰矢3本）
	aimedFrames[4] = frames.InitAbilSlice(150)
	aimedFrames[4][action.ActionDash] = aimedHitmarks[4]
	aimedFrames[4][action.ActionJump] = aimedHitmarks[4]
}

// 標準的な狙い撃ち攻撃
func (c *char) Aimed(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}
	weakspot := p["weakspot"]

	hold, ok := p["hold"]
	if !ok {
		hold = attacks.AimParamLv1
	}
	switch hold {
	case attacks.AimParamPhys:
	case attacks.AimParamLv1:
	case attacks.AimParamLv2:
	case attacks.AimParamLv3:
	case attacks.AimParamLv4:
	default:
		return action.Info{}, fmt.Errorf("invalid hold param supplied, got %v", hold)
	}

	// 火焰矢が命中するまでの時間を調整するためのパラメータ
	// hold < lv.2 の場合は何もしない
	kindlingTravel, ok := p["kindling_travel"]
	if !ok {
		kindlingTravel = 30
	}

	// 通常の矢
	ai := combat.AttackInfo{
		ActorIndex:           c.Index,
		Abil:                 "Fully-Charged Aimed Shot",
		AttackTag:            attacks.AttackTagExtra,
		ICDTag:               attacks.ICDTagNone,
		ICDGroup:             attacks.ICDGroupDefault,
		StrikeType:           attacks.StrikeTypePierce,
		Element:              attributes.Pyro,
		Durability:           25,
		Mult:                 fullaim[c.TalentLvlAttack()],
		HitWeakPoint:         weakspot == 1,
		HitlagHaltFrames:     0.12 * 60,
		HitlagFactor:         0.01,
		HitlagOnHeadshotOnly: true,
		IsDeployable:         true,
	}
	c2CB := c.makeC2CB()
	if hold < attacks.AimParamLv1 {
		ai.Abil = "Aimed Shot"
		ai.Element = attributes.Physical
		ai.Mult = aim[c.TalentLvlAttack()]
		c2CB = nil
	}
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			geometry.Point{Y: -0.5},
			0.1,
			1,
		),
		aimedHitmarks[hold],
		aimedHitmarks[hold]+travel,
		c2CB,
	)

	// 火焰矢
	if hold >= attacks.AimParamLv2 {
		ai.ICDTag = attacks.ICDTagExtraAttack
		ai.Mult = aimExtra[c.TalentLvlAttack()]

		// TODO:
		// 火焰矢は弱点に命中してPrototype Crescent等を発動できるが、常に会心するわけではない
		// 現在の仮定: 弱点には命中しない
		ai.HitWeakPoint = false

		// ヒットラグなし
		ai.HitlagHaltFrames = 0
		ai.HitlagFactor = 0.01
		ai.HitlagOnHeadshotOnly = false
		ai.IsDeployable = false

		for i := 1; i <= hold-1; i++ {
			ai.Abil = fmt.Sprintf("Kindling Arrow %v", i)
			// 火焰矢に少し余分な遅延を追加
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHit(
					c.Core.Combat.Player(),
					c.Core.Combat.PrimaryTarget(),
					nil,
					0.6,
				),
				aimedHitmarks[hold],
				aimedHitmarks[hold]+kindlingTravel,
				c2CB,
			)
		}
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(aimedFrames[hold]),
		AnimationLength: aimedFrames[hold][action.InvalidAction],
		CanQueueAfter:   aimedHitmarks[hold],
		State:           action.AimState,
	}, nil
}
