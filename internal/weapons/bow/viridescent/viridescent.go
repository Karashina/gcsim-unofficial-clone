package viridescent

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
	core.RegisterWeaponFunc(keys.TheViridescentHunt, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 命中時、通常攻撃と重撃は50%の確率でサイクロンを生成し、周囲の敵を引き寄せ続け、
	// 0.5秒毎に攻撃力40%分のダメージを与える（4秒間）。
	// この効果は14秒毎に1回のみ発動可能。
	w := &Weapon{}
	r := p.Refine

	const icdKey = "viridescent-hunt-icd"
	cd := 900 - r*60
	mult := 0.3 + float64(r)*0.1

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		trg := args[0].(combat.Target)

		// 通常攻撃と重撃でのみ発動
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		default:
			return false
		}

		if char.StatusIsActive(icdKey) {
			return false
		}

		if c.Rand.Float64() > 0.5 {
			return false
		}

		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Viridescent",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       mult,
		}

		for i := 0; i <= 240; i += 30 {
			c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, i+1)
		}

		char.AddStatus(icdKey, cd, true)

		return false
	}, fmt.Sprintf("viridescent-%v", char.Base.Key.String()))

	return w, nil
}
