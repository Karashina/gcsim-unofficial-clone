// パッケージ combat は戦闘関連の全機能を処理する:
//   - ターゲット追跡
//   - ターゲット選択
//   - ヒットボックスの衝突判定
//   - 攻撃キューイング
package combat

import (
	"math/rand"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
)

type CharHandler interface {
	CombatByIndex(int) Character
	ApplyHitlag(char int, factor, dur float64)
}

type Character interface {
	ApplyAttackMods(a *AttackEvent, t Target) []interface{}
}

type Handler struct {
	Opt
	enemies     []Target
	gadgets     []Gadget
	player      Target
	TotalDamage float64
	gccount     int
	keycount    targets.TargetKey
}

type Opt struct {
	Events        event.Eventter
	Tasks         task.Tasker
	Team          CharHandler
	Rand          *rand.Rand
	Debug         bool
	Log           glog.Logger
	DamageMode    bool
	DefHalt       bool
	EnableHitlag  bool
	DefaultTarget targets.TargetKey // index for default target
}

func New(opt Opt) *Handler {
	h := &Handler{
		Opt:      opt,
		keycount: 1,
	}
	h.enemies = make([]Target, 0, 5)
	h.gadgets = make([]Gadget, 0, 10)

	return h
}

func (h *Handler) nextkey() targets.TargetKey {
	h.keycount++
	return h.keycount - 1
}

func (h *Handler) Tick() {
	// 衝突判定は各オブジェクトのTick前に行う（衝突によりオブジェクトが削除される可能性があるため）
	// 敵とプレイヤーは衝突判定を行わない
	// ガジェットはプレイヤーと敵に対して衝突判定を行う
	for i := 0; i < len(h.gadgets); i++ {
		if h.gadgets[i] != nil && h.gadgets[i].CollidableWith(targets.TargettablePlayer) {
			if h.gadgets[i].WillCollide(h.player.Shape()) {
				h.gadgets[i].CollidedWith(h.player)
			}
		}
		// ガジェットが消えていないか確認
		if h.gadgets[i] != nil && h.gadgets[i].CollidableWith(targets.TargettableEnemy) {
			for j := 0; j < len(h.enemies) && h.gadgets[i] != nil; j++ {
				if h.gadgets[i].WillCollide(h.enemies[j].Shape()) {
					h.gadgets[i].CollidedWith(h.enemies[j])
				}
			}
		}
	}
	h.player.Tick()
	for _, v := range h.enemies {
		v.Tick()
	}
	for _, v := range h.gadgets {
		if v != nil {
			v.Tick()
		}
	}
	//TODO: 100 Tickごとのクリーンアップは妥当か？
	h.gccount++
	if h.gccount > 100 {
		n := 0
		for i, v := range h.gadgets {
			if v != nil {
				h.gadgets[n] = h.gadgets[i]
				n++
			}
		}
		h.gadgets = h.gadgets[:n]
		h.gccount = 0
	}
}
