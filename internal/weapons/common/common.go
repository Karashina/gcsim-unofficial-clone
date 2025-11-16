package common

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

type NoEffect struct {
	Index int
	data  *model.WeaponData
}

func (n *NoEffect) SetIndex(idx int)        { n.Index = idx }
func (n *NoEffect) Init() error             { return nil }
func (n *NoEffect) Data() *model.WeaponData { return n.data }

func NewNoEffect(data *model.WeaponData) *NoEffect {
	return &NoEffect{data: data}
}

func (n *NoEffect) NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return n, nil
}

