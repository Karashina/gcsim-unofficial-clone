package xingqiu

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstHitmark = 18
	burstKey     = "xingqiuburst"
	burstICDKey  = "xingqiu-burst-icd"
)

func init() {
	burstFrames = frames.InitAbilSlice(40)
	burstFrames[action.ActionAttack] = 33
	burstFrames[action.ActionSkill] = 33
	burstFrames[action.ActionDash] = 33
	burstFrames[action.ActionJump] = 33
}

/**
波ごとに召喚される水の剣の数は特定のパターンに従い、通常2本と3本が交互に発動する。
C6ではこれが強化され、2 → 3 → 5… のパターンで繰り返す。

召喚された水の剣の波の間には約1秒の間隔があり、理論上の最大波数は15または18波。

各波の水の剣は1回の水元素付着が可能で、各剣は個別に会心判定が行われる。
**/

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 3ヒットごとに水元素を付与
	// 通常攻撃でトリガー
	// p=1の場合は発動時にも水元素を付与
	// どうやって？？ 0ダメージをトリガー？

	/** C2
	古華剣・裂雨の持続時間を3秒延長する。
	剣雨攻撃に命中した敵の水元素耐性を4秒間15%減少させる。
	**/
	dur := 15
	if c.Base.Cons >= 2 {
		dur += 3
	}
	dur *= 60
	c.AddStatus(burstKey, dur+33, true) // アニメーション用に33f追加
	c.applyOrbital(dur, burstHitmark)

	c.burstCounter = 0
	c.numSwords = 2
	c.nextRegen = false

	// c.CD[combat.BurstCD] = c.S.F + 20*60
	c.SetCD(action.ActionBurst, 20*60)
	c.ConsumeEnergy(3)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstHitmark,
		State:           action.BurstState,
	}, nil
}

func (c *char) summonSwordWave() {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Guhua Sword: Raincutter",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	// c.nextRegenがtrueかつ最初の剣の場合のみ
	var c2cb, c6cb func(a combat.AttackCB)
	if c.nextRegen {
		done := false
		c6cb = func(a combat.AttackCB) {
			if a.Target.Type() != targets.TargettableEnemy {
				return
			}
			if done {
				return
			}
			c.AddEnergy("xingqiu-c6", 3)
			done = true
		}
	}
	if c.Base.Cons >= 2 {
		icd := -1
		c2cb = func(a combat.AttackCB) {
			if c.Core.F < icd {
				return
			}

			e, ok := a.Target.(*enemy.Enemy)
			if !ok {
				return
			}

			icd = c.Core.F + 1
			c.Core.Tasks.Add(func() {
				e.AddResistMod(combat.ResistMod{
					Base:  modifier.NewBaseWithHitlag("xingqiu-c2", 4*60),
					Ele:   attributes.Hydro,
					Value: -0.15,
				})
			}, 1)
		}
	}

	snap := c.Snapshot(&ai)
	for i := 0; i < c.numSwords; i++ {
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				nil,
				0.5,
			),
			20,
			c2cb,
			c6cb,
		)
		c6cb = nil
		c.burstCounter++
	}

	// 次の波の剣の数を決定
	switch c.numSwords {
	case 2:
		c.numSwords = 3
		c.nextRegen = false
	case 3:
		if c.Base.Cons >= 6 {
			c.numSwords = 5
			c.nextRegen = true
		} else {
			c.numSwords = 2
			c.nextRegen = false
		}
	case 5:
		c.numSwords = 2
		c.nextRegen = false
	}

	c.AddStatus(burstICDKey, 60, true)
}
