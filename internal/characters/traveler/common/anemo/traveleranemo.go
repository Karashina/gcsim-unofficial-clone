package anemo

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/characters/traveler/common"
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

type Traveler struct {
	*tmpl.Character
	qAbsorb              attributes.Element
	qICDTag              attacks.ICDTag
	qAbsorbCheckLocation combat.AttackPattern
	eAbsorb              attributes.Element
	eICDTag              attacks.ICDTag
	eAbsorbCheckLocation combat.AttackPattern
	gender               int
}

func NewTraveler(s *core.Core, w *character.CharWrapper, p info.CharacterProfile, gender int) (*Traveler, error) {
	c := Traveler{
		gender: gender,
	}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.Base.Element = attributes.Anemo
	c.EnergyMax = 60
	c.BurstCon = 3
	c.SkillCon = 5
	c.NormalHitNum = normalHitNum

	common.TravelerStoryBuffs(w, p)
	return &c, nil
}

func (c *Traveler) Init() error {
	c.a4()
	if c.Base.Cons >= 2 {
		c.c2()
	}
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
