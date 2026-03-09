package hydro

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/template/sourcewaterdroplet"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

const a1ICDKey = "sourcewater-droplet-icd"

// 水紋の剣の長押しモードで発射された水珠が敵に命中すると、旅人の近くに源水の雫が生成される。
// 旅人が拾うとHPを7%回復する。
// この方法では毎秒1個まで生成可能で、水紋の剣の使用ごとに最大4個まで生成できる。
func (c *Traveler) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	count := 0
	return func(a combat.AttackCB) {
		if count >= 4 {
			return
		}
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(a1ICDKey) {
			return
		}

		count++
		droplet := c.newDroplet()
		c.Core.Combat.AddGadget(droplet)
		c.AddStatus(a1ICDKey, 60, true)
	}
}

func (c *Traveler) a1PickUp(count int) {
	for _, g := range c.Core.Combat.Gadgets() {
		if count == 0 {
			return
		}

		droplet, ok := g.(*sourcewaterdroplet.Gadget)
		if !ok {
			continue
		}
		droplet.Kill()
		count--

		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "Spotless Waters",
			Src:     c.MaxHP() * 0.07,
			Bonus:   c.Stat(attributes.Heal),
		})

		// 源水の雫を拾うと旅人のエネルギーが2回復する。
		// 固有天賦「Spotless Waters」が必要。
		if c.Base.Cons >= 1 {
			c.AddEnergy("travelerhydro-c1", 2)
		}

		if c.Base.Cons >= 6 {
			c.c6()
		}
	}
}

func (c *Traveler) newDroplet() *sourcewaterdroplet.Gadget {
	player := c.Core.Combat.Player()
	pos := geometry.CalcRandomPointFromCenter(
		geometry.CalcOffsetPoint(
			player.Pos(),
			geometry.Point{Y: 3.5},
			player.Direction(),
		),
		0.3,
		3,
		c.Core.Rand,
	)
	droplet := sourcewaterdroplet.New(c.Core, pos, combat.GadgetTypSourcewaterDropletHydroTrav)
	return droplet
}
