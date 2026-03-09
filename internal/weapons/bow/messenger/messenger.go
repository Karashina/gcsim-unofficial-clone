package messenger

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
)

func init() {
	core.RegisterWeaponFunc(keys.Messenger, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 重撃が弱点に命中した時、攻撃力100/125/150/175/200%分の追加ダメージを会心ダメージとして与える。
// 10秒毎に1回のみ発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	dmg := 0.75 + float64(r)*0.25
	const icdKey = "messenger-icd"

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		trg := args[0].(combat.Target)
		// 武器装備者からのダメージでなければ発動しない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// フィールド外では発動しない
		if c.Player.Active() != char.Index {
			return false
		}
		// 弱点に命中していなければ発動しない
		if !atk.Info.HitWeakPoint {
			return false
		}
		// ICD中は発動しない
		if char.StatusIsActive(icdKey) {
			return false
		}
		// ICDをセット
		char.AddStatus(icdKey, 10*60, true) // 10s icd

		// 単体ターゲットの追加ダメージをキュー
		ai := combat.AttackInfo{
			ActorIndex:   char.Index,
			Abil:         "Messenger Proc",
			AttackTag:    attacks.AttackTagNone,
			ICDTag:       attacks.ICDTagNone,
			ICDGroup:     attacks.ICDGroupDefault,
			StrikeType:   attacks.StrikeTypePierce,
			Element:      attributes.Physical,
			Durability:   100,
			Mult:         dmg,
			HitWeakPoint: true, // ensure crit by marking it as hitting weakspot
		}
		c.QueueAttack(ai, combat.NewSingleTargetHit(trg.Key()), 0, 1)

		return false
	}, fmt.Sprintf("messenger-%v", char.Base.Key.String()))

	return w, nil
}
