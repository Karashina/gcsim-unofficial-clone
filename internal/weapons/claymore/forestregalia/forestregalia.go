package forestregalia

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.ForestRegalia, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey  = "forestregalia-icd"
	buffKey = "forest-sanctuary"
)

var procEvents = []event.Event{
	event.OnBurning,
	event.OnQuicken,
	event.OnAggravate,
	event.OnSpread,
	event.OnBloom,
	event.OnHyperbloom,
	event.OnBurgeon,
}

// 燃焼、激化、超激化、草激化、開花、超開花、
// 烈開花を発動した後、キャラクターの周囲に意識の葉が最大10秒間生成される。
// 拾うとキャラクターの元素熟知が12秒間60/75/90/105/120増加する。
// この方法で生成できる葉は20秒に1枚のみ。この効果はキャラクターが
// フィールドにいなくても発動可能。意識の葉の効果は重ね掛け不可。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	pickupDelay := p.Params["pickup_delay"]

	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = float64(45 + r*15)

	//nolint:unparam // why events have a return value...
	handleProc := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 1200, true)
		c.Log.NewEvent("forestregalia proc'd", glog.LogWeaponEvent, char.Index)
		if pickupDelay <= 0 {
			c.Log.NewEvent("forestregalia leaf ignored", glog.LogWeaponEvent, char.Index)
			return false
		}
		c.Tasks.Add(func() {
			active := c.Player.ActiveChar()
			active.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(buffKey, 720),
				AffectedStat: attributes.NoStat,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
			c.Log.NewEvent(
				fmt.Sprintf("forestregalia leaf picked up by %v", active.Base.Key.String()),
				glog.LogWeaponEvent,
				char.Index,
			)
		}, pickupDelay)
		return false
	}

	key := fmt.Sprintf("forestregalia-%v", char.Base.Key.String())
	for _, e := range procEvents {
		switch e {
		case event.OnHyperbloom, event.OnBurgeon:
			c.Events.Subscribe(e, handleProc, key)
		default:
			c.Events.Subscribe(e, func(args ...interface{}) bool {
				if _, ok := args[0].(*enemy.Enemy); !ok {
					return false
				}
				return handleProc(args...)
			}, key)
		}
	}

	return w, nil
}
