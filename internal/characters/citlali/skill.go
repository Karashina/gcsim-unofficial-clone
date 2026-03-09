package citlali

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/avatar"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

const (
	itzpapaInterval           = 59
	obsidianTzitzimitlHitmark = 20
	particleICDKey            = "citlali-particle-icd"
	opalFireStateKey          = "opal-fire-state"
	frostFallAbil             = "Frostfall Storm DMG"
)

var (
	skillFrames []int
)

func init() {
	skillFrames = frames.InitAbilSlice(50) // E -> Walk
	skillFrames[action.ActionAttack] = 42
	skillFrames[action.ActionCharge] = 42
	skillFrames[action.ActionBurst] = 41
	skillFrames[action.ActionDash] = 49
	skillFrames[action.ActionJump] = 49
	skillFrames[action.ActionSwap] = 41
}

func (c *char) Skill(_ map[string]int) (action.Info, error) {
	// 初撃を実行
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Obsidian Tzitzimitl DMG",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Cryo,
		Durability:     25,
		Mult:           skill[c.TalentLvlSkill()],
		HitlagFactor:   0.01,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6),
		obsidianTzitzimitlHitmark,
		obsidianTzitzimitlHitmark,
		c.particleCB,
	)

	// ディレイ付きで実行
	c.QueueCharTask(func() {
		c.SetCD(action.ActionSkill, 16*60)

		// Itzpapaを召喚し、Opal Fire状態が発動可能か即座に確認
		c.nightsoulState.EnterTimedBlessing(c.nightsoulState.Points()+24, 20*60, c.exitNightsoul)
		c.tryEnterOpalFireState()

		// 氷元素を付与
		player, ok := c.Core.Combat.Player().(*avatar.Player)
		if !ok {
			panic("target 0 should be Player but is not!!")
		}
		player.ApplySelfInfusion(attributes.Cryo, 25, 0.1*60)
	}, 18)

	c.QueueCharTask(c.addShield, 37)

	// 即時実行
	if c.Base.Cons >= 1 {
		c.numStellarBlades = 10
	}

	if c.Base.Cons >= 6 {
		currentPoints := c.nightsoulState.Points()
		c.nightsoulState.ClearPoints()
		c.numC6Stacks = min(maxC6Stacks, currentPoints)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionBurst],
		State:           action.SkillState,
	}, nil
}

func (c *char) exitNightsoul() {
	c.numC6Stacks = 0
	c.numStellarBlades = 0
	c.opalFireSrc = -1
	c.DeleteStatus(opalFireStateKey)
	c.nightsoulState.ExitBlessing()
}

func (c *char) generateNightsoulPoints(amount float64) {
	c.nightsoulState.GeneratePoints(amount)
	c.tryEnterOpalFireState()
}

// イベント購読を避けるため、CitlaliがNSポイントを獲得するたびにOpal Fireの発動を試みる
func (c *char) tryEnterOpalFireState() {
	if !c.nightsoulState.HasBlessing() {
		return
	}
	if c.nightsoulState.Points() < 50 && c.Base.Cons < 6 {
		return
	}
	// Opal Fire状態の発動または再発動の場合
	if c.StatusIsActive(opalFireStateKey) {
		return
	}
	c.AddStatus(opalFireStateKey, -1, false)
	c.opalFireSrc = c.Core.F
	c.itzpapaHitTask(c.opalFireSrc)
	c.nightsoulPointReduceTask(c.opalFireSrc)
}

func (c *char) nightsoulPointReduceTask(src int) {
	const tickInterval = .1
	c.QueueCharTask(func() {
		if c.opalFireSrc != src {
			return
		}
		if !c.StatusIsActive(opalFireStateKey) {
			return
		}

		// 6fごとに0.8ポイント減少（毎秒8ポイント）
		prev := c.nightsoulState.Points()
		c.nightsoulState.ConsumePoints(0.8)
		if c.Base.Cons >= 6 {
			diff := prev - c.nightsoulState.Points()
			c.numC6Stacks = min(maxC6Stacks, c.numC6Stacks+diff)
		}
		if c.nightsoulState.Points() < 0.001 && c.Base.Cons < 6 {
			c.opalFireSrc = -1
			c.DeleteStatus(opalFireStateKey)
			return
		}

		c.nightsoulPointReduceTask(src)
	}, 60*tickInterval)
}

func (c *char) itzpapaHitTask(src int) {
	c.QueueCharTask(func() {
		if src != c.opalFireSrc {
			return
		}
		if !c.StatusIsActive(opalFireStateKey) {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           frostFallAbil,
			AttackTag:      attacks.AttackTagElementalArt,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagCitlaliFrostfallStorm,
			ICDGroup:       attacks.ICDGroupCitlaliFrostfallStorm,
			StrikeType:     attacks.StrikeTypeDefault,
			Element:        attributes.Cryo,
			Durability:     25,
			Mult:           frostfall[c.TalentLvlSkill()],
			FlatDmg:        c.a4Dmg(frostFallAbil),
			HitlagFactor:   0.01,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 6), 0, 0, c.c4SkullCB)
		c.itzpapaHitTask(src)
	}, itzpapaInterval)
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.3*60, false)
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Cryo, c.ParticleDelay)
}
