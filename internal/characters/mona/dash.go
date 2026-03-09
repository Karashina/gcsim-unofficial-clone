package mona

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var dashFrames []int

const dashHitmark = 20

func init() {
	dashFrames = frames.InitAbilSlice(42) // D -> N1
	dashFrames[action.ActionCharge] = 36  // D -> CA
	dashFrames[action.ActionSkill] = 35   // D -> E
	dashFrames[action.ActionBurst] = 21   // D -> Q
	dashFrames[action.ActionDash] = 30    // D -> D
	dashFrames[action.ActionJump] = 500   // D -> J, TODO: このアクションは不正。より適切な処理方法が必要
	dashFrames[action.ActionSwap] = 34    // D -> Swap
}

func (c *char) Dash(p map[string]int) (action.Info, error) {
	f, ok := p["f"]
	if !ok {
		f = 0
	}
	// ダッシュ終了時にダメージなし攻撃
	ai := combat.AttackInfo{
		Abil:       "Dash",
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagDash,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.1}, 2),
		dashHitmark+f,
		dashHitmark+f,
	)

	// 固有天賦1
	if c.Base.Ascension >= 1 {
		c.Core.Tasks.Add(c.a1, 120)
	}
	// 6凸
	if c.Base.Cons >= 6 {
		// 重撃を使用する前に再度ダッシュした場合に備えて6凸スタックをリセット
		c.c6Stacks = 0
		// モナのダッシュダッシュ時に、2回目のダッシュが6凸ティック間に始まる場合のsrc追跡が必要
		// srcチェックがないと２回目のダッシュが1秒経過前にスタックを獲得し、1秒時点でさらにもう1つ獲得する
		c.c6Src = c.Core.F
		c.Core.Tasks.Add(c.c6(c.Core.F), 60)
	}

	// スタミナ消費を処理。CDが不要なためデフォルトのダッシュ実装を使わない
	c.QueueDashStaminaConsumption(p)

	return action.Info{
		Frames:          func(next action.Action) int { return dashFrames[next] + f },
		AnimationLength: dashFrames[action.InvalidAction] + f,
		CanQueueAfter:   dashHitmark + f,
		State:           action.DashState,
	}, nil
}
