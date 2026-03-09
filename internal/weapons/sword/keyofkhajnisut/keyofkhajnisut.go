package keyofkhajnisut

import (
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
	core.RegisterWeaponFunc(keys.KeyOfKhajNisut, NewWeapon)
}

// HPが20%増加する。元素スキルが敵に命中した時、20秒間「大聖詠」効果を得る。
// この効果は装備キャラの最大HPの0.12%分の元素熟知を増加させる。0.3秒毎に1回発動可能。
// 最大3スタック。3スタック時、または3スタック目の持続時間が更新された時、
// 近くのパーティメンバー全員の元素熟知が装備キャラの最大HPの0.2%分増20秒間増加する。
type Weapon struct {
	stacks int
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	buffKey     = "khaj-nisut-buff"
	teamBuffKey = "khaj-nisut-team-buff"
	icdKey      = "khaj-nisut-icd"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	duration := 1200 // 20*60
	cd := 18         // 0.3 * 60

	hp := 0.15 + 0.05*float64(r)
	em := 0.0009 + 0.0003*float64(r)
	emTeam := 0.0015 + 0.0005*float64(r)

	m := make([]float64, attributes.EndStatType)
	m[attributes.HPP] = hp
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("khaj-nisut", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}

		if !char.StatModIsActive(buffKey) {
			w.stacks = 0
		}
		if w.stacks < 3 {
			w.stacks++
		}

		val := make([]float64, attributes.EndStatType)
		val[attributes.EM] = char.MaxHP() * em * float64(w.stacks)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, duration),
			AffectedStat: attributes.EM,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return val, true
			},
		})

		if w.stacks == 3 {
			val := make([]float64, attributes.EndStatType)
			val[attributes.EM] = char.MaxHP() * emTeam
			for _, this := range c.Player.Chars() {
				this.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(teamBuffKey, duration),
					AffectedStat: attributes.EM,
					Extra:        true,
					Amount: func() ([]float64, bool) {
						return val, true
					},
				})
			}
		}

		char.AddStatus(icdKey, cd, true)
		return false
	}, "khaj-nisut")

	return w, nil
}
