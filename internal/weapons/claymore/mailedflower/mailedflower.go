package mailedflower

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
	core.RegisterWeaponFunc(keys.MailedFlower, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// キャラクターの元素スキルが敵に命中するか、元素反応を起こしてから8秒間、
// 攻撃力と元素熟知がそれぞれ12%/15%/18%/21%/24%え48/60/72/84/96増加する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	atk := 0.09 + float64(r)*0.03
	em := 36 + float64(r)*12
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atk
	m[attributes.EM] = em

	f := func() {
		if c.Player.Active() != char.Index {
			return
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("mailedflower", 8*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	fDamage := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		f()
		return false
	}
	fReact := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		f()
		return false
	}

	c.Events.Subscribe(event.OnEnemyDamage, fDamage, "mailedflower-skill-"+char.Base.Key.String())
	// 砕氷を元素反応として扱う
	for i := event.ReactionEventStartDelim + 1; i < event.ReactionEventEndDelim; i++ {
		c.Events.Subscribe(i, fReact, "mailedflower-"+char.Base.Key.String())
	}

	return w, nil
}
