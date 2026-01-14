package ayaka

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

type ReactableWeapon struct {
	*gadget.Gadget
	*reactable.Reactable
	c *char
}

func (c *char) newReactableWeapons() *ReactableWeapon {
	p := &ReactableWeapon{}
	p.Reactable = &reactable.Reactable{}
	p.Gadget = gadget.New(c.Core, c.Core.Combat.Player().Pos(), 0.5, combat.GadgetTypReactableweapon)
	p.Gadget.Duration = -1
	c.Core.Combat.AddGadget(p)
	p.Reactable.Init(p, c.Core)
	return p
}

func (p *ReactableWeapon) Tick() {
	p.Reactable.Tick()
	p.Gadget.Tick()
}

func (p *ReactableWeapon) Type() targets.TargettableType { return targets.TargettableGadget }

func (p *ReactableWeapon) HandleAttack(atk *combat.AttackEvent) float64 {
	return 0
}

func (p *ReactableWeapon) weaponreact(atk *combat.AttackEvent) (float64, bool) {
	if atk.Info.Abil != "Weapon Infusion" {
		return 0, false
	}
	if atk.Info.Durability < 0.00000000001 { //ZeroDur
		return 0, false
	}
	override := false
	if p.AuraCount() != 0 {
		switch atk.Info.Element {
		case attributes.Electro:
			if p.AuraContains(attributes.Hydro) {
				break
			} else {
				p.React(atk)
			}
		case attributes.Hydro:
			if p.AuraContains(attributes.Pyro) {
				override = true
			}
			if p.AuraContains(attributes.Electro) {
				break
			} else {
				p.React(atk)
			}
		case attributes.Pyro:
			if p.AuraContains(attributes.Cryo) {
				override = true
			}
			p.React(atk)
		default:
			if p.AuraContains(attributes.Geo) {
				break
			}
			p.React(atk)
		}
	} else {
		override = true
	}

	if override {
		p.resetgauge()
		switch atk.Info.Element {
		case attributes.Electro:
			p.Durability[reactable.Electro] = atk.Info.Durability
			p.DecayRate[reactable.Electro] = 0
		case attributes.Pyro:
			p.Durability[reactable.Pyro] = atk.Info.Durability
			p.DecayRate[reactable.Electro] = 0
		case attributes.Cryo:
			p.Durability[reactable.Cryo] = atk.Info.Durability
			p.DecayRate[reactable.Cryo] = 0
		case attributes.Hydro:
			p.Durability[reactable.Hydro] = atk.Info.Durability
			p.DecayRate[reactable.Hydro] = 0
		case attributes.Anemo:
			p.Durability[reactable.Anemo] = atk.Info.Durability
			p.DecayRate[reactable.Anemo] = 0
		case attributes.Geo:
			p.Durability[reactable.Geo] = 0 // Geo infusion cannot be used for reaction
			p.DecayRate[reactable.Geo] = 0
		case attributes.Dendro:
			p.Durability[reactable.Dendro] = atk.Info.Durability
			p.DecayRate[reactable.Dendro] = 0
		}
	}
	return 0, false
}

func (p *ReactableWeapon) resetgauge() {
	p.Durability[reactable.Electro] = 0
	p.Durability[reactable.Pyro] = 0
	p.Durability[reactable.Cryo] = 0
	p.Durability[reactable.Hydro] = 0
	p.Durability[reactable.Anemo] = 0
	p.Durability[reactable.Geo] = 0
	p.Durability[reactable.Dendro] = 0
}

