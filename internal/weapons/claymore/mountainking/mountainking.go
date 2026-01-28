package mountainking

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FangOfTheMountainKing, NewWeapon)
}

const (
	buffKey      = "mountainking-buff"
	buffIcdReact = "mountainking-icd-reaction"
	buffIcdSkill = "mountainking-icd-skill"
)

type Weapon struct {
	core   *core.Core
	char   *character.CharWrapper
	refine int
	buff   []float64
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core:   c,
		char:   char,
		refine: p.Refine,
		buff:   make([]float64, attributes.EndStatType),
	}
	stackindex := 0
	stackKey := []string{
		"mountainking-stack-1",
		"mountainking-stack-2",
		"mountainking-stack-3",
		"mountainking-stack-4",
		"mountainking-stack-5",
		"mountainking-stack-6",
	}

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}

		count := 0
		for _, v := range stackKey {
			if char.StatusIsActive(v) {
				count++
			}
		}
		w.buff[attributes.DmgP] = (0.075 + float64(w.refine)*0.025) * float64(count)
		w.applybuff()

		if w.char.StatusIsActive(buffIcdSkill) {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}

		char.AddStatus(stackKey[stackindex], 6*60, true)
		stackindex++
		if stackindex >= 6 {
			stackindex = 0
		}
		w.char.AddStatus(buffIcdSkill, 0.5*60, true)

		return false
	}, fmt.Sprintf("mountainking-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnBurning, func(args ...interface{}) bool {
		if w.char.StatusIsActive(buffIcdReact) {
			return false
		}

		for i := 0; i < 3; i++ {
			char.AddStatus(stackKey[stackindex], 6*60, true)
			stackindex++
			if stackindex >= 6 {
				stackindex = 0
			}
		}
		w.char.AddStatus(buffIcdReact, 2*60, true)

		return false
	}, fmt.Sprintf("mountainking-burning-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnBurgeon, func(args ...interface{}) bool {
		if w.char.StatusIsActive(buffIcdReact) {
			return false
		}

		for i := 0; i < 3; i++ {
			char.AddStatus(stackKey[stackindex], 6*60, true)
			stackindex++
			if stackindex >= 6 {
				stackindex = 0
			}
		}
		w.char.AddStatus(buffIcdReact, 2*60, true)

		return false
	}, fmt.Sprintf("mountainking-burgeon-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) applybuff() {
	w.char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag(buffKey, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalArt:
				return w.buff, true
			case attacks.AttackTagElementalArtHold:
				return w.buff, true
			case attacks.AttackTagElementalBurst:
				return w.buff, true
			default:
				return nil, false
			}
		},
	})
}
