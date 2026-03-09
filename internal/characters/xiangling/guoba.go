package xiangling

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/characters/faruzan"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

type panda struct {
	*gadget.Gadget
	*reactable.Reactable
	c     *char
	ai    combat.AttackInfo
	snap  combat.Snapshot
	timer int
}

func (c *char) newGuoba(ai combat.AttackInfo) *panda {
	p := &panda{
		ai:   ai,
		snap: c.Snapshot(&ai),
		c:    c,
	}
	player := c.Core.Combat.Player()
	pos := geometry.CalcOffsetPoint(
		player.Pos(),
		geometry.Point{Y: 1.3},
		player.Direction(),
	)
	p.Gadget = gadget.New(c.Core, pos, 0.2, combat.GadgetTypGuoba)
	p.Gadget.Duration = 438
	p.Reactable = &reactable.Reactable{}
	p.Reactable.Init(p, c.Core)

	return p
}

func (p *panda) Tick() {
	// reactableとgadgetの両方がtickを必要とするため
	p.Reactable.Tick()
	p.Gadget.Tick()
	p.timer++
	// グゥオパァーは100フレームごとに火を噴く
	// 最初の火噴きは126フレーム目だが、グゥオパァーは13フレーム後にスポーンするため実質113
	// その後は100フレームごと
	// TODO: この実装は改善が必要
	switch p.timer {
	case 103, 203, 303, 403: // 拡散ウィンドウ
		p.Core.Log.NewEvent("guoba self infusion applied", glog.LogElementEvent, p.c.Index).
			SetEnded(p.c.Core.F + infuseWindow + 1)
		p.Durability[reactable.Pyro] = infuseDurability
		p.Core.Tasks.Add(func() {
			p.Durability[reactable.Pyro] = 0
		}, infuseWindow+1) // +1（元素付与ウィンドウが包含的なため）
		p.SetDirectionToClosestEnemy()
		// ライブでの挙動に合わせて事前にキューに追加
		p.breath()
	}
}

func (p *panda) breath() {
	// 固有天賦1を前提とする
	p.Core.QueueAttackWithSnap(
		p.ai,
		p.snap,
		combat.NewCircleHitOnTargetFanAngle(p, nil, p.c.guobaFlameRange, 60),
		10,
		p.c.c1,
		p.c.particleCB,
	)
}

func (p *panda) Type() targets.TargettableType { return targets.TargettableGadget }

func (p *panda) HandleAttack(atk *combat.AttackEvent) float64 {
	p.Core.Events.Emit(event.OnGadgetHit, p, atk)
	p.Attack(atk, nil)
	return 0
}

func (p *panda) Attack(atk *combat.AttackEvent, evt glog.Event) (float64, bool) {
	if atk.Info.AttackTag != attacks.AttackTagElementalArt {
		return 0, false
	}
	// 炎元素ウィンドウをチェック
	if p.Durability[reactable.Pyro] <= reactable.ZeroDur {
		return 0, false
	}

	// ダメージは受けず、スクロースEまたはファルザンEによる拡散反応のみトリガー
	switch p.Core.Player.Chars()[atk.Info.ActorIndex].Base.Key {
	case keys.Sucrose:
		p.Core.Log.NewEvent("guoba hit by sucrose E", glog.LogCharacterEvent, p.c.Index)
	case keys.Faruzan:
		if atk.Info.Abil != faruzan.VortexAbilName {
			return 0, false
		}
		p.Core.Log.NewEvent("guoba hit by faruzan pressurized collapse", glog.LogCharacterEvent, p.c.Index)
	default:
		return 0, false
	}
	// スクロースEとファルザンEはガジェットに対して50元素量
	atk.Info.Durability = 50

	// スクロース/ファルザンEのゲージに合わせて元素量を調整
	oldDur := p.Durability[reactable.Pyro]
	p.Durability[reactable.Pyro] = infuseDurability
	p.React(atk)
	// 元素量を事後に復元
	p.Durability[reactable.Pyro] = oldDur

	return 0, false
}
