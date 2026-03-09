package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 45

func init() {
	// Mona（水元素法器）パターンに基づくスタブ値
	chargeFrames = frames.InitAbilSlice(45) // CA -> D
	chargeFrames[action.ActionAttack] = 83  // CA -> N1
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// 緑の露がMoondew Cleanseに使用可能か確認
	if c.Core.Player.Verdant.Count() >= 1 {
		return c.moondewCleanse()
	}

	// 通常の重撃（水元素ダメージ）
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack (C)",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// 待機または交代時のみ溜め時間を追加
	windup := 14
	if c.Core.Player.CurrentState() == action.Idle || c.Core.Player.CurrentState() == action.SwapState {
		windup = 0
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			3,
		),
		chargeHitmark-windup,
		chargeHitmark-windup,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeFrames[action.ActionSwap] - windup, // 最速キャンセルはヒットマークより前
		State:           action.ChargeAttackState,
	}, nil
}

// moondewCleanseは緑の露が使用可能な時の特殊重撃を実行
// 緑の露1つを消費し、AoE草元素ダメージを3回与える（Lunar-Bloomダメージとして扱われる）
func (c *char) moondewCleanse() (action.Info, error) {
	// まず緑の露を確認、次にMoonridge Dewを確認
	consumed := c.Core.Player.Verdant.Consume(1)
	if consumed == 0 && c.moonridgeDew > 0 {
		c.moonridgeDew--
	}

	// 草元素ダメージ3回をLunar-Bloomダメージとみなす
	for i := 0; i < 3; i++ {
		delay := chargeHitmark + i*8
		c.Core.Tasks.Add(func() {
			c.queueMoondewCleanseHit()
		}, delay)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

// queueMoondewCleanseHitは単一のMoondew Cleanseヒットをキューに入れる（Lunar-Bloomダメージ）
func (c *char) queueMoondewCleanseHit() {
	// Lunar-Bloomダメージとして扱うためAttackTagLBDamageを使用
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Moondew Cleanse (C)",
		AttackTag:        attacks.AttackTagLBDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Dendro,
		Durability:       25,
		IgnoreDefPercent: 1,
	}

	// HP係数 + Lunar-Bloom計算式
	em := c.Stat(attributes.EM)
	baseDmg := c.MaxHP() * moondewCleanse[c.TalentLvlAttack()] * (1 + c.LBBaseReactBonus(ai))
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + emBonus + c.LBReactBonus(ai)) * (1 + c.ElevationBonus(ai))

	snap := combat.Snapshot{
		CharLvl: c.Base.Level,
	}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: 1.5},
		4.0,
	)

	c.Core.QueueAttackWithSnap(ai, snap, ap, 0)

	// ヒットがあればLunar-Bloomイベントを発行
	enemies := c.Core.Combat.EnemiesWithinArea(ap, nil)
	if len(enemies) > 0 {
		ae := &combat.AttackEvent{Info: ai}
		c.Core.Events.Emit(event.OnLunarBloom, enemies[0], ae)
	}
}
