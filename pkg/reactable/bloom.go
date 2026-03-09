package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
)

const DendroCoreDelay = 30

func (r *Reactable) TryBloom(a *combat.AttackEvent) bool {
	// 水開花、草開花、または激化開花の可能性がある
	if a.Info.Durability < ZeroDur {
		return false
	}
	var consumed reactions.Durability
	switch a.Info.Element {
	case attributes.Hydro:
		// この部分はやっかい。草系オーラがいずれか存在すれば開花が発生する
		// ので、3つ全てをチェックする必要がある...
		switch {
		case r.Durability[Dendro] > ZeroDur:
		case r.Durability[Quicken] > ZeroDur:
		case r.Durability[BurningFuel] > ZeroDur:
		default:
			return false
		}
		// reduce は1つの元素しかチェックしないので、激化も確認するために2回呼ぶ必要がある
		consumed = r.reduce(attributes.Dendro, a.Info.Durability, 0.5)
		f := r.reduce(attributes.Quicken, a.Info.Durability, 0.5)
		if f > consumed {
			consumed = f
		}
	case attributes.Dendro:
		if r.Durability[Hydro] < ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Hydro, a.Info.Durability, 2)
	default:
		return false
	}
	a.Info.Durability -= consumed
	a.Info.Durability = max(a.Info.Durability, 0)
	a.Reacted = true

	// ルナ開花のチェック: パーティメンバーに LB-Key があれば OnLunarBloom を発行
	// これは草原核生成とは独立している
	hasLBKey := false
	for _, char := range r.core.Player.Chars() {
		if char.StatusIsActive("LB-Key") {
			hasLBKey = true
			break
		}
	}
	if hasLBKey {
		r.core.Events.Emit(event.OnLunarBloom, r.self, a)
	}

	r.addBloomGadget(a)
	r.core.Events.Emit(event.OnBloom, r.self, a)
	return true
}

// tryQuickenBloom は触媒反応後にのみ呼ばれるべき（現在フレームの末尾にキューされる）
// この反応は水元素が存在するかチェックし、存在すれば開花反応を発生させる
func (r *Reactable) tryQuickenBloom(a *combat.AttackEvent) {
	if r.Durability[Quicken] < ZeroDur {
		// 健全性チェック；触媒直後（同フレーム）に激化が消されない限り実際には発生しない
		return
	}
	if r.Durability[Hydro] < ZeroDur {
		return
	}
	avail := r.Durability[Quicken]
	consumed := r.reduce(attributes.Hydro, avail, 2)
	r.Durability[Quicken] -= consumed

	// 激化開花の場合もルナ開花をチェック
	hasLBKey := false
	for _, char := range r.core.Player.Chars() {
		if char.StatusIsActive("LB-Key") {
			hasLBKey = true
			break
		}
	}
	if hasLBKey {
		r.core.Events.Emit(event.OnLunarBloom, r.self, a)
	}

	r.addBloomGadget(a)
	r.core.Events.Emit(event.OnBloom, r.self, a)
}

type DendroCore struct {
	*gadget.Gadget
	srcFrame  int
	CharIndex int
	// trueの場合、この草原核は詐欺の種（Nefer）である
	IsSeed bool
}

func (r *Reactable) addBloomGadget(a *combat.AttackEvent) {
	// タイミング競合を避けるため、スケジューリング時に詐欺の種かどうかを判定する
	isSeed := false
	for _, ch := range r.core.Player.Chars() {
		if ch.StatusIsActive("nefer-seed-convert") {
			isSeed = true
			break
		}
	}
	// クロージャ内で isSeed をキャプチャ
	capturedIsSeed := isSeed
	r.core.Tasks.Add(func() {
		t := NewDendroCore(r.core, r.self.Shape(), a, capturedIsSeed)
		r.core.Combat.AddGadget(t)
		r.core.Events.Emit(event.OnDendroCore, t, a)
		r.core.Log.NewEvent(
			"dendro core spawned",
			glog.LogElementEvent,
			a.Info.ActorIndex,
		).
			Write("src", t.Src()).
			Write("expiry", r.core.F+t.Duration).
			Write("is_seed", capturedIsSeed)
	}, DendroCoreDelay)
}

