package reliquaryoftruth

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

const (
	secretKey = "reliquary-secret"
	moonKey   = "reliquary-moon"
)

func init() {
	core.RegisterWeaponFunc(keys.ReliquaryOfTruth, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

/*
r1/r2/r3/r4/r5
CRIT Rate is increased by 8%/10%/12%/14%/16%.
When the equipping character unleashes an Elemental Skill, they gain the Secret of Lies effect: Elemental Mastery is increased by 80/100/120/140/160 for 12s.
When the equipping character deals Lunar-Bloom DMG to an opponent, they gain the Moon of Truth effect: CRIT DMG is increased by 24%/30%/36%/42%/48% for 4s.
When both the Secret of Lies and Moon of Truth effects are active at the same time, the results of both effects will be multiplied by 1.5.
*/
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// base CR
	mCR := make([]float64, attributes.EndStatType)
	mCR[attributes.CR] = 0.06 + float64(r)*0.02 // r1..r5 -> 8%..16%
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("reliquaryoftruth-cr", -1),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return mCR, true
		},
	})

	// On skill: grant Secret of Lies (EM buff for 12s)
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		mEM := make([]float64, attributes.EndStatType)
		mEM[attributes.EM] = 60 + float64(r)*20 // 80/100/120/140/160

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(secretKey, 12*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				if char.StatusIsActive(moonKey) {
					m := make([]float64, attributes.EndStatType)
					m[attributes.EM] = mEM[attributes.EM] * 1.5
					return m, true
				}
				return mEM, true
			},
		})
		return false
	}, fmt.Sprintf("reliquaryoftruth-skill-%v", char.Base.Key.String()))

	// On Lunar-Bloom: grant Moon of Truth (CRIT DMG buff for 4s)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// args: target, *combat.AttackEvent
		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagLBDamage {
			return false
		}

		mCD := make([]float64, attributes.EndStatType)
		mCD[attributes.CD] = 0.18 + float64(r)*0.06 // 24%/30%/36%/42%/48%

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(moonKey, 4*60),
			AffectedStat: attributes.CD,
			Amount: func() ([]float64, bool) {
				if char.StatusIsActive(secretKey) {
					m := make([]float64, attributes.EndStatType)
					m[attributes.CD] = mCD[attributes.CD] * 1.5
					return m, true
				}
				return mCD, true
			},
		})
		return false
	}, fmt.Sprintf("reliquaryoftruth-lb-%v", char.Base.Key.String()))

	return w, nil
}

