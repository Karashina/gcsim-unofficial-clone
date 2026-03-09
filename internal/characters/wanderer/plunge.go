package wanderer

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var lowPlungeFrames []int

const lowPlungeHitmark = 41
const lowPlungeCollisionHitmark = 36

func init() {
	lowPlungeFrames = frames.InitAbilSlice(72)
	lowPlungeFrames[action.ActionAttack] = 65
	lowPlungeFrames[action.ActionCharge] = 64
	lowPlungeFrames[action.ActionBurst] = 65
	lowPlungeFrames[action.ActionDash] = 41
	lowPlungeFrames[action.ActionSwap] = 57
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	delay := c.checkForSkillEnd()

	// 落下状態ではない
	if !c.StatusIsActive(plungeAvailableKey) {
		return action.Info{}, errors.New("only plunge after skill ends")
	}
	c.DeleteStatus(plungeAvailableKey)

	// 空中から発動するため遅延を短縮
	if delay > 0 {
		delay = 7
	}

	collision, ok := p["collision"]
	if !ok {
		collision = 0 // 放浪者が衝突ヒットを行うかどうか
	}

	if collision > 0 {
		c.plungeCollision(lowPlungeCollisionHitmark + delay)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Low Plunge Attack",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       lowPlunge[c.TalentLvlAttack()],
	}

	// TODO: スナップショット遅延を確認
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3),
		delay+lowPlungeHitmark, delay+lowPlungeHitmark)

	return action.Info{
		Frames:          func(next action.Action) int { return lowPlungeFrames[next] },
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeHitmark,
		State:           action.PlungeAttackState,
	}, nil
}

func (c *char) plungeCollision(fullDelay int) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Plunge Collision",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 0,
		Mult:       plunge[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.5), fullDelay, fullDelay)
}
