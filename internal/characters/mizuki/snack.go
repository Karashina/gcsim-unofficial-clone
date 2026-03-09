package mizuki

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/avatar"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
)

const (
	snackDmgName              = "Munen Shockwave"
	snackHealName             = "Snack Pick-Up"
	snackDurability           = 25
	snackDmgRadius            = 4
	snackHealTriggerHpRatio   = 0.7
	snackDuration             = 4 * 60
	snackSize                 = 2.5
	snackSizeMizukiMultiplier = 2.1 // Assumption
	snackCantTriggerDuration  = 0.3 * 60
)

type snack struct {
	*gadget.Gadget
	char         *char
	attackInfo   combat.AttackInfo
	snapshot     combat.Snapshot
	pattern      combat.AttackPattern
	allowPickupF int
}

func newSnack(c *char, pos geometry.Point) *snack {
	p := &snack{
		char: c,
		attackInfo: combat.AttackInfo{
			ActorIndex:   c.Index,
			Abil:         snackDmgName,
			AttackTag:    attacks.AttackTagElementalBurst,
			ICDTag:       attacks.ICDTagElementalBurst,
			ICDGroup:     attacks.ICDGroupDefault,
			StrikeType:   attacks.StrikeTypeDefault,
			Element:      attributes.Anemo,
			Durability:   snackDurability,
			Mult:         snackDMG[c.TalentLvlBurst()],
			HitlagFactor: 0.05,
		},
		pattern:      combat.NewCircleHitOnTarget(pos, nil, snackDmgRadius),
		allowPickupF: c.Core.F + snackCantTriggerDuration,
	}
	p.snapshot = c.Snapshot(&p.attackInfo)

	// Dreamdrifter状態中の瑞希が確実に取得できるようおやつサイズを拡大
	// この状態中は瑞希の取得範囲が増加するため。
	// https://docs.google.com/spreadsheets/d/1UU0EVPBatEndl4GRZyIs8Ix8O3kcZUDAwOHqM8_jQJw/edit?gid=339012102#gid=339012102
	p.Gadget = gadget.New(c.Core, pos, snackSize*snackSizeMizukiMultiplier, combat.GadgetTypYumemiSnack)
	p.Gadget.Duration = snackDuration
	c.Core.Combat.AddGadget(p)

	p.Gadget.CollidableTypes[targets.TargettablePlayer] = true
	p.Gadget.OnExpiry = func() {
		p.explode()
		p.Core.Log.NewEvent("Snack exploded by itself", glog.LogCharacterEvent, c.Index)
	}
	p.Gadget.OnCollision = func(target combat.Target) {
		if _, ok := target.(*avatar.Player); !ok {
			return
		}
		if p.Core.F < p.allowPickupF {
			return
		}

		// デフォルトサイズは拡大済み。拡大サイズはDreamdrifter状態の瑞希のみ有効なため、
		// そうでない場合は実際のサイズで衝突判定を行う
		if !c.StatusIsActive(dreamDrifterStateKey) && !p.collidesWithActiveCharacterDefaultSize() {
			return
		}
		p.onPickedUp()
	}

	p.Core.Log.NewEvent("Snack spawned", glog.LogCharacterEvent, c.Index).
		Write("x", pos.X).
		Write("y", pos.Y)
	return p
}

func (p *snack) collidesWithActiveCharacterDefaultSize() bool {
	defaultSize := combat.NewCircleHitOnTarget(p.Gadget, nil, snackSize)
	return p.Core.Combat.Player().IsWithinArea(defaultSize)
}

func (p *snack) onPickedUp() {
	var heal bool
	var dmg bool

	mizuki := p.char
	activeChar := p.Core.Player.ActiveChar()

	// 4凸はダメージと回復の両方を発動
	if mizuki.Base.Cons >= 4 {
		dmg = true
		heal = true
	} else {
		// アクティブキャラのHPが70%以下なら回復、それ以外はダメージ
		dmg = activeChar.CurrentHP() > (activeChar.MaxHP() * snackHealTriggerHpRatio)
		heal = !dmg
	}

	p.Core.Log.NewEvent("Picked up snack", glog.LogCharacterEvent, activeChar.Index).
		Write("heal", heal).
		Write("dmg", dmg)

	if dmg {
		p.explode()
	}

	if heal {
		// 瑞希への回復量は2倍
		healMultiplier := 1.0
		if activeChar.Index == mizuki.Index {
			healMultiplier = 2.0
		}
		mizuki.Core.Player.Heal(info.HealInfo{
			Caller:  mizuki.Index,
			Target:  activeChar.Index,
			Message: snackHealName,
			Src:     ((mizuki.Stat(attributes.EM) * snackHealEM[mizuki.TalentLvlBurst()]) + snackHealFlat[mizuki.TalentLvlBurst()]) * healMultiplier,
			Bonus:   mizuki.Stat(attributes.Heal),
		})
	}

	// 4凸は瑞希に元素エネルギーを最大4回まで5回復
	mizuki.c4()

	p.Kill()
}

func (p *snack) explode() {
	p.Core.QueueAttackWithSnap(p.attackInfo, p.snapshot, p.pattern, 0)
}

func (p *snack) HandleAttack(atk *combat.AttackEvent) float64 {
	// プレイヤーとの衝突か期限切れのみがこれに影響する
	return 0
}

func (p *snack) SetDirection(trg geometry.Point) {}
func (p *snack) SetDirectionToClosestEnemy()     {}
func (p *snack) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}
