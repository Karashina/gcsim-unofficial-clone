package ayaka

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int
var chargeHitmarks = []int{27, 33, 39}

func init() {
	chargeFrames = frames.InitAbilSlice(71)
	chargeFrames[action.ActionSkill] = 62
	chargeFrames[action.ActionBurst] = 63
	chargeFrames[action.ActionDash] = chargeHitmarks[len(chargeHitmarks)-1]
	chargeFrames[action.ActionJump] = chargeHitmarks[len(chargeHitmarks)-1]
	chargeFrames[action.ActionSwap] = chargeHitmarks[len(chargeHitmarks)-1]
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		Abil:       "Charge",
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupAyakaExtraAttack,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       ca[c.TalentLvlAttack()],
	}

	// 最大5回の攻撃を生成
	// 優先度: 敵 > ガジェット
	chargeCount := 5
	checkDelay := chargeHitmarks[0] - 1 // TODO: exact delay unknown
	singleCharge := func(pos geometry.Point, hitmark int) {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(
				pos,
				nil,
				1,
			),
			hitmark,
			hitmark,
			c.c1,
			c.c6,
		)
	}

	charge := func(target combat.Target) {
		for j := 0; j < 3; j++ {
			// ターゲットが移動する可能性があるため重撃ヒットをキューに追加
			c.Core.Tasks.Add(func() {
				singleCharge(target.Pos(), 0)
			}, chargeHitmarks[j]-checkDelay)
		}
	}

	c.Core.Tasks.Add(func() {
		// プレイヤー周囲の敵を探す
		enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), nil)

		// 範囲内に敵がいない場合は何もしない
		if enemies == nil {
			return
		}

		// 見つかった敵の周囲の敵を確認
		anchorEnemy := enemies[0]
		chargeArea := combat.NewCircleHitOnTarget(anchorEnemy, nil, 4)
		enemies = c.Core.Combat.EnemiesWithinArea(chargeArea, func(t combat.Enemy) bool {
			return t.Key() != anchorEnemy.Key() // 同じ敵を2回ターゲットしない
		})
		enemyCount := len(enemies)

		// 敵に攻撃を生成
		charge(anchorEnemy)
		chargeCount -= 1
		for i := 0; i < chargeCount; i++ {
			if i < enemyCount {
				charge(enemies[i])
			}
		}
		chargeCount -= enemyCount

		// 後続の重撃ヒットをキューに追加

		// ターゲットした敵が5体未満の場合、ガジェットを確認
		if chargeCount > 0 {
			gadgets := c.Core.Combat.GadgetsWithinArea(chargeArea, nil)
			gadgetCount := len(gadgets)
			for i := 0; i < chargeCount; i++ {
				if i < gadgetCount {
					charge(gadgets[i])
				}
			}
		}
	}, checkDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[len(chargeHitmarks)-1],
		State:           action.ChargeAttackState,
	}, nil
}
