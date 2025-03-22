package vividnotions

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

func init() {
	core.RegisterWeaponFunc(keys.VividNotions, NewWeapon)
}

const (
	buffkeyp  = "vividnotions-dawnsfirsthue"
	buffkeyeq = "vividnotions-twilightssplendor"
)

type Weapon struct {
	Index int
	core  *core.Core
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
	}
	r := p.Refine

	atkp := 0.28 + 0*float64(r) //REFINE
	matk := make([]float64, attributes.EndStatType)
	matk[attributes.ATKP] = atkp

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("vividnotions-atkp", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			return matk, true
		},
	})

	mDmg := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("vividnotions-plungebuff", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagPlunge {
				return nil, false
			}
			mDmg[attributes.DmgP] = 0
			if char.StatusIsActive(buffkeyp) {
				mDmg[attributes.DmgP] += 0.28 + float64(r)*0 //REFINE
			}
			if char.StatusIsActive(buffkeyeq) {
				mDmg[attributes.DmgP] += 0.40 + float64(r)*0 //REFINE
			}
			return mDmg, true
		},
	})

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		act := args[1].(action.Action)

		switch act {
		case action.ActionHighPlunge:
			char.AddStatus(buffkeyp, 15*60, true)
		case action.ActionLowPlunge:
			char.AddStatus(buffkeyp, 15*60, true)
		case action.ActionSkill:
			char.AddStatus(buffkeyeq, 15*60, true)
		case action.ActionBurst:
			char.AddStatus(buffkeyeq, 15*60, true)
		default:
			return false
		}

		return false
	}, fmt.Sprintf("vividnotion-onplungeexec-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != char.Index {
			return false
		}

		if ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}

		char.QueueCharTask(func() {
			char.DeleteStatus(buffkeyp)
			char.DeleteStatus(buffkeyeq)
		}, 6)

		return false
	}, fmt.Sprintf("vividnotion-deletebuff-%v", char.Base.Key.String()))

	return w, nil
}
