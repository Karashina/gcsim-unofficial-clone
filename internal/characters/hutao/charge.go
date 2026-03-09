package hutao

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var chargeFrames []int
var ppChargeFrames []int

const (
	chargeHitmark   = 19
	ppChargeHitmark = 3
)

func init() {
	// charge -> x
	chargeFrames = frames.InitAbilSlice(62)
	chargeFrames[action.ActionAttack] = 57
	chargeFrames[action.ActionSkill] = 57
	chargeFrames[action.ActionSkill] = 60
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark

	// charge (paramita) -> x
	ppChargeFrames = frames.InitAbilSlice(42)
	ppChargeFrames[action.ActionBurst] = 33
	ppChargeFrames[action.ActionDash] = ppChargeHitmark
	ppChargeFrames[action.ActionJump] = ppChargeHitmark
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.StatModIsActive(paramitaBuff) {
		return c.ppChargeAttack(), nil
	}

	// 粒子生成をチェック
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charge Attack",
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
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) ppChargeAttack() action.Info {
	// PPスライド: 重撃開始時に paramita に1.8秒加算し、重撃終了時に削除される
	c.ExtendStatus(paramitaBuff, 1.8*60)

	//TODO: 現在は弾丸なのでキャスト時スナップショットと仮定。"PPスライド"は未実装
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charge Attack",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupPoleExtraAttack,
		StrikeType:         attacks.StrikeTypeSlash,
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
		ppChargeHitmark,
		c.particleCB,
		c.applyBB,
	)

	// 前のアクションが通常攻撃の場合フレームが変わる
	prevState := -1
	if c.Core.Player.LastAction.Char == c.Index && c.Core.Player.LastAction.Type == action.ActionAttack {
		prevState = c.NormalCounter - 1
		if prevState < 0 {
			prevState = c.NormalHitNum - 1
		}
	}
	ff := func(next action.Action) int {
		if prevState == -1 {
			return ppChargeFrames[next]
		}
		switch next {
		case action.ActionDash, action.ActionJump:
		default:
			return ppChargeFrames[next]
		}
		switch prevState {
		case 0: // N1
			if next == action.ActionDash {
				return 1 // N1D
			}
			return 2 // N1J
		case 1: // N2
			if next == action.ActionDash {
				return 4 // N2D
			}
			return 5 // N2J
		case 2: // N3
			return 2
		case 3: // N4
			return 3
		case 4: // N5
			return 3
		default:
			return 500 //TODO: このアクションは無効。より良いハンドリングが必要
		}
	}

	return action.Info{
		Frames:          ff,
		AnimationLength: ppChargeFrames[action.InvalidAction],
		CanQueueAfter:   1,
		State:           action.ChargeAttackState,
		OnRemoved: func(next action.AnimationState) {
			if next != action.BurstState {
				c.ExtendStatus(paramitaBuff, -1.8*60)
			}
		},
	}
}
