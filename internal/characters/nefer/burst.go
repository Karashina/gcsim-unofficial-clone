package nefer

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var (
	burstFrames []int
)

const (
	burstHitmark1 = 103
	burstHitmark2 = 149
)

func init() {
	burstFrames = frames.InitAbilSlice(122)
}

// Elemental Burst: two AoE Dendro hits; consumes Veil stacks to buff damage.
func (c *char) Burst(p map[string]int) (action.Info, error) {

	// 1st Hit
	c.QueueCharTask(func() {
		ai1atk := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Sacred Vow: True Eye's Phantasm / 1-Hit DMG (Q)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			FlatDmg:    burst1atk[c.TalentLvlBurst()]*c.Stat(attributes.ATK) + burst1em[c.TalentLvlBurst()]*c.Stat(attributes.EM),
		}
		c.Core.QueueAttack(
			ai1atk,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 7),
			0,
			0,
		)
	}, burstHitmark1)

	// 2nd Hit - ATK portion
	c.QueueCharTask(func() {
		ai2atk := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Sacred Vow: True Eye's Phantasm / 2-Hit DMG (Q)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			FlatDmg:    burst2atk[c.TalentLvlBurst()]*c.Stat(attributes.ATK) + burst2em[c.TalentLvlBurst()]*c.Stat(attributes.EM),
		}
		c.Core.QueueAttack(
			ai2atk,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 7),
			0,
			0,
		)
		// Consume Veil of Falsehood stacks
		c.a1count = 0
	}, burstHitmark2)

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

// makeBurstBonus consumes Veil stacks to increase current Burst DMG.
func (c *char) makeBurstBonus() {
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("nefer-q-dmgbuff", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// skip if not burst
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			// apply buff
			m[attributes.DmgP] = burstbonus[c.TalentLvlBurst()] * c.a1count
			return m, true
		},
	})
}
