package endoftheline

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
)

func init() {
	core.RegisterWeaponFunc(keys.EndOfTheLine, NewWeapon)
}

type Weapon struct {
	Index     int
	procCount int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキル使用後、Flowrider効果を発動し、攻撃が敵に命中した時に攻撃力80%分の範囲ダメージを与える。
// Flowriderは15秒後または範囲ダメージを3回与えた後に解除される。
// 範囲ダメージは2秒毎に1回のみ発動可能。
// Flowriderは12秒毎に1回発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	flowriderDmg := 0.60 + float64(r)*0.20
	const effectKey = "endoftheline-effect"
	const effectIcdKey = "endoftheline-effect-icd"
	const dmgIcdKey = "endoftheline-dmg-icd"

	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		// 装備者がフィールド外なら何もしない
		if c.Player.Active() != char.Index {
			return false
		}
		// FlowriderがICD中なら何もしない
		if char.StatusIsActive(effectIcdKey) {
			return false
		}
		// Flowriderステータスを追加し、FlowriderのICDを発動
		char.AddStatus(effectKey, 15*60, true)
		char.AddStatus(effectIcdKey, 12*60, true)
		// ICDをリセット
		if char.StatusIsActive(dmgIcdKey) {
			char.DeleteStatus(dmgIcdKey)
		}
		// 発動回数をリセット
		w.procCount = 0
		return false
	}, fmt.Sprintf("endoftheline-effect-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		// 装備者がフィールド外なら何もしない
		if c.Player.Active() != char.Index {
			return false
		}
		// 装備者からの攻撃でなければ何もしない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// Flowriderがアクティブでなければ何もしない
		if !char.StatusIsActive(effectKey) {
			return false
		}
		// FlowriderダメージがICD中なら何もしない
		if char.StatusIsActive(dmgIcdKey) {
			return false
		}
		// Flowrider発動をキューに追加
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "End of the Line Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       flowriderDmg,
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 2.5), 0, 1)

		w.procCount++
		c.Log.NewEvent("endoftheline proc", glog.LogWeaponEvent, char.Index).
			Write("procCount", w.procCount)
		if w.procCount == 3 {
			char.DeleteStatus(effectKey)
		} else {
			char.AddStatus(dmgIcdKey, 2*60, true)
		}

		return false
	}, fmt.Sprintf("endoftheline-dmg-%v", char.Base.Key.String()))

	return w, nil
}
