package ganyu

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var aimedFrames [][]int

var aimedHitmarks = []int{15, 74, 103}

func init() {
	aimedFrames = make([][]int, 3)

	// 狙い撃ち
	aimedFrames[0] = frames.InitAbilSlice(25)
	aimedFrames[0][action.ActionDash] = aimedHitmarks[0]
	aimedFrames[0][action.ActionJump] = aimedHitmarks[0]

	// フルチャージ狙い撃ちLv.1（フルチャージ狙い撃ち）
	aimedFrames[1] = frames.InitAbilSlice(85)
	aimedFrames[1][action.ActionDash] = aimedHitmarks[1]
	aimedFrames[1][action.ActionJump] = aimedHitmarks[1]

	// フルチャージ狙い撃ちLv.2（霜華の矢 + 霜華満開）
	aimedFrames[2] = frames.InitAbilSlice(113)
	aimedFrames[2][action.ActionDash] = aimedHitmarks[2]
	aimedFrames[2][action.ActionJump] = aimedHitmarks[2]
}

func (c *char) Aimed(p map[string]int) (action.Info, error) {
	hold, ok := p["hold"]
	if !ok {
		hold = attacks.AimParamLv2
	}
	switch hold {
	case attacks.AimParamPhys:
	case attacks.AimParamLv1:
	case attacks.AimParamLv2:
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
		Element:              attributes.Cryo,
		Durability:           25,
		Mult:                 fullaim[c.TalentLvlAttack()],
		HitWeakPoint:         weakspot == 1,
		HitlagHaltFrames:     .12 * 60,
		HitlagOnHeadshotOnly: true,
		IsDeployable:         true,
	}
	if hold < attacks.AimParamLv1 {
		ai.Abil = "Aimed Shot"
		ai.Element = attributes.Physical
		ai.Mult = aim[c.TalentLvlAttack()]
	}
	// TODO: これが意図通りに動作するか不明
	skip := 0
	if c.Core.Status.Duration(c6Key) > 0 && hold == attacks.AimParamLv2 {
		c.Core.Status.Delete(c6Key)
		c.Core.Log.NewEvent(c6Key+" proc used", glog.LogCharacterEvent, c.Index).
			Write("char", c.Index)
		// 狙い撃ちチャージ時間をスキップ
		skip = 83
	}

	// スナップショットの遅延と固有天賦1の処理
	if hold == attacks.AimParamLv2 {
		c.Core.Tasks.Add(func() {
			// Frostflake ArrowとBloomの値が正しいことを確認
			ai.Abil = "Frostflake Arrow"
			ai.Element = attributes.Cryo
			ai.Mult = ffa[c.TalentLvlAttack()]

			snap := c.Snapshot(&ai)
			// 固有天賦1:
			// 霄花矢を発射した後、後続の霄花矢とその眾花の会心率が
			// 5秒間にわたり20%上昇する。
			// - 最初の狙い撃ちには適用されない
			if c.Base.Ascension >= 1 && c.Core.F < c.a1Expiry {
				old := snap.Stats[attributes.CR]
				snap.Stats[attributes.CR] += .20
				c.Core.Log.NewEvent("a1 adding crit rate", glog.LogCharacterEvent, c.Index).
					Write("old", old).
					Write("new", snap.Stats[attributes.CR]).
					Write("expiry", c.a1Expiry)
			}

			c1cb := c.c1()
			c.Core.QueueAttackWithSnap(
				ai,
				snap,
				combat.NewBoxHit(
					c.Core.Combat.Player(),
					c.Core.Combat.PrimaryTarget(),
					geometry.Point{Y: -0.5},
					0.1,
					1,
				),
				travel,
				c1cb,
			)

			ai.Abil = "Frostflake Arrow Bloom"
			ai.Mult = ffb[c.TalentLvlAttack()]
			ai.HitWeakPoint = false
			c.Core.QueueAttackWithSnap(
				ai,
				snap,
				combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
				travel+18, // Bloomは常に矢から18フレーム後に命中
				c1cb,
			)

			// 最初の射撃/ブルームはA1の恩恵を受けない
			c.a1Expiry = c.Core.F + 60*5
		}, aimedHitmarks[hold]-skip)
	} else {
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
		)
	}

	return action.Info{
		Frames:          func(next action.Action) int { return aimedFrames[hold][next] - skip },
		AnimationLength: aimedFrames[hold][action.InvalidAction] - skip,
		CanQueueAfter:   aimedHitmarks[hold] - skip,
		State:           action.AimState,
	}, nil
}
