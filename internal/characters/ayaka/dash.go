package ayaka

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var dashFrames []int

const dashHitmark = 20

func init() {
	dashFrames = frames.InitAbilSlice(35)
	dashFrames[action.ActionDash] = 30
	dashFrames[action.ActionSwap] = 34
}

// TODO: 代わりにPostDashイベントに移動する
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
		Element:    attributes.Cryo,
		Durability: 25,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.1}, 2),
		dashHitmark+f,
		dashHitmark+f,
		c.makeA4CB(),
	)

	// 氷元素付与を追加
	//TODO: 武器付与のタイミングを確認; これで問題ないはず？
	c.Core.Tasks.Add(func() {
		c.Core.Player.AddWeaponInfuse(
			c.Index,
			"ayaka-dash",
			attributes.Cryo,
			300,
			true,
			attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
		)
	}, dashHitmark+f)
	c.Core.Events.Emit(event.OnInfusion, c.Index, attributes.Cryo, 300)

	// スタミナ消費を処理。CDが不要なためデフォルトのダッシュ実装を使わない
	c.QueueDashStaminaConsumption(p)

	return action.Info{
		Frames:          func(next action.Action) int { return dashFrames[next] + f },
		AnimationLength: dashFrames[action.InvalidAction] + f,
		CanQueueAfter:   dashFrames[action.ActionDash] + f, // 最速キャンセル
		State:           action.DashState,
	}, nil
}
