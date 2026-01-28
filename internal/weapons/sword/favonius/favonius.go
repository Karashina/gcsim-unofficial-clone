package favonius

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/weapons/common"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.FavoniusSword, NewWeapon)
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	f := common.NewFavonius(base)
	return f.NewWeapon(c, char, p)
}
