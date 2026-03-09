package freedom

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/weapons/common"
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
	core.RegisterWeaponFunc(keys.FreedomSworn, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 「千年の大楽章」シリーズの一部。風の中をさまよう。
// ダメージが10%増加する。この武器を装備したキャラクターが
// 元素反応を起こした時、「叛逆の紋章」を獲得する。この効果は
// 0.5秒毎に1回発動可能で、キャラがフィールドにいなくても発動する。
// 紋章を2つ所持すると、全て消費され、近くのパーティメンバー全員が
// 12秒間「千年の大楽章：抵抗の歌」を得る。通常攻撃、重撃、
// 落下攻撃のダメージが16%増加し、攻撃力が20%増加する。
// ただし、この効果発動後20秒間は紋章を獲得できない。
// 「千年の大楽章」の同種のバフは重複しない。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.075 + float64(r)*0.025
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("freedom-dmg", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	uniqueVal := make([]float64, attributes.EndStatType)
	uniqueVal[attributes.DmgP] = .12 + 0.04*float64(r)

	sharedVal := make([]float64, attributes.EndStatType)
	sharedVal[attributes.ATKP] = .15 + float64(r)*0.05

	stacks := 0
	buffDuration := 12 * 60
	const icdKey = "freedom-sworn-sigil-icd"
	icd := int(0.5 * 60)
	const cdKey = "freedom-sworn-cooldown"
	cd := 20 * 60

	stackFunc := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
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
		c.Log.NewEvent("freedomsworn gained sigil", glog.LogWeaponEvent, char.Index).
			Write("sigil", stacks)
		if stacks == 2 {
			stacks = 0
			char.AddStatus(cdKey, cd, true)
			for _, char := range c.Player.Chars() {
				// 攻撃バフはスナップショットされるため、別のModで設定する必要がある
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(common.MillennialKey, buffDuration),
					AffectedStat: attributes.ATKP,
					Amount: func() ([]float64, bool) {
						return sharedVal, true
					},
				})
				char.AddAttackMod(character.AttackMod{
					Base: modifier.NewBaseWithHitlag("freedomsworn", buffDuration),
					Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
						switch atk.Info.AttackTag {
						case attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge:
							return uniqueVal, true
						}
						return nil, false
					},
				})
			}
		}
		return false
	}

	for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
		c.Events.Subscribe(i, stackFunc, "freedom-"+char.Base.Key.String())
	}

	return w, nil
}
