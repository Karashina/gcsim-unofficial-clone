package sethos

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Sethos, NewChar)
}

type char struct {
	*tmpl.Character
	skillArea combat.AttackPattern
	a4stacks  int
	a4buff    float64
	c2stacks  int
	c4count   int
	c4Buff    []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.NormalCon = 3
	c.BurstCon = 5

	w.Character = &c

	c.a4stacks = 4

	return nil
}

func (c *char) Init() error {
	c.onExit()
	c.Energygen()
	if c.Base.Cons >= 1 {
		c.c1()
	}
	if c.Base.Cons >= 2 {
		c.c2()
	}
	if c.Base.Cons >= 4 {
		c.c4Buff = make([]float64, attributes.EndStatType)
		c.c4Buff[attributes.EM] = 80
	}

	return nil
}

func (c *char) Snapshot(ai *combat.AttackInfo) combat.Snapshot {
	ds := c.Character.Snapshot(ai)

	// infusion to normal attack only
	if c.StatusIsActive(burstkey) && ai.AttackTag == attacks.AttackTagExtra {
		ai.Element = attributes.Electro
		ai.FlatDmg = burst[c.TalentLvlBurst()] * c.Stat(attributes.EM)
		c.Core.Log.NewEvent("burst buff applied", glog.LogCharacterEvent, c.Index).
			Write("prev", ai.Mult).
			Write("next", burst[c.TalentLvlBurst()]*c.Stat(attributes.EM)).
			Write("char", c.Index)
	}
	return ds
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 17
	}
	return c.Character.AnimationStartDelay(k)
}
