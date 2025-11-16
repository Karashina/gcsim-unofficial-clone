package talkingstick

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/weapons/common"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

// ATK will be increased by 16% for 15s after being affected by Pyro. This effect can be triggered once every 12s.
// All Elemental DMG Bonus will be increased by 12% for 15s after being affected by Hydro, Cryo, Electro, or Dendro.
// This effect can be triggered once every 12s.
// TODO: https://github.com/Karashina/gcsim-unofficial-clone/issues/850
func init() {
	core.RegisterWeaponFunc(keys.TalkingStick, NewWeapon)
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := common.NewNoEffect(base)
	return w.NewWeapon(c, char, p)
}