func NewDendroCore(c *core.Core, shp geometry.Shape, a *combat.AttackEvent, isSeed bool) *DendroCore {
	s := &DendroCore{
		srcFrame:  c.F,
		CharIndex: a.Info.ActorIndex,
		IsSeed:    isSeed,
	}

	circ, ok := shp.(*geometry.Circle)
	if !ok {
		panic("rectangle target hurtbox is not supported for dendro core spawning")
	}

	// 簡略化のため、種は全体から半径+0.5のランダム位置に生成される
	r := circ.Radius() + 0.5
	s.Gadget = gadget.New(c, geometry.CalcRandomPointFromCenter(circ.Pos(), r, r, c.Rand), 2, combat.GadgetTypDendroCore)
	s.Gadget.Duration = 300 // ??

	// 詐欺の種の動作デバッグ用の作成詳細ログ
	c.Log.NewEvent(
		"new dendro core created",
		glog.LogElementEvent,
		a.Info.ActorIndex,
	).Write("is_seed", isSeed).
		Write("src_frame", s.srcFrame)

	char := s.Core.Player.ByIndex(a.Info.ActorIndex)

	explode := func(reason string) func() {
		return func() {
			s.Core.Tasks.Add(func() {
				// 爆発試行と現在の IsSeed 状態をログ
				s.Core.Log.NewEvent(
					"dendro core explode attempt",
					glog.LogElementEvent,
					char.Index,
				).Write("is_seed", s.IsSeed).
					Write("src", s.Src())

				if s.IsSeed {
					// 詐欺の種は爆発しない
					return
				}
				// 開花攻撃
				ai, snap := NewBloomAttack(char, s, nil)
				ap := combat.NewCircleHitOnTarget(s, nil, 5)
				c.QueueAttackWithSnap(ai, snap, ap, 0)

				// 自傷ダメージ
				ai.Abil += reactions.SelfDamageSuffix
				ai.FlatDmg = 0.05 * ai.FlatDmg
				ap.SkipTargets[targets.TargettablePlayer] = false
				ap.SkipTargets[targets.TargettableEnemy] = true
				ap.SkipTargets[targets.TargettableGadget] = true
				c.QueueAttackWithSnap(ai, snap, ap, 0)

				c.Log.NewEvent(
					"dendro core "+reason,
					glog.LogElementEvent,
					char.Index,
				).Write("src", s.Src())
			}, 1)
		}
	}
	if !isSeed {
		s.Gadget.OnExpiry = explode("expired")
		s.Gadget.OnKill = explode("killed")
	} else {
		// 詐欺の種は期限切れや消滅時に爆発や反応を起こさない
		s.Gadget.OnExpiry = nil
		s.Gadget.OnKill = nil
	}

	return s
}

func (s *DendroCore) Tick() {
	// ガジェットのTickに必要
	s.Gadget.Tick()
}

func (s *DendroCore) HandleAttack(atk *combat.AttackEvent) float64 {
	s.Core.Events.Emit(event.OnGadgetHit, s, atk)
	s.Attack(atk, nil)
	return 0
}

