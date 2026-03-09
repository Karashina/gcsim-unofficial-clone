package debateclub

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
	core.RegisterWeaponFunc(keys.DebateClub, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキル使用後、通常攻撃または重撃が命中すると追加で攻撃力の60/75/90/105/120%の範囲ダメージを与える。
// 効果は15秒間持続。ダメージは3秒に1回のみ発生。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	const effectKey = "debate-club-effect"
	const icdKey = "debate-club-icd"
	dmg := 0.45 + float64(r)*0.15

	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		// 武器装備者のスキルでなければ発動しない
		if c.Player.Active() != char.Index {
			return false
		}
		// ICDをリセット
		char.DeleteStatus(icdKey)
		// debate club効果を追加
		char.AddStatus(effectKey, 900, true) // 15s
		return false
	}, fmt.Sprintf("debate-club-activation-%v", char.Base.Key.String()))

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
		// 討論クラブの効果がアクティブでなければ発動しない
		if !char.StatusIsActive(effectKey) {
			return false
		}
		// 通常攻撃または重撃でなければ発動しない
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// ICD中は発動しない
		if char.StatusIsActive(icdKey) {
			return false
		}
		// ICDをセット
		char.AddStatus(icdKey, 180, true) // 3s

		// 発動をキューに追加
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Debate Club Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       dmg,
		}
		c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 1)

		return false
	}, fmt.Sprintf("debate-club-proc-%v", char.Base.Key.String()))

	return w, nil
}
