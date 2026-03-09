package kingssquire

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.KingsSquire, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキルと元素爆発を発動すると、森の教え効果を獲得し、元素熟知が60/80/100/120/140増加する（12秒間）。
// この効果はキャラクター切り替え時に解除される。
// 森の教え効果が終了または解除された時、近くの敵 1体に攻撃力100/120/140/160/180%分のダメージを与える。
// 森の教え効果は20秒毎に1回発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	const buffKey = "kingssquire"
	const icdKey = "kingssquire-icd"

	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 40 + float64(r)*20

	triggerAttack := func() {
		if !char.StatModIsActive(buffKey) {
			return
		}
		char.DeleteStatMod(buffKey)

		// 攻撃位置を決定
		player := c.Combat.Player()
		enemy := c.Combat.ClosestEnemyWithinArea(combat.NewCircleHitOnTarget(player, nil, 15), nil)
		var pos geometry.Point
		if enemy == nil {
			pos = player.Pos()
		} else {
			pos = enemy.Pos()
		}

		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "King's Squire Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Mult:       0.8 + float64(r)*0.2,
		}
		c.QueueAttack(ai, combat.NewCircleHitOnTarget(pos, nil, 1.6), 0, 1)
	}

	f := func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 20*60, true)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, 12*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		char.QueueCharTask(triggerAttack, 12*60)
		return false
	}

	c.Events.Subscribe(event.OnSkill, f, fmt.Sprintf("kingssquire-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnBurst, f, fmt.Sprintf("kingssquire-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		if prev == char.Index {
			triggerAttack()
		}
		return false
	}, fmt.Sprintf("kingssquire-%v", char.Base.Key.String()))

	return w, nil
}
