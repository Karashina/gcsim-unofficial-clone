package skyward

import (
	"fmt"

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
	core.RegisterWeaponFunc(keys.SkywardBlade, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	buffKey = "skyward-blade"
)

// 会心率が4%増加する。元素爆発使用時、「天空を貫く力」を得る:
// 移動速度が10%、攻撃速度が10%増加し、通常攻撃と重撃が
// 攻撃力の20%の追加ダメージを与える。「天空を貫く力」は12秒間持続する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 永続バフ
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.03 + float64(r)*0.01
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("skyward-blade-crit", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	atkspdBuff := make([]float64, attributes.EndStatType)
	atkspdBuff[attributes.AtkSpd] = 0.1
	c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, 720),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return atkspdBuff, true
			},
		})
		return false
	}, fmt.Sprintf("skyward-blade-%v", char.Base.Key.String()))

	// 通常/重撃命中時にダメージ発動。ゲーム内の説明がわかりにくい
	dmgper := .15 + .05*float64(r)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		// キャラクターが正しいかチェック
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// バフが有効かチェック
		if !char.StatModIsActive(buffKey) {
			return false
		}
		if dmg == 0 {
			return false
		}
		// 即座に%ダメージを与える新しいアクションを追加
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Skyward Blade Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       dmgper,
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewSingleTargetHit(trg.Key()), 0, 1)
		return false
	}, fmt.Sprintf("skyward-blade-%v", char.Base.Key.String()))

	return w, nil
}
