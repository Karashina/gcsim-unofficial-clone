package hunterspath

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
	buffKey = "hunterspath-tireless-hunt"
	icdKey  = "hunterspath-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.HuntersPath, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 全元素ダメージボーナス12%増加。重撃が敵に命中した後、
	// Tireless Hunt効果を獲得し、重撃ダメージが元素熟知の160%分増加する。
	// この効果は重撃を 12回当てた後または10秒後に解除される。
	// Tireless Huntは12秒毎に1回のみ獲得可能。
	w := &Weapon{}
	r := p.Refine

	dmgBonus := 0.09 + 0.03*float64(r)
	val := make([]float64, attributes.EndStatType)
	for i := attributes.PyroP; i <= attributes.DendroP; i++ {
		val[i] = dmgBonus
	}
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("hunterspath-dmg-bonus", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	caBoost := 1.2 + 0.4*float64(r)
	procCount := 0
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// バフはping依存のアクションであり、最初の命中には
		// 追加ダメージあり。
		if !char.StatusIsActive(icdKey) {
			char.AddStatus(buffKey, 600, true)
			char.AddStatus(icdKey, 720, true)
			procCount = 12
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
		c.Log.NewEvent("hunterspath proc dmg add", glog.LogPreDamageMod, char.Index).
			Write("base_added_dmg", baseDmgAdd).
			Write("remaining_stacks", procCount)
		return false
	}, fmt.Sprintf("hunterspath-%v", char.Base.Key.String()))

	return w, nil
}
