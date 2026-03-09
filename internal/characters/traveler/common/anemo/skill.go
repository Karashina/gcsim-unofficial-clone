package anemo

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	skillPressFrames     [][]int
	skillHoldDelayFrames [][]int
)

const (
	pressParticleICDKey = "traveleranemo-press-particle-icd"
	holdParticleICDKey  = "traveleranemo-hold-particle-icd"
)

func init() {
	// 元素スキル（単押し）
	skillPressFrames = make([][]int, 2)

	// 男性
	skillPressFrames[0] = frames.InitAbilSlice(74) // Tap E -> N1
	skillPressFrames[0][action.ActionBurst] = 76   // Tap E -> Q
	skillPressFrames[0][action.ActionDash] = 30    // Tap E -> D
	skillPressFrames[0][action.ActionJump] = 31    // Tap E -> J
	skillPressFrames[0][action.ActionSwap] = 66    // Tap E -> Swap

	// 女性
	skillPressFrames[1] = frames.InitAbilSlice(62) // Tap E -> Q
	skillPressFrames[1][action.ActionAttack] = 61  // Tap E -> N1
	skillPressFrames[1][action.ActionDash] = 31    // Tap E -> D
	skillPressFrames[1][action.ActionJump] = 31    // Tap E -> J
	skillPressFrames[1][action.ActionSwap] = 60    // Tap E -> Swap

	// 長押し元素スキルフレームのベースとして短押しホールドEを使用
	// 「2ティック持続 - 2ティック最終ヒットマーク」
	skillHoldDelayFrames = make([][]int, 2)

	// 男性
	skillHoldDelayFrames[0] = frames.InitAbilSlice(98 - 54) // Short Hold E -> N1/Q - Short Hold E -> D
	skillHoldDelayFrames[0][action.ActionDash] = 0          // Short Hold E -> D - Short Hold E -> D
	skillHoldDelayFrames[0][action.ActionJump] = 0          // Short Hold E -> J - Short Hold E -> D
	skillHoldDelayFrames[0][action.ActionSwap] = 89 - 54    // Short Hold E -> Swap - Short Hold E -> D

	// 女性
	skillHoldDelayFrames[1] = frames.InitAbilSlice(84 - 54) // Short Hold E -> Q - Short Hold E -> D
	skillHoldDelayFrames[1][action.ActionAttack] = 83 - 54  // Short Hold E -> N1 - Short Hold E -> D
	skillHoldDelayFrames[1][action.ActionDash] = 0          // Short Hold E -> D - Short Hold E -> D
	skillHoldDelayFrames[1][action.ActionJump] = 0          // Short Hold E -> J - Short Hold E -> D
	skillHoldDelayFrames[1][action.ActionSwap] = 83 - 54    // Short Hold E -> Swap - Short Hold E -> D
}

func (c *Traveler) SkillPress() action.Info {
	hitmark := 34
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Palm Vortex (Tap)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skillInitialStorm[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 6, 100),
		hitmark,
		hitmark,
		c.pressParticleCB,
	)

	c.SetCDWithDelay(action.ActionSkill, 5*60, hitmark-5)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames[c.gender]),
		AnimationLength: skillPressFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   skillPressFrames[c.gender][action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *Traveler) pressParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(pressParticleICDKey) {
		return
	}
	c.AddStatus(pressParticleICDKey, 0.6*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Anemo, c.ParticleDelay)
}

