package adaycarvedfromrisingwinds

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.ADayCarvedFromRisingWinds, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// 2セット効果: 攻撃力+18%
// 4セット効果: 通常攻撃・重撃・元素スキル・元素爆発が敵に命中すると、
// 「牧歌の風の祝福」効果を6秒間獲得し、攻撃力が25%増加する。
// 装備キャラクターが「魔女の宿題」を完了している場合、「牧歌の風の祝福」は
// 「牧歌の風の決意」にアップグレードされ、装備キャラクターの会心率がさらに20%増加する。
// この効果はキャラクターが控えにいる場合でも発動可能。

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		// 2セット: 攻撃力+18%
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.18
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("adaycarved-2pc", -1),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	if count >= 4 {
		const buffKey = "adaycarved-4pc-buff"
		buffDuration := 360 // 6秒

		// 攻撃命中イベントを購読
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != char.Index {
				return false
			}

			// 通常攻撃・重撃・スキル（単押し/長押し）・爆発攻撃かチェック
			if atk.Info.AttackTag != attacks.AttackTagNormal &&
				atk.Info.AttackTag != attacks.AttackTagExtra &&
				atk.Info.AttackTag != attacks.AttackTagElementalArt &&
				atk.Info.AttackTag != attacks.AttackTagElementalArtHold &&
				atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return false
			}

			// バフを有効化
			char.AddStatus(buffKey, buffDuration, true)

			c.Log.NewEvent("a day carved from rising winds 4pc triggered", glog.LogArtifactEvent, char.Index)

			return false
		}, fmt.Sprintf("adaycarved-4pc-%v", char.Base.Key.String()))

		// バフによる攻撃力ボーナス
		atkVal := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("adaycarved-4pc-atk", -1),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				if !char.StatusIsActive(buffKey) {
					return nil, false
				}
				atkVal[attributes.ATKP] = 0.25
				return atkVal, true
			},
		})

		// ヘクセライ（魔女の宿題）属性を持つキャラクターの場合、会心率ボーナスを付与
		crVal := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("adaycarved-4pc-cr", -1),
			AffectedStat: attributes.CR,
			Amount: func() ([]float64, bool) {
				if !char.StatusIsActive(buffKey) {
					return nil, false
				}
				// キャラクターがヘクセライ属性を持つかチェック
				result, err := char.Condition([]string{"hexerei"})
				if err != nil {
					return nil, false
				}
				isHexerei, ok := result.(bool)
				if !ok || !isHexerei {
					return nil, false
				}
				crVal[attributes.CR] = 0.20
				return crVal, true
			},
		})
	}

	return &s, nil
}
