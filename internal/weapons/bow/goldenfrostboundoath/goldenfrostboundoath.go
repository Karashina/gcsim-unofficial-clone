package goldenfrostboundoath

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
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
	core.RegisterWeaponFunc(keys.GoldenFrostboundOath, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	favorKey      = "frost-faes-favor"
	mischiefKey   = "frost-faes-mischief"
	favorDuration = 6 * 60 // 6秒
)

// 防御力増加: 16%/20%/24%/28%/32%
var defBonus = []float64{0.16, 0.20, 0.24, 0.28, 0.32}

// Frost Fae's Favor: 岩ダメージ増加 40%/50%/60%/70%/80%
var geoBonus = []float64{0.40, 0.50, 0.60, 0.70, 0.80}

// Frost Fae's Favor: LCrs反応ダメージ増加 40%/50%/60%/70%/80%
var lcrsBonus = []float64{0.40, 0.50, 0.60, 0.70, 0.80}

// Frost Fae's Mischief: 岩ダメージ増加 20%/25%/30%/35%/40%
var partyGeoBonus = []float64{0.20, 0.25, 0.30, 0.35, 0.40}

// Frost Fae's Mischief: LCrs反応ダメージ増加 20%/25%/30%/35%/40%
var partyLcrsBonus = []float64{0.20, 0.25, 0.30, 0.35, 0.40}

// 霜契の金枝 - Golden Frostbound Oath
// 防御力が16%/20%/24%/28%/32%増加する。
// 装備キャラの元素スキルまたはLunar-Crystallize攻撃が敵に命中すると、
// 「Frost Fae's Favor」効果を6秒間獲得:
// 装備キャラの岩ダメージが40%/50%/60%/70%/80%増加、
// Lunar-Crystallize反応ダメージが40%/50%/60%/70%/80%増加。
// この効果中にMoondriftsが装備キャラの近くにいる場合、
// 他のパーティメンバーに「Frost Fae's Mischief」効果を付与:
// 岩ダメージ+20%/25%/30%/35%/40%、
// Lunar-Crystallize反応ダメージ+20%/25%/30%/35%/40%。
// フィールドにいなくても発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	if r < 1 || r > 5 {
		return nil, fmt.Errorf("goldenfrostboundoath: invalid refine %d", r)
	}

	// 防御力バフ（常時）
	defMod := make([]float64, attributes.EndStatType)
	defMod[attributes.DEFP] = defBonus[r-1]
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("goldenfrostboundoath-def", -1),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return defMod, true
		},
	})

	// 装備キャラの岩ダメージバフ（Frost Fae's Favor）
	geoMod := make([]float64, attributes.EndStatType)
	geoMod[attributes.GeoP] = geoBonus[r-1]
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("goldenfrostboundoath-geo", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if !char.StatusIsActive(favorKey) {
				return nil, false
			}
			if atk.Info.Element != attributes.Geo {
				return nil, false
			}
			return geoMod, true
		},
	})

	// 装備キャラのLCrs反応ダメージバフ（Frost Fae's Favor）
	char.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
		Base: modifier.NewBase("goldenfrostboundoath-lcrs", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if !char.StatusIsActive(favorKey) {
				return 0, false
			}
			return lcrsBonus[r-1], false
		},
	})

	// パーティメンバーへのMischiefバフ
	for _, party := range c.Player.Chars() {
		if party.Index == char.Index {
			continue // 装備キャラ自身は対象外
		}

		idx := party.Index
		partyGeoMod := make([]float64, attributes.EndStatType)
		partyGeoMod[attributes.GeoP] = partyGeoBonus[r-1]

		party.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("goldenfrostboundoath-mischief-geo", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				_ = idx
				if !char.StatusIsActive(mischiefKey) {
					return nil, false
				}
				if atk.Info.Element != attributes.Geo {
					return nil, false
				}
				return partyGeoMod, true
			},
		})

		party.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
			Base: modifier.NewBase("goldenfrostboundoath-mischief-lcrs", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if !char.StatusIsActive(mischiefKey) {
					return 0, false
				}
				return partyLcrsBonus[r-1], false
			},
		})
	}

	// 元素スキルまたはLCrs攻撃が敵に命中した際にFrost Fae's Favorを発動
	key := fmt.Sprintf("goldenfrostboundoath-%v", char.Base.Key.String())
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if ae.Info.ActorIndex != char.Index {
			return false
		}

		// 元素スキルまたはLCrs攻撃のみ
		isSkill := ae.Info.AttackTag == attacks.AttackTagElementalArt ||
			ae.Info.AttackTag == attacks.AttackTagElementalArtHold
		isLCrs := ae.Info.AttackTag == attacks.AttackTagLCrsDamage
		if !isSkill && !isLCrs {
			return false
		}

		char.AddStatus(favorKey, favorDuration, true)
		c.Log.NewEvent("Frost Fae's Favor activated", glog.LogWeaponEvent, char.Index).
			Write("duration", favorDuration)

		// Moondriftsがある場合、パーティにMischiefを付与
		if char.MoonsignAscendant {
			char.AddStatus(mischiefKey, favorDuration, true)
			c.Log.NewEvent("Frost Fae's Mischief activated (Moondrifts present)", glog.LogWeaponEvent, char.Index)
		}

		return false
	}, key)

	return w, nil
}
