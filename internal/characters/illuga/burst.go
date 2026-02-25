package illuga

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var burstFrames []int

const (
	burstHitmark          = 50
	orioleSongDur         = 1255 // 20s (I-3 fix: was 15s)
	baseNightingaleStacks = 21
	stacksPerGeoConstruct = 5
)

func init() {
	burstFrames = frames.InitAbilSlice(66)
}

// Burst performs Shadowless Reflection - Oriole-Song
// Enters Oriole-Song state; Nightingale's Song adds FlatDmg to party Geo hits
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// Calculate initial Nightingale stacks
	// Base 21 + 5 per Geo Construct (max 3 constructs)
	geoConstructs := c.countGeoConstructs()
	initialConstructStacks := geoConstructs * stacksPerGeoConstruct
	if initialConstructStacks > 15 {
		initialConstructStacks = 15
	}
	c.geoConstructBonusStacks = initialConstructStacks // Track initial construct stacks toward 15-cap
	c.nightingaleSongStacks = baseNightingaleStacks + initialConstructStacks

	// Initial burst hit
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Shadowless Reflection",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 50,
	}

	// EM + DEF scaling from burst_gen data
	em := c.Stat(attributes.EM)
	def := c.TotalDef(false)
	emMult := burstEM[c.TalentLvlBurst()]
	defMult := burstDEF[c.TalentLvlBurst()]

	ai.FlatDmg = em*emMult + def*defMult

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 8)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, burstHitmark)

	// Enter Oriole-Song state
	c.QueueCharTask(func() {
		c.enterOrioleSong()
	}, burstHitmark)

	// Apply A1 enhanced buffs (Ascendant Gleam check)
	c.applyLightkeeperOath()

	// I-4 fix: CD 18s → 15s
	c.SetCDWithDelay(action.ActionBurst, 15*60, burstHitmark)
	c.ConsumeEnergy(4)

	c.Core.Log.NewEvent("Illuga uses Shadowless Reflection", glog.LogCharacterEvent, c.Index).
		Write("nightingale_stacks", c.nightingaleSongStacks).
		Write("geo_constructs", geoConstructs).
		Write("flat_dmg", ai.FlatDmg)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// enterOrioleSong activates Oriole-Song state
func (c *char) enterOrioleSong() {
	c.orioleSongActive = true
	c.orioleSongSrc = c.Core.F
	c.AddStatus(orioleSongKey, orioleSongDur, true)

	// I-1/I-2 refactor: Nightingale's Song uses FlatDmg per hit via OnEnemyHit
	// instead of GeoP% StatMod. burstGeoBonusEM and burstLCrsBonusEM are now used.
	c.subscribeNightingaleStackConsumption()

	// I-GC1: Subscribe to Geo Construct spawns for dynamic stack gain
	c.subscribeGeoConstructStacks()

	// Schedule mode exit
	src := c.orioleSongSrc
	c.QueueCharTask(func() {
		if c.orioleSongSrc != src {
			return
		}
		c.exitOrioleSong()
	}, orioleSongDur)

	c.Core.Log.NewEvent("Illuga enters Oriole-Song", glog.LogCharacterEvent, c.Index).
		Write("nightingale_stacks", c.nightingaleSongStacks).
		Write("duration", orioleSongDur)
}