func (c *Traveler) SkillHold(holdTicks int) action.Info {
	c.eAbsorb = attributes.NoElement
	c.eICDTag = attacks.ICDTagNone
	c.eAbsorbCheckLocation = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.2}, 3)

	aiCut := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Palm Vortex Initial Cutting (Hold)",
		AttackTag:  attacks.AttackTagElementalArtHold,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skillInitialCutting[c.TalentLvlSkill()],
	}

	aiCutAbs := aiCut
	aiCutAbs.Abil = "Palm Vortex Initial Cutting Absorbed (Hold)"
	aiCutAbs.ICDTag = attacks.ICDTagNone
	aiCutAbs.StrikeType = attacks.StrikeTypeDefault
	aiCutAbs.Element = attributes.NoElement
	aiCutAbs.Mult = skillInitialCuttingAbsorb[c.TalentLvlSkill()]

	aiMaxCutAbs := aiCutAbs
	aiMaxCutAbs.Abil = "Palm Vortex Max Cutting Absorbed (Hold)"
	aiMaxCutAbs.Mult = skillMaxCuttingAbsorb[c.TalentLvlSkill()]

	// 最初のティックは31f、以降15f間隔、初期から最大に移行する際に追加5fの遅延
	firstTick := 31
	hitmark := firstTick
	for i := 0; i < holdTicks; i += 1 {
		c.Core.QueueAttack(
			aiCut,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.2}, 1.7),
			hitmark,
			hitmark,
		)
		if i > 1 {
			c.Core.Tasks.Add(func() {
				if c.eAbsorb != attributes.NoElement {
					aiMaxCutAbs.Element = c.eAbsorb
					aiMaxCutAbs.ICDTag = c.eICDTag
					c.Core.QueueAttack(
						aiMaxCutAbs,
						combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.2}, 3.6),
						0,
						0,
					)
				}
				// 元素吸収されたかチェック
			}, hitmark)
		} else {
			c.Core.Tasks.Add(func() {
				if c.eAbsorb != attributes.NoElement {
					aiCutAbs.Element = c.eAbsorb
					aiCutAbs.ICDTag = c.eICDTag
					c.Core.QueueAttack(
						aiCutAbs,
						combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.2}, 1.7),
						0,
						0,
					)
				}
				// 元素吸収されたかチェック
			}, hitmark)
		}

		// 次のティックへ
		hitmark += 15
		if i == 1 {
			aiCut.Mult = skillMaxCutting[c.TalentLvlSkill()]
			aiCut.Abil = "Palm Vortex Max Cutting (Hold)"

			// 初期から最大に切り替わる際に5fの遅延がある
			hitmark += 5
		}
	}
	// Stormダメージのためヒットマークを1ティック(15f)戻し5f進める
	hitmark = hitmark - 15 + 5
	aiStorm := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Palm Vortex Initial Storm (Hold)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skillInitialStorm[c.TalentLvlSkill()],
	}

	aiStormAbs := aiStorm
	aiStormAbs.Abil = "Palm Vortex Initial Storm Absorbed (Hold)"
	aiStormAbs.ICDTag = attacks.ICDTagNone
	aiStormAbs.Element = attributes.NoElement
	aiStormAbs.Mult = skillInitialStormAbsorb[c.TalentLvlSkill()]

	var particleCB combat.AttackCBFunc
	// 2ティック以上で最大Stormを実行
	if holdTicks >= 2 {
		aiStorm.Mult = skillMaxStorm[c.TalentLvlSkill()]
		aiStorm.Abil = "Palm Vortex Max Storm (Hold)"

		aiStormAbs.Mult = skillMaxStormAbsorb[c.TalentLvlSkill()]
		aiStormAbs.Abil = "Palm Vortex Max Storm Absorbed (Hold)"
		particleCB = c.holdParticleCB
		c.SetCDWithDelay(action.ActionSkill, 8*60, hitmark-5)
	} else {
		particleCB = c.pressParticleCB
		c.SetCDWithDelay(action.ActionSkill, 5*60, hitmark-5)
	}

	c.Core.QueueAttack(
		aiStorm,
		combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 6, 100),
		hitmark,
		hitmark,
		particleCB,
	)
	c.Core.Tasks.Add(func() {
		if c.eAbsorb != attributes.NoElement {
			aiStormAbs.Element = c.eAbsorb
			aiStormAbs.ICDTag = c.eICDTag
			c.Core.QueueAttack(
				aiStormAbs,
				combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 6, 100),
				0,
				0,
			)
		}
		// 元素吸収されたかチェック
	}, hitmark)

	// 最初のティック後に元素吸収を開始？
	c.Core.Tasks.Add(c.absorbCheckE(c.Core.F, 0, hitmark/18), firstTick+1)
	return action.Info{
		Frames:          func(next action.Action) int { return skillHoldDelayFrames[c.gender][next] + hitmark },
		AnimationLength: skillHoldDelayFrames[c.gender][action.InvalidAction] + hitmark,
		CanQueueAfter:   skillHoldDelayFrames[c.gender][action.ActionDash] + hitmark, // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *Traveler) holdParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(holdParticleICDKey) {
		return
	}
	c.AddStatus(holdParticleICDKey, 0.6*60, true)
	count := 3.0
	if c.Core.Rand.Float64() < 0.33 {
		count = 4
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Anemo, c.ParticleDelay)
}

func (c *Traveler) Skill(p map[string]int) (action.Info, error) {
	holdTicks := 0
	if p["hold"] == 1 {
		holdTicks = 6
	}
	if 0 < p["hold_ticks"] {
		holdTicks = p["hold_ticks"]
	}
	if holdTicks > 6 {
		holdTicks = 6
	}

	if holdTicks == 0 {
		return c.SkillPress(), nil
	}
	return c.SkillHold(holdTicks), nil
}

func (c *Traveler) absorbCheckE(src, count, maxcount int) func() {
	return func() {
		if count == maxcount {
			return
		}
		c.eAbsorb = c.Core.Combat.AbsorbCheck(c.Index, c.eAbsorbCheckLocation, attributes.Cryo, attributes.Pyro, attributes.Hydro, attributes.Electro)
		switch c.eAbsorb {
		case attributes.Cryo:
			c.eICDTag = attacks.ICDTagElementalArtCryo
		case attributes.Pyro:
			c.eICDTag = attacks.ICDTagElementalArtPyro
		case attributes.Electro:
			c.eICDTag = attacks.ICDTagElementalArtElectro
		case attributes.Hydro:
			c.eICDTag = attacks.ICDTagElementalArtHydro
		case attributes.NoElement:
			// それ以外はキューに追加
			c.Core.Tasks.Add(c.absorbCheckE(src, count+1, maxcount), 18)
		}
	}
}
