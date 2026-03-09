package filletblade

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
	core.RegisterWeaponFunc(keys.FilletBlade, NewWeapon)
}

// 命中時、50%の確率で単体の敵に攻撃力の240/280/320/360/400%のダメージを与える。
// 15/14/13/12/11秒毎に1回のみ発動可能。
type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const icdKey = "fillet-blade-icd"

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	cd := 960 - 60*r

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		if c.Rand.Float64() > 0.5 {
			return false
		}
		if dmg == 0 {
			return false
		}
		// 即座に%ダメージを与える新しいアクションを追加
		// 超電導攻撃
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Fillet Blade Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       2.0 + 0.4*float64(r),
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewSingleTargetHit(trg.Key()), 0, 1)

		// クールダウンを発動
		char.AddStatus(icdKey, cd, true)

		return false
	}, fmt.Sprintf("fillet-blade-%v", char.Base.Key.String()))
	return w, nil
}
