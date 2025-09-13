package lauma

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var burstFrames []int

const (
	buffapply     = 104
	paleHymnKey   = "lauma-pale-hymn"
	paleHymnDuration = 30 * 60 // 30 seconds
)

func init() {
	burstFrames = frames.InitAbilSlice(116)
}

// Burst
// gain 18 stacks of Pale Hymn.
// Additionally, if Lauma uses her Elemental Burst while she has Moon Song,or she gains Moon Song within 15s of using her Elemental Burst,
// she will consume all Moon Song stacks and gain 6 stacks of Pale Hymn for every Moon Song stack consumed.
// This effect can only be triggered once for each Elemental Burst used, including the 15 seconds following its use.
// Pale Hymn
// When nearby party members deal Bloom, Hyperbloom, Burgeon, or Lunar-Bloom DMG,
// 1 stack of Pale Hymn will be consumed and the DMG dealt will be increased based on Lauma's Elemental Mastery.
// If this DMG hits multiple opponents at once, then multiple stacks of Pale Hymn will be consumed, depending on how many opponents are hit.
// The duration for each stack of Pale Hymn is counted independently.
// if Lauma is C2 or higher,
// Pale Hymn effects are increased: All nearby party members' Bloom, Hyperbloom, and Burgeon DMG is further increased by 500% of Lauma's Elemental Mastery,
// and their Lunar-Bloom DMG is further increased by 400% of Lauma's Elemental Mastery.
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// Initial 18 Pale Hymn stacks
	initialStacks := 18
	
	// Check for Moon Song conversion
	bonusStacks := 0
	if c.moonSong > 0 {
		bonusStacks = c.moonSong * 6
		c.moonSong = 0
		c.AddStatus("moon-song-consumed", -1, false) // Mark that Moon Song was consumed for this burst
	}
	
	totalStacks := initialStacks + bonusStacks
	c.paleHymn += totalStacks
	
	// Set up Pale Hymn effect monitoring
	c.Core.Tasks.Add(func() {
		c.setupPaleHymnEffects()
	}, buffapply)
	
	// Set up Moon Song monitoring for 15s window
	c.AddStatus("moon-song-window", 15*60, true)
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		if c.StatusIsActive("moon-song-window") && !c.StatusIsActive("moon-song-consumed") {
			if c.moonSong > 0 {
				bonusStacks := c.moonSong * 6
				c.moonSong = 0
				c.paleHymn += bonusStacks
				c.AddStatus("moon-song-consumed", -1, false)
				return true
			}
		}
		return false
	}, "lauma-burst-moon-song")

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(7)
	c.c1() // C1 effect on burst use
	c.c2() // C2 effect check

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

// Setup Pale Hymn reaction damage bonuses
func (c *char) setupPaleHymnEffects() {
	// Subscribe to Bloom/Hyperbloom/Burgeon reactions
	c.Core.Events.Subscribe(event.OnBloom, c.paleHymnReactionBonus("Bloom"), "lauma-pale-hymn-bloom")
	c.Core.Events.Subscribe(event.OnHyperbloom, c.paleHymnReactionBonus("Hyperbloom"), "lauma-pale-hymn-hyperbloom")
	c.Core.Events.Subscribe(event.OnBurgeon, c.paleHymnReactionBonus("Burgeon"), "lauma-pale-hymn-burgeon")
	
	// Subscribe to Lunar-Bloom reactions
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if len(args) < 3 {
			return false
		}
		if attackEvent, ok := args[1].(*combat.AttackEvent); ok {
			if attackEvent.Info.AttackTag == attacks.AttackTagLBDamage {
				return c.paleHymnLunarBloomBonus(args...)
			}
		}
		return false
	}, "lauma-pale-hymn-lunar-bloom")
}

// Pale Hymn reaction bonus for Bloom/Hyperbloom/Burgeon
func (c *char) paleHymnReactionBonus(reactionType string) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		if c.paleHymn <= 0 {
			return false
		}
		
		// Consume 1 stack per enemy hit (need to calculate number of targets)
		enemiesHit := 1 // Default to 1, would need target counting logic for precise implementation
		stacksConsumed := min(c.paleHymn, enemiesHit)
		c.paleHymn -= stacksConsumed
		
		// Add damage bonus based on EM
		em := c.Stat(attributes.EM)
		bonusDamage := burstBuffBloom[c.TalentLvlBurst()] * em
		
		// C2 additional bonus
		if c.Base.Cons >= 2 {
			bonusDamage += 5.0 * em // 500% of EM
		}
		
		// Apply the bonus damage (this would need to be integrated with the reaction system)
		// For now, we'll add it as a modifier to the character dealing the reaction
		if len(args) > 0 {
			if char, ok := args[0].(*character.CharWrapper); ok {
				char.AddReactBonusMod(character.ReactBonusMod{
					Base: modifier.NewBase("lauma-pale-hymn-"+reactionType, 1),
					Amount: func(ai combat.AttackInfo) (float64, bool) {
						return bonusDamage, false
					},
				})
			}
		}
		
		return true
	}
}

// Pale Hymn bonus for Lunar-Bloom damage
func (c *char) paleHymnLunarBloomBonus(args ...interface{}) bool {
	if c.paleHymn <= 0 {
		return false
	}
	
	// Consume 1 stack per enemy hit
	enemiesHit := 1 // Default to 1
	stacksConsumed := min(c.paleHymn, enemiesHit)
	c.paleHymn -= stacksConsumed
	
	// Add damage bonus based on EM
	em := c.Stat(attributes.EM)
	bonusDamage := burstBuffLBloom[c.TalentLvlBurst()] * em
	
	// C2 additional bonus
	if c.Base.Cons >= 2 {
		bonusDamage += 4.0 * em // 400% of EM
	}
	
	// Store the bonus for Lunar-Bloom damage
	c.burstLBBuff = bonusDamage
	
	return true
}
