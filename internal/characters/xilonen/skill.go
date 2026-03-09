package xilonen

import (
	"fmt"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillHitmarks   = 6
	samplerInterval = 0.3 * 60

	skilRecastCD     = "xilonen-e-recast-cd"
	particleICDKey   = "xilonen-particle-icd"
	samplerShredKey  = "xilonen-e-shred"
	activeSamplerKey = "xilonen-samplers-activated"
)

func init() {
	skillFrames = frames.InitAbilSlice(20)
	skillFrames[action.ActionAttack] = 19
	skillFrames[action.ActionDash] = 15
	skillFrames[action.ActionJump] = 15
	skillFrames[action.ActionSwap] = 19
}

func (c *char) reduceNightsoulPoints(val float64) {
	if c.StatusIsActive(c6key) {
		return
	}

	c.nightsoulState.ConsumePoints(val * c.nightsoulConsumptionMul())

	// 通常攻撃/落下攻撃中はナイトソウルを終了しない
	switch c.Core.Player.CurrentState() {
	case action.NormalAttackState, action.PlungeAttackState:
		return
	}

	if c.nightsoulState.Points() < 0.001 {
		c.exitNightsoul()
	}
}

func (c *char) canUseNightsoul() bool {
	return c.nightsoulState.Points() >= 0.001 || c.StatusIsActive(c6key)
}

func (c *char) enterNightsoul() {
	c.nightsoulSrc = c.Core.F
	c.nightsoulPointReduceTask(c.nightsoulSrc)
	c.NormalHitNum = rollerHitNum
	c.NormalCounter = 0

	duration := int(9 * 60 * c.nightsoulDurationMul())
	c.nightsoulState.EnterTimedBlessing(45, duration, c.exitNightsoul)
	c.skillLastStamF = c.Core.Player.LastStamUse
	c.Core.Player.LastStamUse = math.MaxInt
	// 2凸以上の場合はタスクをキューしない
	if c.Base.Cons < 2 && c.samplersConverted < 3 {
		c.activeGeoSampler(c.nightsoulSrc)()
	}
}

func (c *char) exitNightsoul() {
	if !c.nightsoulState.HasBlessing() {
		return
	}
	c.nightsoulState.ExitBlessing()
	c.nightsoulState.ClearPoints()
	c.nightsoulSrc = -1
	c.exitStateSrc = -1
	c.SetCD(action.ActionSkill, 7*60)
	c.NormalHitNum = normalHitNum
	c.NormalCounter = 0
	c.Core.Player.LastStamUse = c.skillLastStamF
	c.DeleteStatus(c6key)
}

func (c *char) nightsoulPointReduceTask(src int) {
	const tickInterval = .1
	c.QueueCharTask(func() {
		if c.nightsoulSrc != src {
			return
		}
		// 6fごとに0.5ポイント消費、秒間5ポイント
		c.reduceNightsoulPoints(0.5)
		c.nightsoulPointReduceTask(src)
	}, 60*tickInterval)
}

func (c *char) applySamplerShred(ele attributes.Element, enemies []combat.Enemy) {
	for _, e := range enemies {
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(fmt.Sprintf("%v-%v", samplerShredKey, ele.String()), 60),
			Ele:   ele,
			Value: -skillShred[c.TalentLvlSkill()],
		})
	}
}

func (c *char) activeGeoSampler(src int) func() {
	return func() {
		// Xilonenは3つのサンプラーしか持たない；3つ全て変換された場合、岩元素サンプラーは残らない
		if c.samplersConverted >= 3 {
			return
		}

		if c.Base.Cons < 2 {
			if c.nightsoulSrc != src {
				return
			}
			if !c.nightsoulState.HasBlessing() {
				return
			}
			if c.StatusIsActive(activeSamplerKey) {
				// activeSamplersに移行
				return
			}
		}
		enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), nil)
		c.applySamplerShred(attributes.Geo, enemies)
		c.QueueCharTask(c.activeGeoSampler(src), samplerInterval)
	}
}

func (c *char) activeSamplers(src int) func() {
	return func() {
		if c.sampleSrc != src {
			return
		}
		if !c.StatusIsActive(activeSamplerKey) {
			return
		}

		enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), nil)
		for ele := range c.shredElements {
			// 2凸以上の場合、岩元素は常時有効のためスキップ
			if ele == attributes.Geo && c.Base.Cons >= 2 {
				continue
			}
			c.applySamplerShred(ele, enemies)
		}

		// QueueCharTaskはアクティブキャラクターで呼び出す必要がある
		active := c.Core.Player.ActiveChar()
		active.QueueCharTask(c.activeSamplers(src), samplerInterval)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() { // canUseNightsoulを使用しない
		c.exitNightsoul()
		return action.Info{
			Frames:          func(_ action.Action) int { return 1 },
			AnimationLength: 1,
			CanQueueAfter:   1, // 最速キャンセル
			State:           action.SkillState,
		}, nil
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Yohual's Scratch",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagElementalArt,
		AdditionalTags:     []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypePierce,
		Element:            attributes.Geo,
		Durability:         25,
		HitlagFactor:       0.01,
		Mult:               skillDMG[c.TalentLvlSkill()],
		UseDef:             true,
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}
	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: 1.0},
		0.8,
	)
	c.Core.QueueAttack(ai, ap, skillHitmarks, skillHitmarks, c.particleCB)
	c.AddStatus(skilRecastCD, 60, true)

	if c.Core.Player.Stam >= 15 {
		c.Core.Player.RestoreStam(5)
	} else {
		// 15に合わせる
		c.Core.Player.RestoreStam(15 - c.Core.Player.Stam)
	}

	c.enterNightsoul()
	c.c4()

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.5*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Geo, c.ParticleDelay)
}
