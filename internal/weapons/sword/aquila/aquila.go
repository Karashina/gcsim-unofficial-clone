package aquila

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
	core.RegisterWeaponFunc(keys.AquilaFavonia, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 攻撃力が20%増加する。ダメージを受けた時に発動: 西風の鷹の魂が目覚め、
// 攻撃力の100%に相当するHPを回復し、周囲の敵に攻撃力の200%のダメージを与える。
// この効果は15秒毎に1回のみ発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 永続バフ
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = .15 + .05*float64(r)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("aquila favonia", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	dmg := 1.7 + .3*float64(r)
	heal := .85 + .15*float64(r)
	const icdKey = "aquila-icd"

	c.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if !di.External {
			return false
		}
		if di.Amount <= 0 {
			return false
		}
		if di.ActorIndex != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 900, true) // 15秒
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Aquila Favonia",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       dmg,
		}
		snap := char.Snapshot(&ai)
		c.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Combat.Player(), nil, 6), 1)

		c.Player.Heal(info.HealInfo{
			Caller:  char.Index,
			Target:  c.Player.Active(),
			Message: "Aquila Favonia",
			Src:     snap.Stats.TotalATK() * heal,
			Bonus:   char.Stat(attributes.Heal),
		})
		return false
	}, fmt.Sprintf("aquila-%v", char.Base.Key.String()))
	return w, nil
}
