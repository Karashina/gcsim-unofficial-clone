package wolf

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.WolfsGravestone, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 攻撃力が20%増加。命中時、HPが30%未満の敵への攻撃で
	// 全パーティメンバーの攻撃力が12秒間40%増加。30秒に1回のみ
	// 発動可能。
	w := &Weapon{}
	r := p.Refine

	// 固定攻撃力%増加
	val := make([]float64, attributes.EndStatType)
	val[attributes.ATKP] = 0.15 + 0.05*float64(r)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("wolf-flat", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	// HP閾値以下での攻撃力増加
	bonus := make([]float64, attributes.EndStatType)
	bonus[attributes.ATKP] = 0.3 + 0.1*float64(r)
	const icdKey = "wolf-gravestone-icd"

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if !c.Flags.DamageMode {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}

		if t.HP()/t.MaxHP() > 0.3 {
			return false
		}
		char.AddStatus(icdKey, 1800, true)

		for _, char := range c.Player.Chars() {
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("wolf-proc", 720),
				AffectedStat: attributes.NoStat,
				Amount: func() ([]float64, bool) {
					return bonus, true
				},
			})
		}
		return false
	}, fmt.Sprintf("wolf-%v", char.Base.Key.String()))
	return w, nil
}
