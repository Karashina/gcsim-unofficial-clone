package halberd

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
	core.RegisterWeaponFunc(keys.Halberd, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 通常攻撃が追加で160/200/240/280/320%のダメージを与える。
// 10秒毎に1回のみ発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	const icdKey = "halberd-icd"
	dmg := 1.20 + float64(r)*0.40

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		trg := args[0].(combat.Target)
		// 武器装備者からのダメージでなければ発動しない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// 通常攻撃でなければ発動しない
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		// ICD中は発動しない
		if char.StatusIsActive(icdKey) {
			return false
		}
		// ICDをセット
		char.AddStatus(icdKey, 600, true) // 10s

		// 単体ターゲット発動をキューに追加
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Halberd Proc",
			AttackTag:  attacks.AttackTagNone,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       dmg,
		}
		c.QueueAttack(ai, combat.NewSingleTargetHit(trg.Key()), 0, 1)

		return false
	}, fmt.Sprintf("halberd-%v", char.Base.Key.String()))

	return w, nil
}
