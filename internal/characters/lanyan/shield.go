package lanyan

import (
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
)

func (c *char) newShield(base float64, dur int) *shd {

	c.shieldsrc = c.Core.F
	c.shieldexp = c.Core.F + dur
	c.shieldamt = base

	n := &shd{}
	n.Tmpl = &shield.Tmpl{}
	n.Tmpl.ActorIndex = c.Index
	n.Tmpl.Target = -1
	n.Tmpl.Src = c.shieldsrc
	n.Tmpl.ShieldType = shield.LanyanSkill
	n.Tmpl.Ele = c.shieldele
	n.Tmpl.HP = c.shieldamt
	n.Tmpl.Name = "Swallow-Wisp Shield"
	n.Tmpl.Expires = c.shieldexp
	n.c = c

	return n
}

type shd struct {
	*shield.Tmpl
	c *char
}
