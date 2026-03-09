package gaming

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var highPlungeFrames []int
var lowPlungeFrames []int
var specialPlungeFrames []int

const lowPlungeHitmark = 43
const highPlungeHitmark = 46
const collisionHitmark = lowPlungeHitmark - 6
const specialPlungeHitmark = 32

const lowPlungePoiseDMG = 150.0
const lowPlungeRadius = 3.0

const highPlungePoiseDMG = 200.0
const highPlungeRadius = 5.0

const hpDrainThreshold = 0.1
const specialPlungeKey = "Charmed Cloudstrider"
const particleICD = 3 * 60
const particleICDKey = "gaming-particle-icd"

func init() {
	// low_plunge -> x
	lowPlungeFrames = frames.InitAbilSlice(84)
	lowPlungeFrames[action.ActionAttack] = 56
	lowPlungeFrames[action.ActionSkill] = 56
	lowPlungeFrames[action.ActionBurst] = 56
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionSwap] = 67

	// high_plunge -> x
	highPlungeFrames = frames.InitAbilSlice(87)
	highPlungeFrames[action.ActionAttack] = 58
	highPlungeFrames[action.ActionSkill] = 57
	highPlungeFrames[action.ActionBurst] = 57
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	highPlungeFrames[action.ActionWalk] = 86
	highPlungeFrames[action.ActionSwap] = 69

	// 特殊落下攻撃
	specialPlungeFrames = frames.InitAbilSlice(99)
	specialPlungeFrames[action.ActionAttack] = 52
	specialPlungeFrames[action.ActionSkill] = 52
	specialPlungeFrames[action.ActionBurst] = 52
	specialPlungeFrames[action.ActionDash] = specialPlungeHitmark // was 30
	specialPlungeFrames[action.ActionWalk] = 74
	specialPlungeFrames[action.ActionSwap] = 69
}

// 低空落下攻撃のダメージキュー生成
// 落下中の攻撃判定を行いたい場合は "collision" オプション引数を使用
// デフォルト = 0
func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	if c.Core.Player.LastAction.Type == action.ActionSkill {
		return c.specialPlunge(p), nil
	}

	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.lowPlungeXY(p), nil
	default:
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
}

func (c *char) lowPlungeXY(p map[string]int) action.Info {
	collision, ok := p["collision"]
	if !ok {
		collision = 0 // 衝突ヒットの有無
	}

	if collision > 0 {
		c.plungeCollision(collisionHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Low Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   lowPlungePoiseDMG,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       lowPlunge[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, lowPlungeRadius),
		lowPlungeHitmark,
		lowPlungeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
	}
}

// 高空落下攻撃のダメージキュー生成
// 落下中の攻撃判定を行いたい場合は "collision" オプション引数を使用
// デフォルト = 0
func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	if c.Core.Player.LastAction.Type == action.ActionSkill {
		return c.specialPlunge(p), nil
	}

	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.highPlungeXY(p), nil
	default:
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
}

func (c *char) highPlungeXY(p map[string]int) action.Info {
	collision, ok := p["collision"]
	if !ok {
		collision = 0 // 衝突ヒットの有無
	}

	if collision > 0 {
		c.plungeCollision(collisionHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "High Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   highPlungePoiseDMG,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       highPlunge[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, highPlungeRadius),
		highPlungeHitmark,
		highPlungeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(highPlungeFrames),
		AnimationLength: highPlungeFrames[action.InvalidAction],
		CanQueueAfter:   highPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
	}
}

// 落下攻撃（通常落下）のダメージキュー生成
// 標準 - 高空/低空落下攻撃に常に含まれる
func (c *char) plungeCollision(delay int) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Plunge Collision",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Physical,
		Durability: 0,
		Mult:       collision[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1), delay, delay)
}

func (c *char) specialPlunge(p map[string]int) action.Info {
	if p[manChaiParam] > 0 {
		c.manChaiWalkBack = p[manChaiParam]
	} else {
		c.manChaiWalkBack = 92
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           specialPlungeKey,
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       lowPlungePoiseDMG,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           skill[c.TalentLvlSkill()],
		IgnoreInfusion: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, c.specialPlungeRadius),
		specialPlungeHitmark,
		specialPlungeHitmark,
		c.particleCB,
		c.makeA1CB(),
		c.makeC4CB(),
	)

	// マンチャイをキューし、ヒットマークの1フレーム後にHPを消費
	c.Core.Tasks.Add(func() {
		if c.StatusIsActive(burstKey) && c.CurrentHPRatio() > 0.5 {
			c.queueManChai()
		}
		// HPが10%超の場合のみHPを消費
		if c.CurrentHPRatio() > hpDrainThreshold {
			currentHP := c.CurrentHP()
			maxHP := c.MaxHP()
			hpdrain := 0.15 * currentHP
			// このスキルにHP消費は10%までしか減らせない。
			if (currentHP-hpdrain)/maxHP <= hpDrainThreshold {
				hpdrain = currentHP - hpDrainThreshold*maxHP
			}
			c.Core.Player.Drain(info.DrainInfo{
				ActorIndex: c.Index,
				Abil:       specialPlungeKey,
				Amount:     hpdrain,
			})
		}
	}, specialPlungeHitmark+1)

	return action.Info{
		Frames:          frames.NewAbilFunc(specialPlungeFrames),
		AnimationLength: specialPlungeFrames[action.InvalidAction],
		CanQueueAfter:   specialPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
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

	c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Pyro, c.ParticleDelay)
}
