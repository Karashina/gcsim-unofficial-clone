package viridescent

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
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.ViridescentVenerer, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.AnemoP] = 0.15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("vv-2pc", -1),
			AffectedStat: attributes.AnemoP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	if count < 4 {
		return &s, nil
	}

	// 反応ダメージ+0.6を追加
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("vv-4pc", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			// 増幅拡散でないことを確認
			if ai.Amped || ai.Catalyzed {
				return 0, false
			}
			switch ai.AttackTag {
			case attacks.AttackTagSwirlCryo:
			case attacks.AttackTagSwirlElectro:
			case attacks.AttackTagSwirlHydro:
			case attacks.AttackTagSwirlPyro:
			default:
				return 0, false
			}
			return 0.6, false
		},
	})

	vvfunc := func(ele attributes.Element, key string) func(args ...interface{}) bool {
		return func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			t, ok := args[0].(*enemy.Enemy)
			if !ok {
				return false
			}
			if atk.Info.ActorIndex != char.Index {
				return false
			}

			// キャラクターがフィールドにいなければ無視
			if c.Player.Active() != char.Index {
				return false
			}

			t.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag(key, 10*60),
				Ele:   ele,
				Value: -0.4,
			})
			c.Log.NewEventBuildMsg(glog.LogArtifactEvent, char.Index, "vv 4pc proc: ", key).Write("reaction", key).Write("char", char.Index).Write("target", t.Key())

			return false
		}
	}
	c.Events.Subscribe(event.OnSwirlCryo, vvfunc(attributes.Cryo, "vvcryo"), fmt.Sprintf("vv-4pc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnSwirlElectro, vvfunc(attributes.Electro, "vvelectro"), fmt.Sprintf("vv-4pc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnSwirlHydro, vvfunc(attributes.Hydro, "vvhydro"), fmt.Sprintf("vv-4pc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnSwirlPyro, vvfunc(attributes.Pyro, "vvpyro"), fmt.Sprintf("vv-4pc-%v", char.Base.Key.String()))

	// 二次ターゲットへのダメージ発動用の追加イベント
	// 上記のvvfuncを修正して対応しようとしたが予期せぬ結果が出たため、別途コピーしている
	// クロージャ関連の問題かもしれない
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}

		// キャラクターがフィールドにいなければ無視
		if c.Player.Active() != char.Index {
			return false
		}

		ele := atk.Info.Element
		key := "vv" + ele.String()
		switch atk.Info.AttackTag {
		case attacks.AttackTagSwirlCryo:
		case attacks.AttackTagSwirlElectro:
		case attacks.AttackTagSwirlHydro:
		case attacks.AttackTagSwirlPyro:
		default:
			return false
		}

		t.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(key, 10*60),
			Ele:   ele,
			Value: -0.4,
		})
		c.Log.NewEventBuildMsg(glog.LogArtifactEvent, char.Index, "vv 4pc proc: ", key).Write("reaction", key).Write("char", char.Index).Write("target", t.Key())

		return false
	}, fmt.Sprintf("vv-4pc-secondary-%v", char.Base.Key.String()))

	return &s, nil
}
