package itto

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	skillRelease   = 14
	particleICDKey = "itto-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(42) // E -> N1/Q
	skillFrames[action.ActionCharge] = 28  // 丑二郎は常に命中してスタック獲得と仮定、E -> CA1/CAF を使用
	skillFrames[action.ActionSkill] = 28   // E -> E
	skillFrames[action.ActionDash] = 28    // E -> D
	skillFrames[action.ActionJump] = 28    // E -> J
	skillFrames[action.ActionSwap] = 41    // E -> Swap
}

// 元素スキル:
// 荒瀧派の若き赤牛・丑二郎を投げつけ、命中した敵に岩元素ダメージを与える。
// 丑二郎が敵に命中すると、荒瀧一斗は怒髪衝天スタックを1つ獲得する。
// 丑二郎はフィールドに留まり、以下の方法でサポートする:
// - 周囲の敵を挑発し、敵の攻撃を引きつける。
// - 荒瀧一斗のHP上限の一定割合に基づくHPを継承する。
// - 丑二郎がダメージを受けると、荒瀧一斗が怒髪衝天スタックを1つ獲得する。2秒ごとに1スタックのみ獲得可能。
// - 丑二郎はHPが0になるか持続時間が終了するとフィールドを離れる。離脱時に荒瀧一斗に怒髪衝天スタックを1つ付与する。
// 丑二郎は岩元素構造物とみなされる。荒瀧一斗はフィールド上に丑二郎を1体のみ配置可能。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 通常攻撃/スキル以外 -> スキルはsavedNormalCounterをリセットすべき
	// ダッシュのAnimationLengthがDash -> Skillと同じためIdleに切り替わるので、ここでCurrentStateは使用不可
	switch c.Core.Player.LastAction.Type {
	case action.ActionAttack:
	case action.ActionSkill:
	default:
		c.savedNormalCounter = 0
	}

	// 丑二郎は投げられて着地まで12フレームかかるため、将来のために"travel"パラメータを追加
	travel, ok := p["travel"]
	if !ok {
		travel = 4
	}

	// TODO: 敵の攻撃が実装されたらリファクタリング
	ushihit, ok := p["ushihit"]
	if !ok {
		ushihit = 0
	}
	if ushihit < 0 {
		ushihit = 0
	}
	if ushihit > 3 {
		ushihit = 3
	}

	// 生成時にダメージを与える
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Masatsu Zetsugi: Akaushi Burst!",
		AttackTag:        attacks.AttackTagElementalArt,
		ICDTag:           attacks.ICDTagElementalArt,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		PoiseDMG:         250,
		Element:          attributes.Geo,
		Durability:       25,
		Mult:             skill[c.TalentLvlSkill()],
		HitlagHaltFrames: 0.02 * 60,
		HitlagFactor:     0.01,
		IsDeployable:     true,
	}

	ushiDir := c.Core.Combat.Player().Direction()
	ushiPos := c.Core.Combat.PrimaryTarget().Pos()

	// 攻撃
	// 丑二郎の構造物生成用コールバック
	done := false
	cb := func(a combat.AttackCB) {
		if done {
			return
		}
		done = true

		// 丑二郎を生成。フィールド上に6秒間存在
		c.Core.Constructs.New(c.newUshi(6*60, ushiDir, ushiPos), true)

		// パラメータでスタックを追加
		// 2秒ICDでランダムにスタックを獲得
		if ushihit > 0 {
			startLimit := 6 - 2*(ushihit-1)
			nextPossibleGain := 0
			for i := 0; i < ushihit; i++ {
				gain := c.Core.Rand.Intn((startLimit+2*i)*60-nextPossibleGain) + nextPossibleGain
				c.Core.Tasks.Add(func() { c.addStrStack("ushi-hit", 1) }, gain)
				nextPossibleGain = gain + 2*60
			}
		}
	}

	// 丑二郎は常に命中してスタックを獲得すると仮定
	c.Core.Tasks.Add(func() { c.addStrStack("ushi-dmg", 1) }, skillRelease+travel)
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3.5),
		skillRelease,
		skillRelease+travel,
		cb,
		c.particleCB,
	)

	// クールダウン
	c.SetCDWithDelay(action.ActionSkill, 600, skillRelease) // CDはリリース時に開始

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.2*60, true)

	count := 3.0
	if c.Core.Rand.Float64() < 0.50 {
		count = 4
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Geo, c.ParticleDelay)
}
