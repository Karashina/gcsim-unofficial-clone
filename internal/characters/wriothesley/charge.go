package wriothesley

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
	chargeHitmark = 19
)

func init() {
	chargeFrames = frames.InitAbilSlice(52) // CA -> N1/E/Q
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionWalk] = 51
	chargeFrames[action.ActionSwap] = 49
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Charge Attack",
		AttackTag:        attacks.AttackTagExtra,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		PoiseDMG:         110,
		Element:          attributes.Cryo,
		Durability:       25,
		Mult:             charge[c.TalentLvlAttack()],
		HitlagFactor:     0.01,
		HitlagHaltFrames: 0.09 * 60,
	}

	// TODO: スナップショットのタイミング
	snap := c.Snapshot(&ai)
	var ap combat.AttackPattern
	var rebukeCB combat.AttackCBFunc
	var particleCB combat.AttackCBFunc
	var c6Attack bool
	if c.Base.Ascension >= 1 {
		if c.Base.Cons >= 1 {
			rebukeCB, c6Attack = c.c1(&ai, &snap)
		} else {
			rebukeCB = c.a1(&ai, &snap)
		}

		if rebukeCB != nil {
			particleCB = c.particleCB
			ap = combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.8}, 4, 5)
		} else {
			ap = combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.2}, 2.8, 3.6)
		}
	}

	c.Core.QueueAttackWithSnap(ai, snap, ap, chargeHitmark, rebukeCB, particleCB)
	// 発動時、「重裁の嘃き：飛蹴」の基礎ダメージの100%のダメージを与える氷柱も発射する。
	// このダメージは重撃ダメージとみなされる。
	// 固有天賦「司罪の傅執」を先に解放する必要がある。
	if c6Attack {
		ai.Abil += " (C6)"
		ai.StrikeType = attacks.StrikeTypeDefault
		ai.PoiseDMG = 50
		c.Core.QueueAttackWithSnap(ai, snap, ap, chargeHitmark, rebukeCB, particleCB)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}
