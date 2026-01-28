package lanyan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

func (c *char) absorbA1() attributes.Element {
	if c.Base.Ascension < 1 {
		return attributes.Anemo
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4.0)
	ele := c.Core.Combat.AbsorbCheck(c.Index, ap, attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)
	if ele == attributes.NoElement {
		return attributes.Anemo
	}
	return ele
}
