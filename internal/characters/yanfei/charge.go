package yanfei

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var (
	chargeFrames []int
	chargeRadius = []float64{2.5, 3, 3.5, 4, 4}
)

const chargeHitmark = 63

func init() {
	chargeFrames = frames.InitAbilSlice(79)          // CA -> N1
	chargeFrames[action.ActionCharge] = 78           // CA -> CA
	chargeFrames[action.ActionSkill] = chargeHitmark // CA -> E
	chargeFrames[action.ActionBurst] = chargeHitmark // CA -> Q
	chargeFrames[action.ActionDash] = 51             // CA -> D
	chargeFrames[action.ActionJump] = 49             // CA -> J
	chargeFrames[action.ActionSwap] = 59             // CA -> Swap
}

// 重撃関数 - 朱印消費を処理
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// 朱印スタックを確認
	if !c.StatusIsActive(sealBuffKey) {
		c.sealCount = 0
	}

	// 固有天賦1を適用
	c.a1(c.sealCount)

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   80,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       charge[c.sealCount][c.TalentLvlAttack()],
	}

	// 待機または交代時のみ溜め時間を追加
	windup := 16
	if c.Core.Player.CurrentState() == action.Idle || c.Core.Player.CurrentState() == action.SwapState {
		windup = 0
	}
	radius := chargeRadius[c.sealCount]
	// TODO: スナップショットのタイミングが不明
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, radius),
		chargeHitmark-windup,
		chargeHitmark-windup,
		c.makeA4CB(),
	)

	c.Core.Log.NewEvent("yanfei charge attack consumed seals", glog.LogCharacterEvent, c.Index).
		Write("current_seals", c.sealCount)

	// スタミナチェックが遅延した場合に備え、次フレームで朱印をクリア
	c.Core.Tasks.Add(func() {
		c.sealCount = 0
		c.DeleteStatus(sealBuffKey)
	}, 1)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeFrames[action.ActionJump] - windup, // 最速キャンセルはヒットマークより前
		State:           action.ChargeAttackState,
	}, nil
}
