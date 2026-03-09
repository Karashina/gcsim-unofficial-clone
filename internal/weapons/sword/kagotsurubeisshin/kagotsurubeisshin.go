package kagotsurubeisshin

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
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.KagotsurubeIsshin, NewWeapon)
}

// 通常攻撃、重撃、または落下攻撃が敵に命中した時、「斬風」を発生させ、
// 攻撃力の180%の範囲ダメージを与え、攻撃力が8秒間15%増加する。
// この効果は8秒毎に1回のみ発動可能。
type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const icdKey = "kagotsurube-isshin-icd"

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}

	duration := 480
	cd := 480

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra && atk.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		val := make([]float64, attributes.EndStatType)
		val[attributes.ATKP] = 0.15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("kagotsurube-isshin", duration),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return val, true
			},
		})
		// 即座に%ダメージを与える新しいアクションを追加
		// 超電導攻撃
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Kagotsurube Isshin Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       1.8,
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 1)

		// クールダウンを発動
		char.AddStatus(icdKey, cd, true)

		return false
	}, fmt.Sprintf("kagotsurube-isshin-%v", char.Base.Key.String()))
	return w, nil
}
