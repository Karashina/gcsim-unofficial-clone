package dendro

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/characters/traveler/common"
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

type Traveler struct {
	*tmpl.Character
	burstPos                   geometry.Point
	burstRadius                float64
	burstOverflowingLotuslight int
	skillC1                    bool // this variable also ensures that C1 only restores energy once per cast
	burstTransfig              attributes.Element
	gender                     int
}

func NewTraveler(s *core.Core, w *character.CharWrapper, p info.CharacterProfile, gender int) (*Traveler, error) {
	c := Traveler{
		gender: gender,
	}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.Base.Element = attributes.Dendro
	c.EnergyMax = 80
	c.BurstCon = 5
	c.SkillCon = 3
	c.NormalHitNum = normalHitNum

	common.TravelerStoryBuffs(w, p)
	return &c, nil
}

func (c *Traveler) Init() error {
	c.a1Init()
	c.a4Init()

	if c.Base.Cons >= 6 {
		c.c6Init()
	}

	c.skillC1 = false
	return nil
}

func (c *Traveler) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		if c.gender == 0 {
			return 8
		}
		return 7
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
