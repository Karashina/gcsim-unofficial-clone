package surfsup

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffDurKey = "surfsup-duration"
	buffKey    = "surfsup-buff"
	buffIcd    = "surfsup-icd"
	naICD      = "surfsup-na-hit-icd"
	vapeICD    = "surfsup-Vape-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.SurfsUp, NewWeapon)
}

type Weapon struct {
	stacks int
	core   *core.Core
	char   *character.CharWrapper
	refine int
	buffNA []float64
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core:   c,
		char:   char,
		refine: p.Refine,
		buffNA: make([]float64, attributes.EndStatType),
	}

	hpp := 0.20 + float64(p.Refine)*0 //todo: refines
	val := make([]float64, attributes.EndStatType)
	val[attributes.HPP] = hpp
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("surfsup-hpp", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		index := args[0].(int)
		e := args[1].(action.Action)
		if c.Player.Active() != index {
			return false
		}
		if e != action.ActionSkill {
			return false
		}

		w.onSkill()
		return false
	}, fmt.Sprintf("surfsup-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if c.Player.Active() != atk.Info.ActorIndex {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}

		w.onNAHit()
		return false
	}, fmt.Sprintf("surfsup-na-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnVaporize, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if c.Player.Active() != atk.Info.ActorIndex {
			return false
		}
		w.onVape()
		return false
	}, fmt.Sprintf("surfsup-vape-%v", char.Base.Key.String()))

	if !w.char.StatModIsActive(buffDurKey) {
		w.stacks = 0
	}

	w.char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag(buffKey, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			w.buffNA[attributes.DmgP] = 0.12 * float64(w.stacks) // todo:refine
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal:
				return w.buffNA, true
			default:
				return nil, false
			}
		},
	})

	return w, nil
}

func (w *Weapon) onSkill() {
	if w.char.StatusIsActive(buffIcd) {
		return
	}

	w.stacks = 4
	w.char.AddStatus(buffDurKey, 14*60, true)
	w.char.AddStatus(buffIcd, 15*60, true)
}

func (w *Weapon) onNAHit() {
	if w.char.StatusIsActive(naICD) {
		return
	}
	if w.stacks > 0 {
		w.stacks--
	}
	if w.stacks <= 0 {
		w.stacks = 0
	}
	if !w.char.StatModIsActive(buffDurKey) {
		w.stacks = 0
	}

	w.char.AddStatus(naICD, 1.5*60, true)
}

func (w *Weapon) onVape() {
	if w.char.StatusIsActive(vapeICD) {
		return
	}
	if w.stacks < 4 {
		w.stacks++
	}
	if w.stacks >= 4 {
		w.stacks = 4
	}
	if !w.char.StatModIsActive(buffDurKey) {
		w.stacks = 0
	}

	w.char.AddStatus(vapeICD, 1.5*60, true)
}
