package swordofdescension

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
	core.RegisterWeaponFunc(keys.SwordOfDescension, NewWeapon)
}

// ディセンション
// この武器の効果は以下のプラットフォームでのみ適用される:
// "PlayStation Network"
// 通常攻撃または重撃が敵に命中した時、50%の確率で攻撃力の200%のダメージを小範囲に与える。この効果は10秒毎に1回のみ発動可能。
// また、旅人が Sword of Descension を装備すると、攻撃力が66増加する。
//   - 精錬はこの武器に影響しない
type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey = "swordofdescension-icd"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)

	passive, ok := p.Params["passive"]
	if !ok {
		passive = 1
	}

	if passive != 1 {
		return w, nil
	}

	if char.Base.Key < keys.TravelerDelim {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("swordofdescension", -1),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				m[attributes.ATK] = 66
				return m, true
			},
		})
	}

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// キャラクターがフィールドにいなければ無視
		if c.Player.Active() != char.Index {
			return false
		}
		// 重撃でも通常攻撃でもない場合は無視
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// ICDがまだ有効な場合は無視
		if char.StatusIsActive(icdKey) {
			return false
		}
		// 50%の確率で無視、1:1の比率
		if c.Rand.Float64() < 0.5 {
			return false
		}
		if dmg == 0 {
			return false
		}
		char.AddStatus(icdKey, 600, true)

		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Sword of Descension Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       2.00,
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 1.5), 0, 1)

		return false
	}, fmt.Sprintf("swordofdescension-%v", char.Base.Key.String()))

	return w, nil
}
