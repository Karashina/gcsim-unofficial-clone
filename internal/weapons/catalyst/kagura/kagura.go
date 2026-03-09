package kagura

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
	core.RegisterWeaponFunc(keys.KagurasVerity, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 元素スキル使用時に神楽の舞効果を獲得し、
	// 装備キャラクターの元素スキルダメージが16秒間12%増加する。
	// 最大3スタック。スタックが3の時、全元素ダメージボーナスが
	// 12%増加する。
	w := &Weapon{}
	r := p.Refine
	stacks := 0
	key := fmt.Sprintf("kaguradance-%v", char.Base.Key.String())
	dmg := 0.12 + 0.03*float64(r-1)
	val := make([]float64, attributes.EndStatType)
	const stackKey = "kaguras-verity-stacks"
	stackDuration := 960 // 16s * 60

	//TODO: 以前はpostskillで発動していた。変更で問題がないか確認が必要
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		if !char.StatusIsActive(stackKey) {
			// スタックがリセットされていればスタックを0に戻す
			stacks = 0
		}
		char.AddStatus(stackKey, stackDuration, true)
		if stacks < 3 {
			stacks++
		}
		// 3スタック時の元素ダメージボーナス
		if stacks == 3 {
			val[attributes.PyroP] = dmg
			val[attributes.HydroP] = dmg
			val[attributes.CryoP] = dmg
			val[attributes.ElectroP] = dmg
			val[attributes.AnemoP] = dmg
			val[attributes.GeoP] = dmg
			val[attributes.PhyP] = dmg
			val[attributes.DendroP] = dmg
		} else {
			// 3スタック未満の場合は元素ダメージ%をクリア
			val[attributes.PyroP] = 0
			val[attributes.HydroP] = 0
			val[attributes.CryoP] = 0
			val[attributes.ElectroP] = 0
			val[attributes.AnemoP] = 0
			val[attributes.GeoP] = 0
			val[attributes.PhyP] = 0
			val[attributes.DendroP] = 0
		}
		// 持続時間分の修飾子を追加、前回を上書き
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("kaguras-verity", stackDuration),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.ActorIndex != char.Index {
					return nil, false
				}
				if atk.Info.AttackTag == attacks.AttackTagElementalArt || atk.Info.AttackTag == attacks.AttackTagElementalArtHold {
					val[attributes.DmgP] = dmg * float64(stacks)
				} else {
					val[attributes.DmgP] = 0
				}
				return val, true
			},
		})
		return false
	}, key)

	return w, nil
}
