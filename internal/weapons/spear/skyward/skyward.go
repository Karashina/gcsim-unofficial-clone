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
	core.RegisterWeaponFunc(keys.SkywardSpine, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 会心率が8%増加し、通常攻撃速度が12%増加する。さらに、
// 通常攻撃と重撃が敵に命中した時、50%の確率で真空刃を発動し、
// 攻撃力40%分のダメージを小範囲AoEで与える。この効果は2秒毎に1回のみ発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 永続バフ
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.06 + float64(r)*0.02
	m[attributes.AtkSpd] = 0.12
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("skyward spine", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	const icdKey = "skyward-spine-icd"
	atk := .25 + .15*float64(r)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		// キャラクターが正しいかチェック
		if ae.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// CDがクールダウン中かチェック
		if char.StatusIsActive(icdKey) {
			return false
		}
		if c.Rand.Float64() > .5 {
			return false
		}

		// 即座に%ダメージを与える新しいアクションを追加
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Skyward Spine Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       atk,
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewBoxHitOnTarget(trg, nil, 0.1, 0.1), 0, 1)

		// クールダウンを発動
		char.AddStatus(icdKey, 120, true)
		return false
	}, fmt.Sprintf("skyward-spine-%v", char.Base.Key.String()))
	return w, nil
}
