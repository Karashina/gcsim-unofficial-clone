package xilonen

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Xilonen, NewChar)
}

type char struct {
	*tmpl.Character
	NormalSCounter int
	SoundScapeSlot []attributes.Element
	isSlotActive   []bool
	A1Mode         int
	a1buff         []float64
	a4buff         []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.CharZone = info.ZoneNatlan
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.NightsoulPointMax = 90
	c.NightsoulPoint = 0
	c.HasNightsoul = true
	c.OnNightsoul = false

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.SoundScapeSlot = make([]attributes.Element, 3)
	c.isSlotActive = make([]bool, 3)
	c.CheckSoundscapes()
	c.NightsoulWatcher()
	c.a1()
	c.a4()
	if c.Base.Cons >= 2 {
		c.c2()
	}
	if c.Base.Cons >= 4 {
		c.c4()
	}
	if c.Base.Cons >= 6 {
		c.c6()
	}
	c.onExitField()
	return nil
}

func (c *char) AdvanceNormalIndex() {
	if c.StatusIsActive(skillKey) {
		c.NormalSCounter++
		if c.NormalSCounter == 4 {
			c.NormalSCounter = 0
		}
		return
	}
	c.NormalCounter++
	if c.NormalCounter == c.NormalHitNum {
		c.NormalCounter = 0
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul-points":
		if c.NightsoulPoint <= 0 {
			return 0, nil
		}
		return c.NightsoulPoint, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatModIsActive(skillKey) {
			c.skillEndRoutine()
		}
		c.Core.Log.NewEvent("Xilonen onExitField", glog.LogCharacterEvent, c.Index)
		return false
	}, "xilonen-exit")
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		if c.StatusIsActive(skillKey) {
			return 12
		}
		return 0
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

func (c *char) CheckSoundscapes() {
	chars := c.Core.Player.Chars()
	i := 0
	a1count := 0
	for _, this := range chars {
		if this.Index == c.Index {
			continue
		}
		switch this.Base.Element {
		case attributes.Pyro:
			c.SoundScapeSlot[i] = attributes.Pyro
			a1count++
		case attributes.Hydro:
			c.SoundScapeSlot[i] = attributes.Hydro
			a1count++
		case attributes.Cryo:
			c.SoundScapeSlot[i] = attributes.Cryo
			a1count++
		case attributes.Electro:
			c.SoundScapeSlot[i] = attributes.Electro
			a1count++
		default:
			c.SoundScapeSlot[i] = attributes.Geo
		}
		c.isSlotActive[i] = false
		i++
	}
	if a1count < 2 {
		c.A1Mode = 1
		c.Core.Log.NewEvent("Sample A1 is on mode 1", glog.LogCharacterEvent, c.Index)
	} else {
		c.A1Mode = 2
		c.Core.Log.NewEvent("Sample A1 is on mode 2", glog.LogCharacterEvent, c.Index)
	}
	c.Core.Log.NewEvent("Xilonen Samples Init", glog.LogCharacterEvent, c.Index).
		Write("Slot-0", c.SoundScapeSlot[0]).
		Write("Slot-1", c.SoundScapeSlot[1]).
		Write("Slot-2", c.SoundScapeSlot[2])
}

func (c *char) ResetSoundscapes() {
	for i := 0; i < 3; i++ {
		if c.SoundScapeSlot[i] == attributes.Geo && c.Base.Cons < 2 {
			c.isSlotActive[i] = false
		} else {
			c.isSlotActive[i] = false
		}
	}
	c.Core.Log.NewEvent("Xilonen Samples Reset", glog.LogCharacterEvent, c.Index)
}
