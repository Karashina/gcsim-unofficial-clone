package oceanhuedclam

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
	core.RegisterSetFunc(keys.OceanHuedClam, NewSet)
}

type Set struct {
	bubbleHealStacks     float64
	bubbleDurationExpiry int
	core                 *core.Core
	Index                int
	Count                int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error {
	// 現在OHC発動中のキャラクター。-1 = 非アクティブ
	s.core.Flags.Custom["OHCActiveChar"] = -1
	return nil
}

// 2セット効果: 与える治療効果 +15%
// 4セット効果: この聖遺物セットを装備したキャラがパーティメンバーを回復すると、
// 海染の泡が3秒間出現し、回復量（超過回復含む）を蓄積。
// 持続時間終了時、海染の泡が爆発し、蓄積された回復量の90%に基づくダメージを周囲の敵に与える。

// （このダメージは感電、超伝導等の反応と同様に計算されるが、
// 元素熟知、キャラクターレベル、反応ダメージボーナスの影響を受けない）。
//
//		海染の泡は3.5秒に1個のみ生成可能。各海染の泡は最大30,000HPまで蓄積可能
//	 （超過回復含む）。同時に存在できる海染の泡は1個まで。
//		この効果は装備キャラがフィールド上にいなくても発動可能。
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		core:  c,
		Count: count,
	}

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.Heal] = 0.15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("ohc-2pc", -1),
			AffectedStat: attributes.Heal,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {
		bubbleICDExpiry := 0

		// 回復イベント登録で回復量の蓄積を開始
		c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
			src := args[0].(*info.HealInfo)
			healAmt := args[4].(float64)

			if src.Caller != char.Index {
				return false
			}

			// OHCが非アクティブ、またはこの装備キャラのOHCバブルがアクティブである必要がある
			if c.Flags.Custom["OHCActiveChar"] != -1 && c.Flags.Custom["OHCActiveChar"] != float64(char.Index) {
				return false
			}

			s.bubbleHealStacks += healAmt
			if s.bubbleHealStacks >= 30000 {
				s.bubbleHealStacks = 30000
			}

			// このキャラのバブルCDが明けていればバブルを有効化し、バブル破裂タスクを追加
			if bubbleICDExpiry < c.F {
				s.bubbleDurationExpiry = c.F + 3*60
				bubbleICDExpiry = c.F + 3.5*60

				c.Flags.Custom["OHCActiveChar"] = float64(char.Index)

				// バブル破裂タスク
				c.Tasks.Add(func() {
					// 泡は物理ダメージ。反応ダメージ関数で処理されるため、物理ダメージボーナスや敵の防御の影響を受けない
					// d := c.Snapshot(
					// 	"OHC Damage",
					// 	core.AttackTagNone,
					// 	core.ICDTagNone,
					// 	core.ICDGroupDefault,
					// 	core.StrikeTypeDefault,
					// 	core.Physical,
					// 	0,
					// 	0,
					// )
					// d.Targets = core.TargetAll
					// d.IsOHCDamage = true
					// d.FlatDmg = bubbleHealStacks * .9
					// c.QueueDmg(&d, 0)

					atk := combat.AttackInfo{
						ActorIndex:       char.Index,
						Abil:             "Sea-Dyed Foam",
						AttackTag:        attacks.AttackTagNoneStat,
						ICDTag:           attacks.ICDTagNone,
						ICDGroup:         attacks.ICDGroupDefault,
						StrikeType:       attacks.StrikeTypeDefault,
						Element:          attributes.Physical,
						IgnoreDefPercent: 1,
						FlatDmg:          s.bubbleHealStacks * .9,
					}
					// ステータス不要のため-1でスナップショット
					c.QueueAttack(atk, combat.NewCircleHitOnTarget(c.Combat.Player(), nil, 6), -1, 1)

					// リセット
					c.Flags.Custom["OHCActiveChar"] = -1
					s.bubbleHealStacks = 0
				}, 3*60)

				c.Log.NewEvent("ohc bubble activated", glog.LogArtifactEvent, char.Index).
					Write("bubble_pops_at", s.bubbleDurationExpiry).
					Write("ohc_icd_expiry", bubbleICDExpiry)
			}

			c.Log.NewEvent("ohc bubble accumulation", glog.LogArtifactEvent, char.Index).
				Write("bubble_pops_at", s.bubbleDurationExpiry).
				Write("bubble_total", s.bubbleHealStacks)

			return false
		}, fmt.Sprintf("ohc-4pc-heal-accumulation-%v", char.Base.Key.String()))
	}

	return &s, nil
}
