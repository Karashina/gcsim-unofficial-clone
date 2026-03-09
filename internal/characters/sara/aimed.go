package sara

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var aimedFrames [][]int
var aimedA1Frames []int

var aimedHitmarks = []int{15, 86}

const aimedA1Hitmark = 50

func init() {
	// スキルステータス外
	aimedFrames = make([][]int, 2)

	// 狙い撃ち
	aimedFrames[0] = frames.InitAbilSlice(25)
	aimedFrames[0][action.ActionDash] = aimedHitmarks[0]
	aimedFrames[0][action.ActionJump] = aimedHitmarks[0]

	// フルチャージ狙い撃ち
	aimedFrames[1] = frames.InitAbilSlice(96)
	aimedFrames[1][action.ActionDash] = aimedHitmarks[1]
	aimedFrames[1][action.ActionJump] = aimedHitmarks[1]

	// Fully-Charged Aimed Shot (Crowfeather)
	aimedA1Frames = frames.InitAbilSlice(60)
	aimedA1Frames[action.ActionDash] = aimedA1Hitmark
	aimedA1Frames[action.ActionJump] = aimedA1Hitmark
}

// 狙い撃ち重撃のダメージキュー生成
// クロウフェザー状態、元素スキルダメージ、固有天賦4も処理する
// "travel"パラメータあり。矢が空中にいるフレーム数を設定（デフォルト = 10）
// weak_pointは矢が弱点に命中するかどうか（デフォルト = 1）
func (c *char) Aimed(p map[string]int) (action.Info, error) {
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

	// 固有天賦1:
	// 天狗召嗚のクロウフェザー保護状態中、狙い撃ちのチャージ時間が60%短縮される。
	skillActive := c.Base.Ascension >= 1 && c.Core.Status.Duration(coverKey) > 0

	ai := combat.AttackInfo{
		ActorIndex:           c.Index,
		Abil:                 "Fully-Charged Aimed Shot",
		AttackTag:            attacks.AttackTagExtra,
		ICDTag:               attacks.ICDTagNone,
		ICDGroup:             attacks.ICDGroupDefault,
		StrikeType:           attacks.StrikeTypePierce,
		Element:              attributes.Electro,
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

	var a action.Info

	if skillActive && hold == attacks.AimParamLv1 {
		ai.Abil += " (A1)"
		a = action.Info{
			Frames:          frames.NewAbilFunc(aimedA1Frames),
			AnimationLength: aimedA1Frames[action.InvalidAction],
			CanQueueAfter:   aimedA1Hitmark,
			State:           action.AimState,
		}
	} else {
		a = action.Info{
			Frames:          frames.NewAbilFunc(aimedFrames[hold]),
			AnimationLength: aimedFrames[hold][action.InvalidAction],
			CanQueueAfter:   aimedHitmarks[hold],
			State:           action.AimState,
		}
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			geometry.Point{Y: -0.5},
			0.1,
			1,
		),
		a.CanQueueAfter,
		a.CanQueueAfter+travel,
	)

	// 保護状態の処理 - クロウフェザーを落とし、1.5秒後に爆発する
	if skillActive && hold == attacks.AimParamLv1 {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Tengu Juurai: Ambush",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skill[c.TalentLvlSkill()],
		}
		ap := combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6)

		// TODO: スナップショットのタイミング？
		// 粒子は待ち伏せ攻撃が命中した後に生成される
		c.Core.QueueAttack(ai, ap, aimedA1Hitmark, aimedA1Hitmark+travel+90, c.makeA4CB(), c.particleCB)
		c.attackBuff(ap, aimedA1Hitmark+travel+90)

		c.Core.Status.Delete(coverKey)
	}

	return a, nil
}
