package foliar

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

const (
	buffKey = "foliar-whitemoon-bristle"
	icdKey  = "foliar-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.LightOfFoliarIncision, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 会心率が4%増加する。
	// 通常攻撃が元素ダメージを与えた後、「葉の切り口」効果を獲得し、
	// 通常攻撃と元素スキルのダメージが元素熟知の120%分増加する。
	// この効果は28回のダメージまたは12秒後に消滅する。12秒毎に1回獲得可能。
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.03 + float64(r)*0.01
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("foliar-crit-rate", -1),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	caBoost := 0.9 + 0.3*float64(r)
	procCount := 0
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if c.Player.Active() != char.Index {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if !(atk.Info.AttackTag == attacks.AttackTagNormal || atk.Info.AttackTag == attacks.AttackTagElementalArt || atk.Info.AttackTag == attacks.AttackTagElementalArtHold) {
			return false
		}
		// バフはping依存のアクションで、最初のヒットには
		// 追加ダメージあり。
		if !char.StatusIsActive(icdKey) && atk.Info.AttackTag == attacks.AttackTagNormal && atk.Info.Element != attributes.Physical {
			char.AddStatus(buffKey, 720, true)
			char.AddStatus(icdKey, 720, true)
			procCount = 28
			return false
		}
		if !char.StatusIsActive(buffKey) {
			return false
		}
		baseDmgAdd := char.Stat(attributes.EM) * caBoost
		atk.Info.FlatDmg += baseDmgAdd
		procCount -= 1
		if procCount <= 0 {
			char.DeleteStatus(buffKey)
		}
		c.Log.NewEvent("foliarincision proc dmg add", glog.LogPreDamageMod, char.Index).
			Write("base_added_dmg", baseDmgAdd).
			Write("remaining_stacks", procCount)
		return false
	}, fmt.Sprintf("foliarincision-%v", char.Base.Key.String()))

	return w, nil
}
