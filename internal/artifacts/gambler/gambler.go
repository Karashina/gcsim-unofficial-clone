package gambler

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.Gambler, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// 4セット効果: 敵を倒すとスキルCDをリセット
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	// 2セット: 元素スキルダメージ+20%
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.20
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("gambler-2pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagElementalArt {
					return nil, false
				}
				return m, true
			},
		})
	}

	// 4セット: 敵を倒すと100%の確率で元素スキルCDを解除。15秒に1回のみ発動可能。
	if count >= 4 {
		const icdKey = "gambler-4pc-icd"
		c.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
			// ICD中は発動しない
			if char.StatusIsActive(icdKey) {
				return false
			}
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

			// スキルクールダウンをリセット
			char.ResetActionCooldown(action.ActionSkill)
			c.Log.NewEvent("gambler-4pc proc'd", glog.LogArtifactEvent, char.Index)

			// ICDをセット
			char.AddStatus(icdKey, 900, true) // 15s

			return false
		}, fmt.Sprintf("gambler-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