// subscribeNightingaleStackConsumption subscribes to party member Geo damage hits
// I-1: Nightingale's Song adds FlatDmg to each qualifying Geo hit, consuming 1 stack per hit
// I-13: Applies to all party members' Geo attacks
// I-14: Only Normal/Charged/Plunge/Skill/Burst attack tags qualify
func (c *char) subscribeNightingaleStackConsumption() {
	cb := func(args ...interface{}) bool {
		if !c.orioleSongActive {
			return false
		}
		if c.nightingaleSongStacks <= 0 {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		t := args[0].(combat.Target)

		// Only apply to hits against enemies
		if t.Type() != targets.TargettableEnemy {
			return false
		}

		// Only qualify Geo element hits
		if atk.Info.Element != attributes.Geo {
			return false
		}

		// I-14: Only Normal/Charged/Plunge/Skill/Burst attack tags qualify
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal,
			attacks.AttackTagExtra,
			attacks.AttackTagPlunge,
			attacks.AttackTagElementalArt,
			attacks.AttackTagElementalArtHold,
			attacks.AttackTagElementalBurst,
			attacks.AttackTagLCrsDamage:
			// valid
		default:
			return false
		}

		// I-1/I-2: Add FlatDmg based on Illuga's EM and talent level
		illugaEM := c.Stat(attributes.EM)
		var flatDmg float64
		if atk.Info.AttackTag == attacks.AttackTagLCrsDamage {
			// LCrs hits use burstLCrsBonusEM multiplier
			flatDmg = burstLCrsBonusEM[c.TalentLvlBurst()] * illugaEM
			flatDmg += c.getA4LCrsBonus() // A4 LCrs enhancement
		} else {
			// Geo hits use burstGeoBonusEM multiplier
			flatDmg = burstGeoBonusEM[c.TalentLvlBurst()] * illugaEM
			flatDmg += c.getA4GeoBonus() // A4 Geo enhancement
		}
		atk.Info.FlatDmg += flatDmg

		// Consume 1 stack per enemy hit
		c.nightingaleSongStacks--

		// C2: Check if enough stacks consumed for lamp attack
		if c.Base.Cons >= 2 {
			c.c2StackCounter++
			for c.c2StackCounter >= 7 {
				c.c2LampAttack()
				c.c2StackCounter -= 7
			}
		}

		c.Core.Log.NewEvent("Illuga Nightingale stack consumed", glog.LogCharacterEvent, c.Index).
			Write("remaining_stacks", c.nightingaleSongStacks).
			Write("flat_dmg_added", flatDmg)

		if c.nightingaleSongStacks <= 0 {
			c.nightingaleSongStacks = 0
			c.exitOrioleSong()
		}

		return false
	}

	// I-10: Use key-based subscribe to avoid duplicate registration on burst reuse
	c.Core.Events.Subscribe(event.OnEnemyHit, cb, "illuga-nightingale-consume")
}

// exitOrioleSong ends Oriole-Song state
func (c *char) exitOrioleSong() {
	c.orioleSongActive = false
	c.orioleSongSrc = -1
	c.DeleteStatus(orioleSongKey)
	c.nightingaleSongStacks = 0
	c.c2StackCounter = 0
	c.geoConstructBonusStacks = 0

	c.Core.Log.NewEvent("Illuga exits Oriole-Song", glog.LogCharacterEvent, c.Index)
}

// subscribeGeoConstructStacks subscribes to Geo Construct creation events
// When a Geo Construct is created during Oriole-Song, Illuga gains 5 stacks per
// construct currently on the battlefield, up to 15 additional stacks total.
func (c *char) subscribeGeoConstructStacks() {
	c.Core.Events.Subscribe(event.OnConstructSpawned, func(args ...interface{}) bool {
		if !c.orioleSongActive {
			return false
		}
		if c.geoConstructBonusStacks >= 15 {
			return false
		}

		constructCount := c.countGeoConstructs()
		gain := constructCount * stacksPerGeoConstruct
		remaining := 15 - c.geoConstructBonusStacks
		if gain > remaining {
			gain = remaining
		}

		c.nightingaleSongStacks += gain
		c.geoConstructBonusStacks += gain

		c.Core.Log.NewEvent("Illuga gains Nightingale stacks from Geo Construct", glog.LogCharacterEvent, c.Index).
			Write("gain", gain).
			Write("total_bonus_stacks", c.geoConstructBonusStacks).
			Write("nightingale_stacks", c.nightingaleSongStacks)

		return false
	}, "illuga-geo-construct-stacks")
}

// countGeoConstructs returns the number of active Geo constructs
func (c *char) countGeoConstructs() int {
	return c.Core.Constructs.Count()
}
