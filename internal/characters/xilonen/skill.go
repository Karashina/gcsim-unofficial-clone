package xilonen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	skillKey           = "Nightsoul's Blessing: Xilonen"
	SampleActiveDurKey = "xilonen-elemental-samples"
	SampleActiveKey    = "xilonen-elemental-samples-active"
	c6BypassKey        = "xilonen-c6-bypass"
	skillHitmark       = 8
)

var (
	skillFramesNormal []int
	skillFramesEnd    []int
)

func init() {
	skillFramesNormal = frames.InitAbilSlice(23)
	skillFramesEnd = frames.InitAbilSlice(5)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if !c.StatusIsActive(skillKey) {
		return c.skillActivate(), nil
	}
	return c.skillDeactivate(), nil
}

func (c *char) skillActivate() action.Info {
	c.Core.Tasks.Add(c.skillStartRoutine, 3)

	// Initial Skill Damage
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Yohual's Scratch: Rush DMG",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
		UseDef:     true,
		Alignment:  attacks.AdditionalTagNightsoul,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1),
		skillHitmark,
		skillHitmark,
		c.particleCB,
	)

	if c.Base.Cons >= 4 {
		for _, char := range c.Core.Player.Chars() {
			char.SetTag(c4buffkey, 6)
			char.AddStatus(c4buffkey, 15*60, true)
		}
	}

	// Return ActionInfo
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFramesNormal),
		AnimationLength: skillFramesNormal[action.InvalidAction],
		CanQueueAfter:   skillFramesNormal[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillStartRoutine() {
	c.AddNightsoul("xilonen-skill-init", 45)
	if c.Base.Cons < 1 {
		c.AddStatus(skillKey, 540, true)
	} else {
		c.AddStatus(skillKey, 780, true)
	}
	c.OnNightsoul = true
	c.Activatesample("geo")
	c.Core.Tasks.Add(c.depleteNightsoulPoints, 12)
}

func (c *char) skillDeactivate() action.Info {
	c.skillEndRoutine()

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFramesEnd),
		AnimationLength: 5,
		CanQueueAfter:   5,
		State:           action.Idle,
	}
}

func (c *char) skillEndRoutine() {
	if c.StatusIsActive(skillKey) {
		c.DeleteStatus(skillKey)
		c.Core.Log.NewEvent("Skill Deactivated by Nightsoul Depletion", glog.LogCharacterEvent, c.Index)
	}

	c.Core.Player.SwapCD = 5
	c.NightsoulPoint = 0
	c.OnNightsoul = false
	c.SetCD(action.ActionSkill, 7*60)
}

func (c *char) depleteNightsoulPoints() {
	if !c.StatusIsActive(skillKey) {
		return
	}
	if !c.StatusIsActive(c6buffKey) {
		if c.Base.Cons < 1 {
			c.ConsumeNightsoul(1)
		} else {
			c.ConsumeNightsoul(0.7)
		}
	} else {
		c.Core.Log.NewEvent("Skill Nightsoul consumption Bypassed by C6", glog.LogCharacterEvent, c.Index)
	}
	c.Core.Tasks.Add(c.depleteNightsoulPoints, 12)
}

func (c *char) NightsoulWatcher() {
	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		if c.OnNightsoul && !c.StatusIsActive(skillKey) {
			c.Core.Log.NewEvent("Skill Duration Ended", glog.LogCharacterEvent, c.Index)
			c.skillEndRoutine()
		}
		if c.StatusIsActive(SampleActiveKey) && !c.StatusIsActive(SampleActiveDurKey) {
			c.Core.Log.NewEvent("Sample Duration Ended", glog.LogCharacterEvent, c.Index)
			c.ResetSoundscapes()
			c.DeleteStatus(SampleActiveKey)
		}
		return false
	}, "xilonen-nightsoul-tick")

	c.Core.Events.Subscribe(event.OnNightsoulConsume, func(args ...interface{}) bool {
		idx := args[0].(int)
		if idx != c.Index {
			return false
		}
		if c.NightsoulPoint <= 0 {
			c.skillEndRoutine()
			c.Core.Log.NewEvent("Nightsoul Depleted", glog.LogCharacterEvent, c.Index)
		}
		return false
	}, "xilonen-nightsoul-watcher-consume")

	c.Core.Events.Subscribe(event.OnNightsoulGenerate, func(args ...interface{}) bool {
		idx := args[0].(int)
		if idx != c.Index {
			return false
		}
		if c.NightsoulPoint >= c.NightsoulPointMax {
			if !c.StatusIsActive(c6BypassKey) {
				c.Core.Log.NewEvent("Nightsoul Point Reaced Max", glog.LogCharacterEvent, c.Index)
				c.a4AdditionalNsB()
				c.Activatesample("ele")
			}
			if !c.StatusIsActive(c6buffKey) {
				c.ConsumeNightsoul(90)
			} else {
				c.AddStatus(c6BypassKey, c.StatusDuration(c6buffKey), true)
				c.Core.Log.NewEvent("Max Nightsoul Check Bypassed by C6", glog.LogCharacterEvent, c.Index)
			}
		}
		return false
	}, "xilonen-nightsoul-watcher-generate")
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Geo, c.ParticleDelay)
}

func (c *char) Activatesample(mode string) {

	if c.Base.Cons >= 2 && mode == "geo" {
		return // because Geo sample is always active
	}

	switch mode {
	case "geo":
		i := 0
		for _, e := range c.SoundScapeSlot {
			if e == attributes.Geo {
				c.ActivateRES(attributes.Geo)
				c.isSlotActive[i] = true
				c.Core.Log.NewEvent("Xilonen Geo Sample Activation from Skill init", glog.LogCharacterEvent, c.Index).
					Write("activated slot", i).
					Write("activated element", e)
			}
			i++
		}
	case "ele":
		i := 0
		for _, e := range c.SoundScapeSlot {
			c.ActivateRES(e)
			c.isSlotActive[i] = true
			c.Core.Log.NewEvent("Xilonen Elem Sample Activation from Nightsoul Points", glog.LogCharacterEvent, c.Index).
				Write("activated slot", i).
				Write("activated element", e)
			c.AddStatus(SampleActiveDurKey, 15*60, true)
			c.AddStatus(SampleActiveKey, -1, false)
			i++
		}
	}
}

func (c *char) ActivateRES(element attributes.Element) {

	if c.Base.Cons >= 2 && element == attributes.Geo {
		return // because RES buff is always active
	}

	// stole code from zhongli but 900f(15s) duration
	for i := 0; i <= 900; i += 18 {
		c.Core.Tasks.Add(func() {
			enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7.5), nil)
			key := fmt.Sprintf("xilonen-%v", element.String())
			for _, e := range enemies {
				e.AddResistMod(combat.ResistMod{
					Base:  modifier.NewBaseWithHitlag(key, 60),
					Ele:   element,
					Value: skillRes[c.TalentLvlSkill()] * -1,
				})
			}
		}, i)
	}
}
