package ashgravendrinkinghorn

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.AshGravenDrinkingHorn, NewWeapon)
}

const icdKey = "drinkinghorn-icd"

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != char.Index {
			return false
		}

		if ae.Info.Abil != "Ash Graven Drinking Horn Proc" {
			return false
		}

		snap := char.Snapshot(&ae.Info)
		ae.Snapshot.Stats[attributes.CR] = snap.Stats[attributes.CR]
		ae.Snapshot.Stats[attributes.CD] = snap.Stats[attributes.CD]

		return false
	}, "ashgraven-crit")

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 15*60, true)
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Ash Graven Drinking Horn Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			FlatDmg:    char.MaxHP() * (0.3 + float64(r)*0.1),
		}
		trg := args[0].(combat.Target)

		c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 1)

		return false
	}, fmt.Sprintf("ashgravendrinkinghorn-%v", char.Base.Key.String()))
	return w, nil
}
