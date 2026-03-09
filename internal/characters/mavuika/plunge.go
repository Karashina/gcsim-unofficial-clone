package mavuika

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var highPlungeFrames []int
var lowPlungeFrames []int
var lowBikePlungeFrames []int
var highBikePlungeFrames []int

const lowPlungeHitmark = 37
const highPlungeHitmark = 41
const lowBikePungeHitmark = 41
const highBikePungeHitmark = 45
const collisionHitmark = lowPlungeHitmark - 6

const lowPlungePoiseDMG = 150.0
const lowPlungeRadius = 3.0

const highPlungePoiseDMG = 200.0
const highPlungeRadius = 5.0

const bikePlungePoiseDMG = 150.0
const bikePlungeRadius = 5.0

func init() {
	// low_plunge -> x
	lowPlungeFrames = frames.InitAbilSlice(80) // low_plunge -> Jump
	lowPlungeFrames[action.ActionAttack] = 51
	lowPlungeFrames[action.ActionCharge] = 52
	lowPlungeFrames[action.ActionSkill] = 37 // low_plunge -> skill[recast=1]
	lowPlungeFrames[action.ActionBurst] = 51
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionWalk] = 79
	lowPlungeFrames[action.ActionSwap] = 62

	// high_plunge -> x
	highPlungeFrames = frames.InitAbilSlice(83)
	highPlungeFrames[action.ActionAttack] = 54
	highPlungeFrames[action.ActionCharge] = 55
	highPlungeFrames[action.ActionSkill] = 40 // low_plunge -> skill[recast=1]
	highPlungeFrames[action.ActionBurst] = 53
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	highPlungeFrames[action.ActionWalk] = 82
	highPlungeFrames[action.ActionSwap] = 65

	// Flamestrider low_plunge -> X
	lowBikePlungeFrames = frames.InitAbilSlice(77) // low_plunge -> Walk
	lowBikePlungeFrames[action.ActionAttack] = 60
	lowBikePlungeFrames[action.ActionCharge] = 60
	lowBikePlungeFrames[action.ActionSkill] = 41 // low_plunge -> skill[recast=1]
	lowBikePlungeFrames[action.ActionBurst] = 61
	lowBikePlungeFrames[action.ActionDash] = lowBikePungeHitmark
	lowBikePlungeFrames[action.ActionJump] = 76
	lowBikePlungeFrames[action.ActionSwap] = 75

	// Flamestrider high_plunge -> X
	highBikePlungeFrames = frames.InitAbilSlice(80) // low_plunge -> Walk
	highBikePlungeFrames[action.ActionAttack] = 63
	highBikePlungeFrames[action.ActionCharge] = 63
	highBikePlungeFrames[action.ActionSkill] = 44 // low_plunge -> skill[recast=1]
	highBikePlungeFrames[action.ActionBurst] = 64
	highBikePlungeFrames[action.ActionDash] = highBikePungeHitmark
	highBikePlungeFrames[action.ActionJump] = 79
	highBikePlungeFrames[action.ActionSwap] = 78
}

// 低空落下攻撃のダメージキュー生成
// 落下中の攻撃判定を行いたい場合は "collision" オプション引数を使用
// デフォルト = 0
func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	if c.Core.Player.Airborne() == player.AirborneXianyun || c.canBikePlunge {
		if c.nightsoulState.HasBlessing() {
			return c.bikePlungeAttack(lowBikePlungeFrames, lowPlungeHitmark), nil
		}
		return c.lowPlungeXY(p), nil
	}
	return action.Info{}, errors.New("low_plunge can only be used while airborne")
}

// バイクジャンプ（歩行状態）→ ナイトソウル消失 → 落下攻撃 にも使用
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
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		if c.nightsoulState.HasBlessing() && c.armamentState == bike {
			return c.bikePlungeAttack(highBikePlungeFrames, highPlungeHitmark), nil
		}
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

// 炎駆巡者落下攻撃のダメージキュー生成
func (c *char) bikePlungeAttack(bikePlungeFrames []int, delay int) action.Info {
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Flamestrider Plunge",
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagMavuikaFlamestrider,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       bikePlungePoiseDMG,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           skillPlunge[c.TalentLvlSkill()],
		HitlagFactor:   0.1,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, bikePlungeRadius),
		delay,
		delay,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(bikePlungeFrames),
		AnimationLength: bikePlungeFrames[action.InvalidAction],
		CanQueueAfter:   bikePlungeFrames[action.ActionDash],
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
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 1), delay, delay)
}
