package lumine

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/characters/traveler/common/geo"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

type char struct {
	*geo.Traveler
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	t, err := geo.NewTraveler(s, w, p, 1)
	if err != nil {
		return err
	}
	c := &char{
		Traveler: t,
	}
	w.Character = c

	return nil
}

func init() {
	core.RegisterCharFunc(keys.LumineGeo, NewChar)
}
