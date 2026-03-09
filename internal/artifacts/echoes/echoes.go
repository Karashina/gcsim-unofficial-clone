package echoes

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
	core.RegisterSetFunc(keys.EchoesOfAnOffering, NewSet)
}

type Set struct {
	prob        float64
	icd         int
	procExpireF int
	Index       int
	Count       int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// 2セット - 攻撃力 +18%
// 4セット - 通常攻撃が敵に命中すると、36%の確率でValley Riteが発動し、通常攻撃ダメージが攻撃力の70%分増加。
//
//	この効果は通常攻撃がダメージを与えた0.05秒後に解除。
//	Valley Riteが発動しなかった場合、次回の発動確率が20%増加。
//	このトリガーは0.2秒に1回発動可能。
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	procDuration := 3 // 0.05s

	s := Set{Count: count}
	s.prob = 0.36
	s.icd = 0
	s.procExpireF = 0

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.18
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("echoes-2pc", -1),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	if count < 4 {
		return &s, nil
	}

	var dmgAdded float64

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		// if the active char is not the equipped char then ignore
		if c.Player.Active() != char.Index {
			return false
		}

		// 装備キャラクターの攻撃でなければ無視
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// 通常攻撃でなければ無視
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}

		// if buff is already active then buff attack
		snATK := char.TotalAtk()
		if c.F < s.procExpireF {
			dmgAdded = snATK * 0.7
			atk.Info.FlatDmg += dmgAdded
			c.Log.NewEvent("echoes 4pc adding dmg", glog.LogArtifactEvent, char.Index).
				Write("dmg_added", dmgAdded).
				Write("buff_expiry", s.procExpireF).
				Write("icd_up", s.icd)

			return false
		}

		// 聖遺物セット効果がまだCD中なら無視
		if c.F < s.icd {
			c.Log.NewEvent("echoes 4pc failed to proc due icd", glog.LogArtifactEvent, char.Index).
				Write("icd_up", s.icd)
			return false
		}

		if c.Rand.Float64() > s.prob {
			s.icd = c.F + 12 // 0.2s
			s.prob += 0.2
			c.Log.NewEvent("echoes 4pc failed to proc due to chance", glog.LogArtifactEvent, char.Index).
				Write("probability_now", s.prob).
				Write("icd_up", s.icd)
			return false
		}

		dmgAdded = snATK * 0.7
		atk.Info.FlatDmg += dmgAdded

		s.procExpireF = c.F + procDuration
		s.icd = c.F + 12 // 0.2s

		s.prob = 0.36

		c.Log.NewEvent("echoes 4pc adding dmg", glog.LogArtifactEvent, char.Index).
			Write("dmg_added", dmgAdded).
			Write("buff_expiry", s.procExpireF).
			Write("icd_up", s.icd)

		return false
	}, fmt.Sprintf("echoes-4pc-%v", char.Base.Key.String()))

	return &s, nil
}
