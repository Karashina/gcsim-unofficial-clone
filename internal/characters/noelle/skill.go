package noelle

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const skillHitmark = 14

func init() {
	skillFrames = frames.InitAbilSlice(78)
	skillFrames[action.ActionAttack] = 12
	skillFrames[action.ActionCharge] = 13
	skillFrames[action.ActionSkill] = 14 // 元素爆発フレームを使用
	skillFrames[action.ActionBurst] = 14
	skillFrames[action.ActionDash] = 11
	skillFrames[action.ActionJump] = 11
	skillFrames[action.ActionWalk] = 43
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Breastplate",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagElementalArt,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           100,
		Element:            attributes.Geo,
		Durability:         50,
		Mult:               shieldDmg[c.TalentLvlSkill()],
		UseDef:             true,
		CanBeDefenseHalted: true,
	}
	snap := c.Snapshot(&ai)

	// まずシールドを追加
	defFactor := snap.Stats.TotalDEF()
	shieldhp := shieldFlat[c.TalentLvlSkill()] + shieldDef[c.TalentLvlSkill()]*defFactor
	c.Core.Player.Shields.Add(c.newShield(shieldhp, shield.NoelleSkill, 720))

	// シールドタイマーを有効化、終了時に爆発
	c.shieldTimer = c.Core.F + 720 // 12秒

	c.a4Counter = 0

	// 元素スキルの初撃も回復を発動可能
	cb := c.skillHealCB()

	// プレイヤー中心
	// 4凸に備えてchar queueを使用
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2),
			0,
			0,
			cb,
		)
	}, skillHitmark)

	// 4凸を処理
	if c.Base.Cons >= 4 {
		c.Core.Tasks.Add(func() {
			if c.shieldTimer == c.Core.F {
				// ダメージを与える
				c.explodeShield()
			}
		}, 720)
	}

	c.SetCDWithDelay(action.ActionSkill, 24*60, 6)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) skillHealCB() combat.AttackCBFunc {
	done := false
	return func(atk combat.AttackCB) {
		if atk.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		// 回復をチェック
		if c.Core.Player.Shields.Get(shield.NoelleSkill) != nil {
			var prob float64
			if c.Base.Cons >= 1 && c.StatModIsActive(burstBuffKey) {
				prob = 1
			} else {
				prob = healChance[c.TalentLvlSkill()]
			}
			if c.Core.Rand.Float64() < prob {
				// ターゲットを回復
				def := atk.AttackEvent.Snapshot.Stats.TotalDEF()
				heal := shieldHeal[c.TalentLvlSkill()]*def + shieldHealFlat[c.TalentLvlSkill()]
				c.Core.Player.Heal(info.HealInfo{
					Caller:  c.Index,
					Target:  -1,
					Message: "Breastplate (Attack)",
					Src:     heal,
					Bonus:   atk.AttackEvent.Snapshot.Stats[attributes.Heal],
				})
				done = true
			}
		}
	}
}

// 4凸:
// 護心の持続時間が終了した時、またはダメージにより破壊された時、
// 周囲の敵に攻撃力400%の岩元素ダメージを与える。
func (c *char) explodeShield() {
	c.shieldTimer = 0
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Breastplate (C4)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagElementalArt,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           100,
		Element:            attributes.Geo,
		Durability:         50,
		Mult:               4,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.15 * 60,
		CanBeDefenseHalted: true,
	}

	// プレイヤー中心
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 0, 0)
}
