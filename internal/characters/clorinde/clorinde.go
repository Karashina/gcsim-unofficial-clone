package clorinde

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Clorinde, NewChar)
}

type char struct {
	*tmpl.Character
	bollevel  int
	a1stacks  int
	a1max     float64
	a1buff    float64
	a4buff    []float64
	a4stacks  int
	a1mult    float64
	c4mult    float64
	c4max     float64
	c6count   int
	healtobol float64
}

const (
	particleICDKey = "ayato-particle-icd"
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.BurstCon = 5
	c.SkillCon = 3
	c.NormalHitNum = normalHitNum
	if c.Base.Cons >= 2 {
		c.a1mult = 0.3
		c.a1max = 2700
	} else {
		c.a1mult = 0.2
		c.a1max = 1800
	}
	if c.Base.Cons < 4 {
		c.c4mult = 0
		c.c4max = 0
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a1()
	c.a4()
	c.onExitField()
	c.CheckBoL()
	if c.Base.Cons >= 4 {
		c.c4()
	}
	return nil
}

func (c *char) AdvanceNormalIndex() {
	c.NormalCounter++

}

func (c *char) CheckBoL() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != c.Index {
			return false
		}

		c.bollevel = 0

		currentbol := c.CurrentHPDebt() / c.MaxHP()
		c.Core.Log.NewEvent("Current BoL Amount", glog.LogCharacterEvent, c.Index).
			Write("amt", currentbol)

		if currentbol <= 0 {
			c.bollevel = 0
		} else if currentbol < 1 {
			c.bollevel = 1
		} else {
			c.bollevel = 2
		}

		return false
	}, "clorinde-bol-check")
}

// heal override
func (c *char) Heal(hi *info.HealInfo) (float64, float64) {
	if c.StatusIsActive(skillBuffKey) {
		hp, bonus := c.CalcHealAmount(hi)

		// save previous hp related values for logging
		prevHPRatio := c.CurrentHPRatio()
		prevHP := c.CurrentHP()
		prevHPDebt := c.CurrentHPDebt()

		// calc original heal amount
		healAmt := hp * bonus

		// calc actual heal amount considering hp debt
		heal := healAmt - c.CurrentHPDebt()
		if heal < 0 {
			heal = 0
		}

		// overheal is always 0 when the healing is blocked
		overheal := 0.0

		c.ModifyHPDebtByAmount(healAmt * c.healtobol)

		// still emit event for clam, sodp, rightful reward, etc
		c.Core.Log.NewEvent(hi.Message, glog.LogHealEvent, c.Index).
			Write("previous_hp_ratio", prevHPRatio).
			Write("previous_hp", prevHP).
			Write("previous_hp_debt", prevHPDebt).
			Write("base amount", hp).
			Write("bonus", bonus).
			Write("final amount before hp debt", healAmt).
			Write("final amount after hp debt", heal).
			Write("overheal", overheal).
			Write("current_hp_ratio", c.CurrentHPRatio()).
			Write("current_hp", c.CurrentHP()).
			Write("current_hp_debt", c.CurrentHPDebt()).
			Write("max_hp", c.MaxHP())

		c.Core.Events.Emit(event.OnHeal, hi, c.Index, heal, overheal, healAmt)

		return heal, healAmt
	} else {
		hp, bonus := c.CalcHealAmount(hi)

		// save previous hp related values for logging
		prevHPRatio := c.CurrentHPRatio()
		prevHP := c.CurrentHP()
		prevHPDebt := c.CurrentHPDebt()

		// calc original heal amount
		healAmt := hp * bonus

		// calc actual heal amount considering hp debt
		heal := healAmt - c.CurrentHPDebt()
		if heal < 0 {
			heal = 0
		}

		// calc overheal
		overheal := prevHP + heal - c.MaxHP()
		if overheal < 0 {
			overheal = 0
		}

		// update hp debt based on original heal amount
		c.ModifyHPDebtByAmount(-healAmt)

		// perform heal based on actual heal amount
		c.ModifyHPByAmount(heal)

		c.Core.Log.NewEvent(hi.Message, glog.LogHealEvent, c.Index).
			Write("previous_hp_ratio", prevHPRatio).
			Write("previous_hp", prevHP).
			Write("previous_hp_debt", prevHPDebt).
			Write("base amount", hp).
			Write("bonus", bonus).
			Write("final amount before hp debt", healAmt).
			Write("final amount after hp debt", heal).
			Write("overheal", overheal).
			Write("current_hp_ratio", c.CurrentHPRatio()).
			Write("current_hp", c.CurrentHP()).
			Write("current_hp_debt", c.CurrentHPDebt()).
			Write("max_hp", c.MaxHP())

		c.Core.Events.Emit(event.OnHeal, hi, c.Index, heal, overheal, healAmt)

		return heal, healAmt
	}
}

func (c *char) getTotalAtk() float64 {
	stats, _ := c.Stats()
	return c.Base.Atk*(1+stats[attributes.ATKP]) + stats[attributes.ATK]
}
