package beidou

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstHitmark = 28
	burstKey     = "beidouburst"
	burstICDKey  = "beidou-burst-icd"
)

func init() {
	burstFrames = frames.InitAbilSlice(58)
	burstFrames[action.ActionAttack] = 55
	burstFrames[action.ActionDash] = 48
	burstFrames[action.ActionJump] = 48
	burstFrames[action.ActionSwap] = 46
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Stormbreaker (Q)",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Electro,
		Durability:         100,
		Mult:               burstonhit[c.TalentLvlBurst()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.1 * 60,
		CanBeDefenseHalted: false,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4),
		burstHitmark,
		burstHitmark,
	)

	dur := 15 * 60
	// 北斗の爆発はヒットラグで延長されない
	c.AddStatus(burstKey, dur, false)

	procAI := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Stormbreak Proc (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burstproc[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&procAI)
	c.burstAtk = &combat.AttackEvent{
		Info:     procAI,
		Snapshot: snap,
	}

	if c.Base.Cons >= 1 {
		// シールドを作成
		c.Core.Player.Shields.Add(&shield.Tmpl{
			ActorIndex: c.Index,
			Target:     -1,
			Src:        c.Core.F,
			ShieldType: shield.BeidouC1,
			Name:       "Beidou C1",
			HP:         .16 * c.MaxHP(),
			Ele:        attributes.Electro,
			Expires:    c.Core.F + dur,
		})
	}

	// ヒットマーク後に適用
	if c.Base.Cons >= 6 {
		for i := 30; i <= dur; i += 30 {
			c.Core.Tasks.Add(func() {
				enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), nil)
				for _, v := range enemies {
					v.AddResistMod(combat.ResistMod{
						Base:  modifier.NewBaseWithHitlag("beidouc6", 90),
						Ele:   attributes.Electro,
						Value: -0.15,
					})
				}
			}, i)
		}
	}

	c.ConsumeEnergy(6)
	c.SetCD(action.ActionBurst, 1200)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) burstProc() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		t := args[0].(combat.Target)
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// 攻撃をトリガーしたキャラクターがまだフィールドにいることを確認
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if !c.StatusIsActive(burstKey) {
			return false
		}
		if c.StatusIsActive(burstICDKey) {
			c.Core.Log.NewEvent("beidou Q (active) on icd", glog.LogCharacterEvent, c.Index)
			return false
		}

		// 最初のターゲットから連鎖攻撃をトリガー
		atk := *c.burstAtk
		atk.SourceFrame = c.Core.F
		atk.Pattern = combat.NewSingleTargetHit(t.Key())
		cb := c.chain(c.Core.F, 1)
		if cb != nil {
			atk.Callbacks = append(atk.Callbacks, cb)
		}
		c.Core.QueueAttackEvent(&atk, 1)

		c.Core.Log.NewEvent("beidou Q proc'd", glog.LogCharacterEvent, c.Index).
			Write("char", ae.Info.ActorIndex).
			Write("attack tag", ae.Info.AttackTag)

		// このICDはおそらく設置物に紐づいているため、ヒットラグで延長されない
		c.AddStatus(burstICDKey, 60, false)
		return false
	}, "beidou-burst")
}

func (c *char) chain(src, count int) combat.AttackCBFunc {
	if c.Base.Cons >= 2 && count == 5 {
		return nil
	}
	if c.Base.Cons <= 1 && count == 3 {
		return nil
	}
	return func(a combat.AttackCB) {
		// 命中時に次のターゲットを決定
		next := c.Core.Combat.RandomEnemyWithinArea(combat.NewCircleHitOnTarget(a.Target, nil, 8), func(t combat.Enemy) bool {
			return a.Target.Key() != t.Key()
		})
		if next == nil {
			// 自身以外のターゲットがなければ何もしない
			return
		}
		// 次のターゲットへの攻撃をキューに追加
		atk := *c.burstAtk
		atk.SourceFrame = src
		atk.Pattern = combat.NewSingleTargetHit(next.Key())
		cb := c.chain(src, count+1)
		if cb != nil {
			atk.Callbacks = append(atk.Callbacks, cb)
		}
		c.Core.QueueAttackEvent(&atk, 1)
	}
}
