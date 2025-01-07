package citlali

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
)

func (c *char) newShield(base float64, dur int) *shd {
	n := &shd{}
	n.Tmpl = &shield.Tmpl{}
	n.Tmpl.ActorIndex = c.Index
	n.Tmpl.Target = -1
	n.Tmpl.Src = c.Core.F
	n.Tmpl.ShieldType = shield.CitlaliSkill
	n.Tmpl.Ele = attributes.Cryo
	n.Tmpl.HP = base
	n.Tmpl.Name = "Opal Shield"
	n.Tmpl.Expires = c.Core.F + dur
	n.c = c
	return n
}

type shd struct {
	*shield.Tmpl
	c *char
}
