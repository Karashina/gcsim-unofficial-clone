package lumidouce

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.LumidouceElegy, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	core   *core.Core
	char   *character.CharWrapper
	refine int
	buff   []float64
}

const (
	buffKey = "lumidouce-buff"
	buffIcd = "lumidouce-icd"
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core:   c,
		char:   char,
		refine: p.Refine,
	}

	mATK := make([]float64, attributes.EndStatType)
	mATK[attributes.ATKP] = 0.15 + float64(w.refine)*0
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("lumidouce-atk", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return mATK, true
		},
	})

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		trg, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)

		if !ok {
			return false
		}

		if atk.Info.ActorIndex != char.Index {
			return false
		}

		if !trg.IsBurning() && atk.Info.Element != attributes.Dendro {
			return false
		}

		w.StackHandle()
		return false
	}, fmt.Sprintf("lumidouce-onstack-%v", char.Base.Key.String()))
	return w, nil
}

func (w *Weapon) StackHandle() {
	if !w.char.StatModIsActive(buffKey) {
		w.stacks = 0
	}
	if w.stacks < 2 {
		w.stacks++
	}

	w.char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag(buffKey, 8*60),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			w.buff[attributes.DmgP] = (0.18 + 0*float64(w.refine)) * float64(w.stacks)
			return w.buff, true
		},
	})
	if w.stacks == 2 && !w.char.StatusIsActive(buffIcd) {
		w.char.AddEnergy("lumidouce-energy", 12*float64(w.refine))
		w.char.AddStatus(buffIcd, 12*60, true)
	}
}
