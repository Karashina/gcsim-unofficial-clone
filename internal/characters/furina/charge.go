package furina

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var (
	chargeFrames  []int
	chargeHitmark = 33
)

func init() {
	chargeFrames = frames.InitAbilSlice(253) // C -> Walk
	chargeFrames[action.ActionAttack] = 37
	chargeFrames[action.ActionSkill] = 47
	chargeFrames[action.ActionBurst] = chargeHitmark
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = chargeHitmark
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	windup := 0
	if c.Core.Player.CurrentState() == action.NormalAttackState {
		windup = 11
	}
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             fmt.Sprintf("Charge %v", c.arkhe),
		AttackTag:        attacks.AttackTagExtra,
		ICDTag:           attacks.ICDTagExtraAttack,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeSlash,
		Element:          attributes.Physical,
		HitlagHaltFrames: 0.02 * 60,
		Durability:       25,
		Mult:             charge[c.TalentLvlAttack()],
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.6)
	c.QueueCharTask(func() {
		var c6cb combat.AttackCBFunc
		// TODO: 重撃開始からヒットまでの間に6凸が切れた場合、ダメージボーナスが適用されるか要確認
		if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
			ai.FlatDmg = c.c6BonusDMG()
			c6cb = c.c6cb
			ai.Element = attributes.Hydro
			ai.IgnoreInfusion = true
		}
		c.Core.QueueAttack(ai, ap, 0, 0, c6cb)
	}, chargeHitmark-windup)

	arkheChangeFunc := func() {
		c.arkhe = pneuma
		if c.StatusIsActive(skillKey) {
			c.summonSinger(0)
		}
	}
	if c.arkhe == pneuma {
		arkheChangeFunc = func() {
			c.arkhe = ousia
			if c.StatusIsActive(skillKey) {
				c.summonSalonMembers(0)
			}
		}
	}

	// +1してc6のダメージボーナス/回復ステータスが正しく評価されるようにする
	c.QueueCharTask(arkheChangeFunc, chargeHitmark-windup+1)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeHitmark - windup,
		State:           action.ChargeAttackState,
	}, nil
}
