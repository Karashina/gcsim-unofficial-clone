package athousandblazingsuns

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/template/nightsoul"
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
	SkillIcdKey  = "atbs-skill-icd"
	AttackIcdKey = "atbs-attack-icd"
	DurKey       = "atbs-skill-duration"
)

func init() {
	core.RegisterWeaponFunc(keys.AThousandBlazingSuns, NewWeapon)
}

type Weapon struct {
	Index int
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		char: char,
	}
	r := p.Refine

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		chr := args[0].(int)
		act := args[1].(action.Action)

		if chr != w.char.Index {
			return false
		}

		if act != action.ActionSkill && act != action.ActionBurst {
			return false
		}

		if char.StatusIsActive(SkillIcdKey) {
			return false
		}

		char.AddStatus(SkillIcdKey, 10*60, false)
		char.AddStatus(DurKey, 6*60, false)

		// buffs
		nonightsoul := make([]float64, attributes.EndStatType)
		nonightsoul[attributes.CD] = 0.20 + float64(r)*0
		nonightsoul[attributes.ATKP] = 0.28 + float64(r)*0

		hasnightsoul := make([]float64, attributes.EndStatType)
		hasnightsoul[attributes.CD] = nonightsoul[attributes.CD] * 1.75
		hasnightsoul[attributes.ATKP] = nonightsoul[attributes.ATKP] * 1.75

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("athousandblazingsuns", -1),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				if !char.StatusIsActive(DurKey) {
					return nil, false
				}
				if char.StatusIsActive(nightsoul.NightsoulBlessingStatus) {
					return nonightsoul, true
				} else {
					return hasnightsoul, true
				}
			},
		})

		return false
	}, fmt.Sprintf("athousandblazingsuns-onskillburstuse-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if c.Player.Active() != char.Index {
			return false
		}
		if ae.Info.ActorIndex != char.Index {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		if char.StatusIsActive(AttackIcdKey) {
			return false
		}

		char.AddStatus(AttackIcdKey, 1*60, false)
		char.ExtendStatus(DurKey, 2*60)

		// buffs
		nonightsoul := make([]float64, attributes.EndStatType)
		nonightsoul[attributes.CD] = 0.20 + float64(r)*0
		nonightsoul[attributes.ATKP] = 0.28 + float64(r)*0

		hasnightsoul := make([]float64, attributes.EndStatType)
		hasnightsoul[attributes.CD] = nonightsoul[attributes.CD] * 1.75
		hasnightsoul[attributes.ATKP] = nonightsoul[attributes.ATKP] * 1.75

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("athousandblazingsuns", -1),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				if !char.StatusIsActive(DurKey) {
					return nil, false
				}
				if char.StatusIsActive(nightsoul.NightsoulBlessingStatus) {
					return nonightsoul, true
				} else {
					return hasnightsoul, true
				}
			},
		})

		return false
	}, fmt.Sprintf("athousandblazingsuns-onattack-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {

		if !char.StatusIsActive(DurKey) {
			return false
		}
		if c.Player.Active() == char.Index {
			return false
		}
		if !char.StatusIsActive(nightsoul.NightsoulBlessingStatus) {
			return false
		}

		char.ExtendStatus(DurKey, 1)

		return false
	}, fmt.Sprintf("athousandblazingsuns-ontick-%v", char.Base.Key.String()))

	return w, nil
}
