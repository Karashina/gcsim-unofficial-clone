package lauma

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

var burstFrames []int

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
	}

	totalStacks := initialStacks + bonusStacks
	c.paleHymn += totalStacks
	c.QueueCharTask(func() {
		c.paleHymn -= min(totalStacks, c.paleHymn) // Remove stacks after 15s
	}, 15*60)

	// Set up Pale Hymn monitoring for 15s window
	c.AddStatus("pale-hymn-window", 15*60, true)

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

	// Subscribe to Bloom reactions
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		switch ae.Info.AttackTag {
		case attacks.AttackTagBloom, attacks.AttackTagHyperbloom, attacks.AttackTagBurgeon, attacks.AttackTagBountifulCore:
			return c.paleHymnReactionBonus(ae)
		case attacks.AttackTagLBDamage:
			return c.paleHymnLunarBloomBonus(ae)
		}
		return false
	}, "lauma-pale-hymn-lunar-bloom")
}

// Pale Hymn reaction bonus for Bloom/Hyperbloom/Burgeon
func (c *char) paleHymnReactionBonus(ae *combat.AttackEvent) bool {
	if !c.StatusIsActive("pale-hymn-window") {
		return false
	}
	if c.paleHymn <= 0 {
		return false
	}

	targets := len(c.Core.Combat.EnemiesWithinArea(ae.Pattern, nil))
	// Consume 1 stack per enemy hit (need to calculate number of targets)
	enemiesHit := targets
	stacksConsumed := min(c.paleHymn, enemiesHit)
	c.paleHymn -= stacksConsumed

	// Add damage bonus based on EM
	em := c.Stat(attributes.EM)
	bonusDamage := burstBuffBloom[c.TalentLvlBurst()] * em

	// C2 additional bonus
	if c.Base.Cons >= 2 {
		bonusDamage += 5.0 * em // 500% of EM
	}

	ae.Info.FlatDmg += bonusDamage
	ae.Info.FlatDmg *= c.c6mult

	return false
}

// Pale Hymn bonus for Lunar-Bloom damage
func (c *char) paleHymnLunarBloomBonus(ae *combat.AttackEvent) bool {
	if !c.StatusIsActive("pale-hymn-window") {
		return false
	}
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
	ae.Info.FlatDmg += bonusDamage
	return false
}

