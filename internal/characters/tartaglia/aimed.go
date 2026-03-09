package tartaglia

import (
	"errors"
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var aimedFrames [][]int

var aimedHitmarks = []int{15, 86}

func init() {
	aimedFrames = make([][]int, 2)

	// 狙い撃ち
	aimedFrames[0] = frames.InitAbilSlice(23)
	aimedFrames[0][action.ActionDash] = aimedHitmarks[0]
	aimedFrames[0][action.ActionJump] = aimedHitmarks[0]

	// フルチャージ狙い撃ち
	aimedFrames[1] = frames.InitAbilSlice(94)
	aimedFrames[1][action.ActionDash] = aimedHitmarks[1]
	aimedFrames[1][action.ActionJump] = aimedHitmarks[1]
}

// フルチャージ時、水元素ダメージを与えて断流を付与する。
func (c *char) Aimed(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(meleeKey) {
		return action.Info{}, errors.New("aim called when not in ranged stance")
	}
	hold, ok := p["hold"]
	if !ok {
		hold = attacks.AimParamLv1
	}
	switch hold {
	case attacks.AimParamPhys:
	case attacks.AimParamLv1:
	default:
		return action.Info{}, fmt.Errorf("invalid hold param supplied, got %v", hold)
	}
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}
	weakspot := p["weakspot"]

	ai := combat.AttackInfo{
		ActorIndex:           c.Index,
		Abil:                 "Fully-Charged Aimed Shot",
		AttackTag:            attacks.AttackTagExtra,
		ICDTag:               attacks.ICDTagNone,
		ICDGroup:             attacks.ICDGroupDefault,
		StrikeType:           attacks.StrikeTypePierce,
		Element:              attributes.Hydro,
		Durability:           25,
		Mult:                 fullaim[c.TalentLvlAttack()],
		HitWeakPoint:         weakspot == 1,
		HitlagHaltFrames:     0.12 * 60, // 設置型ヒットラグ（弱点のみ）
		HitlagFactor:         0.01,
		HitlagOnHeadshotOnly: true,
		IsDeployable:         true,
	}
	if hold < attacks.AimParamLv1 {
		ai.Abil = "Aimed Shot"
		ai.Element = attributes.Physical
		ai.Mult = aim[c.TalentLvlAttack()]
	}
	// 元素スキルが命中前に発動された場合:
	// - 物理狙い撃ちは断流を付与し斉しをトリガーする
	// - フルチャージ狙い撃ちは断流を付与し斉しをトリガーする
	// そうでない場合:
	// - 物理狙い撃ちは何もしない
	// - フルチャージ狙い撃ちは断流を付与する
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			geometry.Point{Y: -0.5},
			0.1,
			1,
		),
		aimedHitmarks[hold],
		aimedHitmarks[hold]+travel,
		// 断流付与前にフラッシュを発動する必要がある（フルチャージ狙い撃ちのみではトリガーされないため）
		c.rtFlashCallback,
		c.aimedApplyRiptide,
		c.rtSlashCallback,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(aimedFrames[hold]),
		AnimationLength: aimedFrames[hold][action.InvalidAction],
		CanQueueAfter:   aimedHitmarks[hold],
		State:           action.AimState,
	}, nil
}
