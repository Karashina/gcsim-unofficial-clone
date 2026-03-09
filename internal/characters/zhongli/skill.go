package zhongli

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var skillPressFrames []int
var skillHoldFrames []int

const skillPressHimark = 24
const skillHoldHitmark = 48

func init() {
	// skill (press) -> x
	skillPressFrames = frames.InitAbilSlice(38)
	skillPressFrames[action.ActionAttack] = 37
	skillPressFrames[action.ActionBurst] = 38
	skillPressFrames[action.ActionDash] = 23
	skillPressFrames[action.ActionJump] = 23
	skillPressFrames[action.ActionSwap] = 37

	// skill (hold) -> x
	skillHoldFrames = frames.InitAbilSlice(96)
	skillHoldFrames[action.ActionDash] = 55
	skillHoldFrames[action.ActionJump] = 55
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	h := p["hold"]
	nostele := p["hold_nostele"] > 0
	if h > 0 || nostele {
		return c.skillHold(!nostele), nil
	}
	return c.skillPress(), nil
}

func (c *char) skillPress() action.Info {
	c.Core.Tasks.Add(func() {
		c.newStele(1860)
	}, skillPressHimark)

	c.SetCDWithDelay(action.ActionSkill, 240, 22)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) skillHold(createStele bool) action.Info {
	// 長押しはダメージを与える
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Stone Stele (Hold)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   142.9,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       skillHold[c.TalentLvlSkill()],
		FlatDmg:    c.a4Skill(),
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), 0, skillHoldHitmark)

	// 鍾離の最大岩柱数未満かつプレイヤーが望む場合に岩柱を生成
	if (c.steleCount < c.maxStele) && createStele {
		c.Core.Tasks.Add(func() {
			c.newStele(1860) // 31秒
		}, skillHoldHitmark)
	}

	// シールドを生成 - 敵のデバフ矢印はゲーム内でダメージ数値が表示された3-5フレーム後に出現
	c.Core.Tasks.Add(func() {
		c.addJadeShield()
	}, skillHoldHitmark)

	c.SetCDWithDelay(action.ActionSkill, 720, 47)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}
