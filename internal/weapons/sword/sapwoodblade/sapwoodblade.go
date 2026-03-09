package sapwoodblade

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
	core.RegisterWeaponFunc(keys.SapwoodBlade, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey  = "sapwoodblade-icd"
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

// 燃焼、激化、激化、草激化、開花、超開花、烈開花を起こした後、
// キャラクターの周囲に最奇10秒間「意識の葉」が生成される。
// 拾うとキャラに12秒間60/75/90/105/120の元素熟知が付与される。
// 20秒毎に1枚のみ生成可能。キャラがフィールドにいなくても発動する。
// 「意識の葉」の効果は重複しない。
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
		c.Log.NewEvent("sapwood proc'd", glog.LogWeaponEvent, char.Index)
		if pickupDelay <= 0 {
			c.Log.NewEvent("sapwood leaf ignored", glog.LogWeaponEvent, char.Index)
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
				fmt.Sprintf("sapwood leaf picked up by %v", active.Base.Key.String()),
				glog.LogWeaponEvent,
				char.Index,
			)
		}, pickupDelay)
		return false
	}

	key := fmt.Sprintf("sapwoodblade-%v", char.Base.Key.String())
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
