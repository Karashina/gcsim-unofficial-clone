package lynette

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

var (
	skillPressFrames   []int
	skillHoldEndFrames []int
)

const (
	skillCD = 12 * 60

	skillPressHitmark        = 28
	skillPressAlignedHitmark = 58
	skillPressC6Start        = 17
	skillPressCDStart        = 26

	skillHoldShadowsignStart    = 17
	skillHoldShadowsignInterval = 0.1 * 60
	skillHoldEndHitmark         = 16
	skillHoldEndAlignedHitmark  = 44
	skillHoldEndC6Start         = 14 - 9 // 9f before cd start
	skillHoldEndCDStart         = 14

	particleICDKey = "lynette-particle-icd"
	particleICD    = 0.6 * 60
	particleCount  = 4

	skillTag           = "lynette-shadowsign"
	skillAlignedICDKey = "lynette-aligned-icd"
	skillAlignedICD    = 10 * 60
)

func init() {
	// 元素スキル（単押し）
	skillPressFrames = frames.InitAbilSlice(58) // E -> Walk
	skillPressFrames[action.ActionAttack] = 43
	skillPressFrames[action.ActionSkill] = 44
	skillPressFrames[action.ActionBurst] = 45
	skillPressFrames[action.ActionDash] = 44
	skillPressFrames[action.ActionJump] = 43
	skillPressFrames[action.ActionSwap] = 42

	// 元素スキル（長押し）
	skillHoldEndFrames = frames.InitAbilSlice(41) // Hold E -> Walk
	skillHoldEndFrames[action.ActionAttack] = 29
	skillHoldEndFrames[action.ActionSkill] = 29
	skillHoldEndFrames[action.ActionBurst] = 29
	skillHoldEndFrames[action.ActionDash] = 28
	skillHoldEndFrames[action.ActionJump] = 31
	skillHoldEndFrames[action.ActionSwap] = 27
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"]
	if hold > 0 {
		if hold > 150 {
			hold = 150
		}
		// スキル状態の最小持続時間: 紵35f
		// スキル状態の最大持続時間: 紵184f
		// -> 34fのオフセットで 1 <= hold <= 150
		return c.skillHold(hold + 34), nil
	}
	return c.skillPress(), nil
}

func (c *char) skillPress() action.Info {
	// 単押し攻撃と整合攻撃
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			c.skillAI,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.5}, 1.8, 4.5),
			0,
			0,
			c.particleCB,
			c.makeSkillHealAndDrainCB(),
		)
		c.skillAligned(skillPressAlignedHitmark - skillPressHitmark) // TODO: 整列CDのチェックがいつ発生するか不明
	}, skillPressHitmark)

	c.QueueCharTask(c.c6, skillPressC6Start)
	c.AddStatus(weaponoutkey, 20*60, false)

	c.SetCDWithDelay(action.ActionSkill, skillCD, skillPressCDStart)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionSwap],
		State:           action.SkillState,
	}
}

func (c *char) skillHold(duration int) action.Info {
	// 影標の発動
	c.QueueCharTask(func() {
		c.shadowsignSrc = c.Core.F
		c.applyShadowsign(c.Core.F)
	}, skillHoldShadowsignStart)

	// 影標の終了、長押し攻撃と整合攻撃
	c.QueueCharTask(func() {
		c.clearShadowSign()
		c.shadowsignSrc = -1 // スキル終了のためティックをキャンセル

		c.Core.QueueAttack(
			c.skillAI,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.5}, 1.8, 5),
			0,
			0,
			c.particleCB,
			c.makeSkillHealAndDrainCB(),
		)
		c.skillAligned(skillHoldEndAlignedHitmark - skillHoldEndHitmark) // TODO: 整列CDのチェックがいつ発生するか不明
	}, duration+skillHoldEndHitmark)

	c.QueueCharTask(c.c6, duration+skillHoldEndC6Start)

	c.SetCDWithDelay(action.ActionSkill, skillCD, duration+skillHoldEndCDStart)

	return action.Info{
		Frames:          func(next action.Action) int { return duration + skillHoldEndFrames[next] },
		AnimationLength: duration + skillHoldEndFrames[action.InvalidAction],
		CanQueueAfter:   duration + skillHoldEndFrames[action.ActionSwap],
		State:           action.SkillState,
	}
}

func (c *char) applyShadowsign(src int) func() {
	return func() {
		if src != c.shadowsignSrc {
			return
		}

		c.clearShadowSign()

		// 最も近い敵に影標を適用
		// TODO: 本来は最も近いではなく最高スコアの敵を選択すべき
		enemy := c.Core.Combat.ClosestEnemyWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8), nil)
		enemy.SetTag(skillTag, 1)

		// 次の影標適用をキューに追加
		c.QueueCharTask(c.applyShadowsign(src), skillHoldShadowsignInterval)
	}
}

// 全ての敵から影標を解除
func (c *char) clearShadowSign() {
	for _, t := range c.Core.Combat.Enemies() {
		if e, ok := t.(*enemy.Enemy); ok {
			e.SetTag(skillTag, 0)
		}
	}
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, particleICD, true)
	c.Core.QueueParticle(c.Base.Key.String(), particleCount, attributes.Anemo, c.ParticleDelay)
}

func (c *char) skillAligned(hitmark int) {
	if c.StatusIsActive(skillAlignedICDKey) {
		return
	}
	c.AddStatus(skillAlignedICDKey, skillAlignedICD, true)

	c.Core.QueueAttack(
		c.skillAlignedAI,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.3}, 1.2, 4.5),
		hitmark,
		hitmark,
	)
}

// 「エニグマスラスト」が敵に命中すると、リネットのHPが最大HPに基づき回復し、
// その後4秒間、毎秒一定量のHPが消耗される。
func (c *char) makeSkillHealAndDrainCB() combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		// 回復
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "Enigmatic Feint",
			Src:     0.25 * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})

		// HP消耗
		// TODO: これは本当に重複する？
		// TODO: 間隔の正確なフレーム
		c.QueueCharTask(c.skillDrain(0), 1*60)
	}
}

func (c *char) skillDrain(count int) func() {
	return func() {
		count++
		if count == 4 {
			return
		}
		c.Core.Player.Drain(info.DrainInfo{
			ActorIndex: c.Index,
			Abil:       "Enigmatic Feint",
			Amount:     0.06 * c.CurrentHP(),
		})
		c.QueueCharTask(c.skillDrain(count), 1*60)
	}
}
