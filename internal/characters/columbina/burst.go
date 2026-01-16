package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

// TODO: Frame data not measured, using stub values
const (
	burstHitmark      = 999
	lunarDomainDur    = 20 * 60 // 20 seconds
	lunarDomainKey    = "lunar-domain"
	lunarDomainModKey = "lunar-domain-bonus"
)

func init() {
	burstFrames = frames.InitAbilSlice(999)
	burstFrames[action.ActionAttack] = 999
	burstFrames[action.ActionCharge] = 999
	burstFrames[action.ActionSkill] = 999
	burstFrames[action.ActionDash] = 999
	burstFrames[action.ActionJump] = 999
	burstFrames[action.ActionSwap] = 999
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// Initial burst damage (AoE Hydro DMG)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Moonlit Melancholy",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 50,
	}
	ai.FlatDmg = c.MaxHP() * burstDmg[c.TalentLvlBurst()]

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.Core.QueueAttack(ai, ap, burstHitmark, burstHitmark)

	// Activate Lunar Domain
	c.AddStatus(lunarDomainKey, lunarDomainDur, true)
	c.lunarDomainSrc = c.Core.F
	c.lunarDomainActive = true

	// Apply Lunar Domain buff to all party members
	c.applyLunarDomainBuff()

	c.Core.Log.NewEvent("Lunar Domain activated", glog.LogCharacterEvent, c.Index).
		Write("duration", lunarDomainDur).
		Write("bonus", burstBonus[c.TalentLvlBurst()])

	// Schedule cleanup
	c.Core.Tasks.Add(func() {
		if c.lunarDomainSrc != c.Core.F-lunarDomainDur {
			return
		}
		c.lunarDomainActive = false
	}, lunarDomainDur)

	// Energy and cooldown
	c.ConsumeEnergy(5)
	c.SetCDWithDelay(action.ActionBurst, 15*60, 2)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// applyLunarDomainBuff applies Lunar Reaction DMG Bonus to all party members
func (c *char) applyLunarDomainBuff() {
	bonus := burstBonus[c.TalentLvlBurst()]
	dur := lunarDomainDur

	for _, char := range c.Core.Player.Chars() {
		// Add Lunar reaction bonus mods
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarDomainModKey+"-lc", dur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLCDamage {
					return bonus, false
				}
				return 0, false
			},
		})
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarDomainModKey+"-lb", dur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLBDamage {
					return bonus, false
				}
				return 0, false
			},
		})
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarDomainModKey+"-lcrs", dur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLCrsDamage {
					return bonus, false
				}
				return 0, false
			},
		})
	}
}

// subscribeToBurst subscribes to events for Lunar Domain interactions
func (c *char) subscribeToBurst() {
	// Subscribe to character switch to maintain Lunar Domain buff
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		next := args[1].(int)

		if !c.StatusIsActive(lunarDomainKey) {
			return false
		}

		// Re-apply buff to new active character if needed
		if prev != next {
			c.Core.Log.NewEvent("Lunar Domain buff carried", glog.LogCharacterEvent, c.Index).
				Write("from", c.Core.Player.Chars()[prev].Base.Key).
				Write("to", c.Core.Player.Chars()[next].Base.Key)
		}

		return false
	}, "columbina-lunar-domain-swap")
}

// isLunarDomainActive returns whether Lunar Domain is currently active
func (c *char) isLunarDomainActive() bool {
	return c.StatusIsActive(lunarDomainKey) && c.lunarDomainActive
}
