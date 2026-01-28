package aether

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/characters/traveler/common/pyro"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

type char struct {
	*pyro.Traveler
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	t, err := pyro.NewTraveler(s, w, p, 0)
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
	core.RegisterCharFunc(keys.AetherPyro, NewChar)
}
