package common

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

type Blackcliff struct {
	Index int
	data  *model.WeaponData
}

func (b *Blackcliff) SetIndex(idx int)        { b.Index = idx }
func (b *Blackcliff) Init() error             { return nil }
func (b *Blackcliff) Data() *model.WeaponData { return b.data }

func NewBlackcliff(data *model.WeaponData) *Blackcliff {
	return &Blackcliff{
		data: data,
	}
}

func (b *Blackcliff) NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	atk := 0.09 + float64(p.Refine)*0.03
	index := 0
	stackKey := []string{
		"blackcliff-stack-1",
		"blackcliff-stack-2",
		"blackcliff-stack-3",
	}
	m := make([]float64, attributes.EndStatType)

	amtfn := func() ([]float64, bool) {
		count := 0
		for _, v := range stackKey {
			if char.StatusIsActive(v) {
				count++
			}
		}
		m[attributes.ATKP] = atk * float64(count)
		return m, true
	}

	c.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		_, ok := args[0].(*enemy.Enemy)
		// 敵でなければ無視
		if !ok {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		// 別のキャラクターが敵を倒した場合は発動しない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// フィールド外では発動しない
		if c.Player.Active() != char.Index {
			return false
		}
		// 指定インデックスのキャラにステータスを追加
		char.AddStatus(stackKey[index], 1800, true)
		// バフを更新
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("blackcliff", 1800),
			AffectedStat: attributes.ATKP,
			Amount:       amtfn,
		})
		index++
		if index == 3 {
			index = 0
		}
		return false
	}, fmt.Sprintf("blackcliff-%v", char.Base.Key.String()))

	return b, nil
}
