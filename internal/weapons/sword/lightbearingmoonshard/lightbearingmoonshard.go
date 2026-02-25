package lightbearingmoonshard

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.LightbearingMoonshard, NewWeapon)
}

// Lightbearing Moonshard
// Sword, 4-star
// Base ATK: 44
// Sub-stat: CRIT DMG 19.2% (at Lv90)
//
// Passive: Radiance of the Lunar Palace
// DEF +20/25/30/35/40% (permanent passive)
// On Elemental Skill use:
// - Lunar-Crystallize (LCrs) reaction DMG +64/80/96/112/128% for 5s

const (
	defKey       = "lightbearingmoonshard-def"
	lcrsBonusKey = "lightbearingmoonshard-lcrs"
	lcrsBonusDur = 5 * 60 // 5s
)

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}

	r := p.Refine

	// DEF% bonus by refine level (permanent passive)
	defBonus := []float64{0.20, 0.25, 0.30, 0.35, 0.40}
	// LCrs DMG bonus by refine level
	lcrsBonus := []float64{0.64, 0.80, 0.96, 1.12, 1.28}

	// W-2: DEF% is a permanent passive, not skill-triggered
	mDef := make([]float64, attributes.EndStatType)
	mDef[attributes.DEFP] = defBonus[r-1]
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(defKey, -1),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return mDef, true
		},
	})

	// W-1/W-3: On Elemental Skill use, grant LCrs DMG bonus via LCrsReactBonusMod for 5s
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		char.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lcrsBonusKey, lcrsBonusDur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return lcrsBonus[r-1], false
			},
		})

		c.Log.NewEvent("Lightbearing Moonshard LCrs bonus activated", glog.LogWeaponEvent, char.Index).
			Write("def_bonus", defBonus[r-1]).
			Write("lcrs_bonus", lcrsBonus[r-1]).
			Write("lcrs_duration", lcrsBonusDur)

		return false
	}, "lightbearingmoonshard-skill-"+char.Base.Key.String())

	return w, nil
}
