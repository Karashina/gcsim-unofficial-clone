package swordofnarzissenkreuz

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
	core.RegisterWeaponFunc(keys.SwordOfNarzissenkreuz, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey  = "swordofnarzissenkreuz-icd"
	icd     = 12 * 60
	hitmark = 0.1 * 60 // 概算値
)

// 装備キャラがアルケーを持たない場合: 通常攻撃、重撃、または落下攻撃が命中した時、
// プネウマまたはウーシアのエネルギーブラストが放たれ、攻撃力の160/200/240/280/320%のダメージを与える。
// この効果は12秒毎に1回発動可能。エネルギーブラストのタイプは Sword of Narzissenkreuz の現在のタイプによる。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// TODO: アルケーは現時点で何も影響しない
	// 0: プネウマ、 1: ウーシア
	arkhe, ok := p.Params["arkhe"]
	if !ok {
		arkhe = 1
	}
	if arkhe < 0 {
		arkhe = 0
	}
	if arkhe > 1 {
		arkhe = 1
	}
	c.Log.NewEvent("swordofnarzissenkreuz arkhe", glog.LogWeaponEvent, char.Index).
		Write("arkhe", arkhe)

	// キャラがアルケーを持つ場合はイベント登録しない
	if char.HasArkhe {
		return w, nil
	}

	mult := 1.2 + float64(r)*0.4

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		trg := args[0].(combat.Target)

		// 通常攻撃、重撃、落下攻撃のダメージ時のみ発動
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		case attacks.AttackTagPlunge:
		default:
			return false
		}

		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, icd, true)

		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Sword of Narzissenkreuz",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       mult,
		}
		// ヒットマークのタイミングはヒットラグの影響を受ける
		char.QueueCharTask(func() {
			c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 0)
		}, hitmark)

		return false
	}, fmt.Sprintf("swordofnarzissenkreuz-%v", char.Base.Key.String()))

	return w, nil
}
