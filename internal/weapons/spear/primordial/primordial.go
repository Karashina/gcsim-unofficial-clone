package primordial

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.PrimordialJadeWingedSpear, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	buff   []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 命中時、攻撃力が6秒間3.2%増加。最大7スタック。
	// この効果は0.3秒毎に1回発動可能。
	// 最大スタック所持時、与えるダメージが12%増加する。
	w := &Weapon{}
	r := p.Refine
	const icdKey = "primordial-jade-spear-icd"
	const buffKey = "primordial"
	w.buff = make([]float64, attributes.EndStatType)
	perStackBuff := float64(r)*0.007 + 0.025
	dmgBuffAtMax := 0.09 + float64(r)*0.03

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		// キャラクターが正しいかチェック
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		// 槍のICD中かチェック
		if char.StatusIsActive(icdKey) {
			return false
		}
		// バフが期限切れならスタックをリセット
		if !char.StatModIsActive(buffKey) {
			w.stacks = 0
			w.buff[attributes.DmgP] = 0
		}
		// 0.3秒ごと
		char.AddStatus(icdKey, 18, true)

		if w.stacks < 7 {
			w.stacks++
			// 最大スタックか増加量をチェック
			if w.stacks == 7 {
				w.buff[attributes.DmgP] = dmgBuffAtMax
			}
			w.buff[attributes.ATKP] = float64(w.stacks) * perStackBuff
		}

		// 修飾子を更新
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, 6*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return w.buff, true
			},
		})

		return false
	}, fmt.Sprintf("primordial-%v", char.Base.Key.String()))
	return w, nil
}
