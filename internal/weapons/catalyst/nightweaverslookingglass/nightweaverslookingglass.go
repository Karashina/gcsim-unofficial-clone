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

		c.Log.NewEvent("[Nightweaver] potfnKey triggered", 3, char.Index).
			Write("frame", c.F).
			Write("ability", atk.Info.Abil)

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
		c.Log.NewEvent("[Nightweaver] nmvKey triggered", 3, char.Index).
			Write("frame", c.F)

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
				hasNMV := char.StatusIsActive(nmvKey)
				hasPOTFN := char.StatusIsActive(potfnKey)
				if hasNMV && hasPOTFN {
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
				hasNMV := char.StatusIsActive(nmvKey)
				hasPOTFN := char.StatusIsActive(potfnKey)
				bonus := 0.0
				if hasNMV && hasPOTFN {
					bonus = 0.3 + 0.1*float64(r)
				}
				c.Log.NewEvent("[Nightweaver] LBReactBonus check", 3, char.Index).
					Write("chr_index", chr.Index).
					Write("ability", ai.Abil).
					Write("has_nmv", hasNMV).
					Write("has_potfn", hasPOTFN).
					Write("bonus", bonus)
				return bonus, false
			},
		})
	}

	return w, nil
}
