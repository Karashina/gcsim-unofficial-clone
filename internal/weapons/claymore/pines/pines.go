package pines

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/weapons/common"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
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
	core.RegisterWeaponFunc(keys.SongOfBrokenPines, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 風の中を彷徨う「千年の大楽章」の一部。
// 攻撃力が16%増加し、通常攻撃または重撃が敵に命中すると
// 囁きの印を獲得する。この効果は0.3秒に1回発動可能。
// 囁きの印を4つ所持すると、全て消費され近くの全パーティメンバーが
// 「千年の大楽章・旗振りの歌」効果を12秒間獲得する。
// 「千年の大楽章・旗振りの歌」は通常攻撃速度を12%、攻撃力を20%増加させる。
// この効果が発動すると、20秒間囁きの印を獲得できない。
// 「千年の大楽章」の多くの効果のうち、
// 同じタイプのバフは重ね掛け不可。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.12 + float64(r)*0.04
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("pines-atk", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	uniqueVal := make([]float64, attributes.EndStatType)
	uniqueVal[attributes.AtkSpd] = 0.09 + 0.03*float64(r)

	sharedVal := make([]float64, attributes.EndStatType)
	sharedVal[attributes.ATKP] = 0.15 + 0.05*float64(r)

	stacks := 0
	buffDuration := 12 * 60
	const icdKey = "songofbrokenpines-icd"
	icd := int(0.3 * 60)
	const cdKey = "songofbrokenpines-cooldown"
	cd := 20 * 60

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		if char.StatusIsActive(cdKey) {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}

		char.AddStatus(icdKey, icd, true)
		stacks++
		if stacks == 4 {
			stacks = 0
			char.AddStatus(cdKey, cd, true)
			for _, char := range c.Player.Chars() {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("pines-proc", buffDuration),
					AffectedStat: attributes.AtkSpd,
					Amount: func() ([]float64, bool) {
						if c.Player.CurrentState() != action.NormalAttackState {
							return nil, false
						}
						return uniqueVal, true
					},
				})
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(common.MillennialKey, buffDuration),
					AffectedStat: attributes.ATKP,
					Amount: func() ([]float64, bool) {
						return sharedVal, true
					},
				})
			}
		}
		return false
	}, fmt.Sprintf("pines-%v", char.Base.Key.String()))

	return w, nil
}
