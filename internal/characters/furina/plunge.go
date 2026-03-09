package furina

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

const lowPlungeHitmark = 46
const highPlungeHitmark = 47
const collisionHitmark = lowPlungeHitmark - 6

const lowPlungePoiseDMG = 100.0
const lowPlungeRadius = 3.0

const highPlungePoiseDMG = 150.0
const highPlungeRadius = 5.0

func init() {
	// low_plunge -> x
	lowPlungeFrames = frames.InitAbilSlice(75)
	lowPlungeFrames[action.ActionAttack] = 58
	lowPlungeFrames[action.ActionCharge] = 58 - 2
	lowPlungeFrames[action.ActionSkill] = 58
	lowPlungeFrames[action.ActionBurst] = 58
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionWalk] = 73
	lowPlungeFrames[action.ActionSwap] = 60

	// high_plunge -> x
	highPlungeFrames = frames.InitAbilSlice(76)
	highPlungeFrames[action.ActionAttack] = 59
	highPlungeFrames[action.ActionCharge] = 59 - 4
	highPlungeFrames[action.ActionSkill] = 59
	highPlungeFrames[action.ActionBurst] = 59
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	highPlungeFrames[action.ActionWalk] = 74
	highPlungeFrames[action.ActionSwap] = 63
}

// 低空落下攻撃のダメージキュー生成
// 落下中の攻撃判定を行いたい場合は "collision" オプション引数を使用
// デフォルト = 0
func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
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

	// TODO: 6凸が落下攻撃中に期限切れになるか要確認
	var c6cb combat.AttackCBFunc
	if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
		ai.Element = attributes.Hydro
		ai.IgnoreInfusion = true
		ai.FlatDmg = c.c6BonusDMG()
		c6cb = c.c6cb
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, lowPlungeRadius),
		lowPlungeHitmark,
		lowPlungeHitmark,
		c6cb,
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

	// TODO: 6凸が落下攻撃中に期限切れになるか要確認
	var c6cb combat.AttackCBFunc
	if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
		ai.Element = attributes.Hydro
		ai.IgnoreInfusion = true
		ai.FlatDmg = c.c6BonusDMG()
		c6cb = c.c6cb
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, highPlungeRadius),
		highPlungeHitmark,
		highPlungeHitmark,
		c6cb,
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

	// TODO: 6凸が落下攻撃中に期限切れになるか要確認
	if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
		ai.Element = attributes.Hydro
		ai.IgnoreInfusion = true
		ai.FlatDmg = c.c6BonusDMGNoExtra()
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 1), delay, delay)
}
