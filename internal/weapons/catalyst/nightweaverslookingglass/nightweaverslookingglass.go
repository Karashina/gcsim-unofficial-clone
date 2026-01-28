package nightweaverslookingglass

import (
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
	core.RegisterWeaponFunc(keys.NightweaversLookingGlass, NewWeapon)
}

type Weapon struct {
	Index int
	c     *core.Core
	self  *character.CharWrapper
}

const (
	potfnKey = "PrayeroftheFarNorth"
	nmvKey   = "NewMoonVerse"
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		c:    c,
		self: char,
	}
	r := p.Refine

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}
		if atk.Info.Element != attributes.Dendro && atk.Info.Element != attributes.Hydro {
			return false
		}

		mEM := make([]float64, attributes.EndStatType)
		mEM[attributes.EM] = 45 + float64(r)*15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("nightweaverslookingglass-elemdmg", 4.5*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return mEM, true
			},
		})
		char.AddStatus(potfnKey, 4.5*60, true)
		return false
	}, "nightweaverslookingglass-elemdmg")

	c.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		mEM := make([]float64, attributes.EndStatType)
		mEM[attributes.EM] = 45 + float64(r)*15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("nightweaverslookingglass-lb", 10*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return mEM, true
			},
		})
		char.AddStatus(nmvKey, 10*60, true)
		return false
	}, "nightweaverslookingglass-lb")

	for _, chr := range c.Player.Chars() {
		chr.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag("nightweaverslookingglass-bloom-buff", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if char.StatusIsActive(nmvKey) && char.StatusIsActive(potfnKey) {
					if ai.AttackTag == attacks.AttackTagBloom {
						return 0.9 + 0.3*float64(r), false
					}
					if ai.AttackTag == attacks.AttackTagHyperbloom || ai.AttackTag == attacks.AttackTagBurgeon {
						return 0.6 + 0.2*float64(r), false
					}
				}
				return 0, false
			},
		})
		chr.AddLBReactBonusMod(character.LBReactBonusMod{
			Base: modifier.NewBaseWithHitlag("nightweaverslookingglass-lb-buff", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if char.StatusIsActive(nmvKey) && char.StatusIsActive(potfnKey) {
					return 0.3 + 0.1*float64(r), false
				}
				return 0, false
			},
		})
	}

	return w, nil
}