func (s *DendroCore) Attack(atk *combat.AttackEvent, evt glog.Event) (float64, bool) {
	if atk.Info.Durability < ZeroDur {
		return 0, false
	}

	char := s.Core.Player.ByIndex(atk.Info.ActorIndex)
	// 炎/雷との接触のみが烈開花/超開花をそれぞれ発動する
	switch atk.Info.Element {
	case attributes.Electro:
		if s.IsSeed {
			// 詐欺の種は超開花を発動できない
			return 0, false
		}
		// 超開花は最も近い敵をターゲットにする
		// 小範囲AoEでプレイヤーにもダメージを与える
		s.Core.Tasks.Add(func() {
			ai, snap := NewHyperbloomAttack(char, s)
			// 半径15以内の最も近い敵にダメージをキュー
			enemy := s.Core.Combat.ClosestEnemyWithinArea(combat.NewCircleHitOnTarget(s.Gadget, nil, 15), nil)
			if enemy != nil {
				ap := combat.NewCircleHitOnTarget(enemy, nil, 1)
				s.Core.QueueAttackWithSnap(ai, snap, ap, 0)

				// 自傷ダメージもキュー
				ai.Abil += reactions.SelfDamageSuffix
				ai.FlatDmg = 0.05 * ai.FlatDmg
				ap.SkipTargets[targets.TargettablePlayer] = false
				ap.SkipTargets[targets.TargettableEnemy] = true
				ap.SkipTargets[targets.TargettableGadget] = true
				s.Core.QueueAttackWithSnap(ai, snap, ap, 0)
			}
		}, 60)

		s.Gadget.OnKill = nil
		s.Gadget.Kill()
		s.Core.Events.Emit(event.OnHyperbloom, s, atk)
		s.Core.Log.NewEvent(
			"hyperbloom triggered",
			glog.LogElementEvent,
			char.Index,
		).
			Write("dendro_core_char", s.CharIndex).
			Write("dendro_core_src", s.Gadget.Src())
	case attributes.Pyro:
		if s.IsSeed {
			// 詐欺の種は烈開花を発動できない
			return 0, false
		}
		// 烈開花を発動、AoE草元素ダメージ
		// 自傷ダメージ
		s.Core.Tasks.Add(func() {
			ai, snap := NewBurgeonAttack(char, s)
			ap := combat.NewCircleHitOnTarget(s, nil, 5)
			s.Core.QueueAttackWithSnap(ai, snap, ap, 0)

			// 自傷ダメージをキュー
			ai.Abil += reactions.SelfDamageSuffix
			ai.FlatDmg = 0.05 * ai.FlatDmg
			ap.SkipTargets[targets.TargettablePlayer] = false
			ap.SkipTargets[targets.TargettableEnemy] = true
			ap.SkipTargets[targets.TargettableGadget] = true
			s.Core.QueueAttackWithSnap(ai, snap, ap, 0)
		}, 1)

		s.Gadget.OnKill = nil
		s.Gadget.Kill()
		s.Core.Events.Emit(event.OnBurgeon, s, atk)
		s.Core.Log.NewEvent(
			"burgeon triggered",
			glog.LogElementEvent,
			char.Index,
		).
			Write("dendro_core_char", s.CharIndex).
			Write("dendro_core_src", s.Gadget.Src())
	default:
		return 0, false
	}

	return 0, false
}

const (
	BloomMultiplier      = 2
	BurgeonMultiplier    = 3
	HyperbloomMultiplier = 3
)

func NewBloomAttack(char *character.CharWrapper, src combat.Target, modify func(*combat.AttackInfo)) (combat.AttackInfo, combat.Snapshot) {
	em := char.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       char.Index,
		DamageSrc:        src.Key(),
		Element:          attributes.Dendro,
		AttackTag:        attacks.AttackTagBloom,
		ICDTag:           attacks.ICDTagBloomDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Abil:             string(reactions.Bloom),
		IgnoreDefPercent: 1,
	}
	if modify != nil {
		modify(&ai)
	}
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = BloomMultiplier * flatdmg
	return ai, snap
}

func NewBurgeonAttack(char *character.CharWrapper, src combat.Target) (combat.AttackInfo, combat.Snapshot) {
	em := char.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       char.Index,
		DamageSrc:        src.Key(),
		Element:          attributes.Dendro,
		AttackTag:        attacks.AttackTagBurgeon,
		ICDTag:           attacks.ICDTagBurgeonDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Abil:             string(reactions.Burgeon),
		IgnoreDefPercent: 1,
	}
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = BurgeonMultiplier * flatdmg
	return ai, snap
}

func NewHyperbloomAttack(char *character.CharWrapper, src combat.Target) (combat.AttackInfo, combat.Snapshot) {
	em := char.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       char.Index,
		DamageSrc:        src.Key(),
		Element:          attributes.Dendro,
		AttackTag:        attacks.AttackTagHyperbloom,
		ICDTag:           attacks.ICDTagHyperbloomDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Abil:             string(reactions.Hyperbloom),
		IgnoreDefPercent: 1,
	}
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = HyperbloomMultiplier * flatdmg
	return ai, snap
}

func (s *DendroCore) SetDirection(trg geometry.Point) {}
func (s *DendroCore) SetDirectionToClosestEnemy()     {}
func (s *DendroCore) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}
