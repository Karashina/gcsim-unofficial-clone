package flute

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
)

func init() {
	core.RegisterWeaponFunc(keys.TheFlute, NewWeapon)
}

// 通常攻撃または重撃が命中時、「調和」を獲得する。5つの調和を獲得すると
// 音楽の力が発動し、周囲の敵に攻撃力の100%のダメージを与える。調和は最夰30秒間持続し、
// 0.5秒毎に最大1個獲得可能。
type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey      = "flute-icd"
	durationKey = "flute-stack-duration"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	stacks := 0

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 30, true) // 0.5秒毎
		if !char.StatusIsActive(durationKey) {
			stacks = 0
		}
		stacks++
		// スタックは30秒持続
		char.AddStatus(durationKey, 1800, true)

		if stacks == 5 {
			// 5スタックでダメージ発動
			stacks = 0
			char.DeleteStatus(durationKey)

			ai := combat.AttackInfo{
				ActorIndex: char.Index,
				Abil:       "Flute Proc",
				AttackTag:  attacks.AttackTagWeaponSkill,
				ICDTag:     attacks.ICDTagNone,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeDefault,
				Element:    attributes.Physical,
				Durability: 100,
				Mult:       0.75 + 0.25*float64(r),
			}
			trg := args[0].(combat.Target)
			c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 4), 0, 1)
		}
		return false
	}, fmt.Sprintf("flute-%v", char.Base.Key.String()))
	return w, nil
}
