package cranesechoingcall

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
	core.RegisterWeaponFunc(keys.CranesEchoingCall, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	buffKey      = "crane-dmg%"
	buffDuration = 20 * 60
	energySrc    = "crane"
	energyIcdKey = "crane-energy-icd"
	energyIcd    = int(0.7 * 60)
)

// 装備キャラクターが落下攻撃で敵に命中した後、
// 近くのパーティメンバー全員の落下攻撃ダメージが20秒間28/41/54/67/80%増加する。
// 近くのパーティメンバーが落下攻撃で敵に命中した時、
// 装備キャラクターに2.5/2.75/3/3.25/3.5エネルギーを回復する。
// エネルギーは0.7秒に1回回復可能。
// 装備キャラクターがフィールドにいなくてもエネルギー回復効果は発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	mDmg := make([]float64, attributes.EndStatType)
	mDmg[attributes.DmgP] = 0.15 + float64(r)*0.13

	energyRestore := 2.25 + float64(r)*0.25

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		// 落下ダメージでのみ発動可能
		if atk.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}

		// 装備キャラからのダメージの場合、チームの落下ダメージをバフ
		if atk.Info.ActorIndex == char.Index {
			for _, char := range c.Player.Chars() {
				char.AddAttackMod(character.AttackMod{
					Base: modifier.NewBaseWithHitlag(buffKey, buffDuration),
					Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
						if atk.Info.AttackTag != attacks.AttackTagPlunge {
							return nil, false
						}
						return mDmg, true
					},
				})
			}
		}

		// 落下ダメージを行ったキャラに関係なくエネルギーを回復
		if char.StatusIsActive(energyIcdKey) {
			return false
		}
		char.AddStatus(energyIcdKey, energyIcd, true)
		char.AddEnergy(energySrc, energyRestore)

		return false
	}, fmt.Sprintf("crane-onhit-%v", char.Base.Key.String()))

	return w, nil
}
