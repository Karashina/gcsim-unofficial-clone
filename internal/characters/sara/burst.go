package sara

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var burstFrames []int

const burstStart = 47           // CD開始と同時
const burstInitialHitmark = 51  // 初撃
const burstClusterHitmark = 100 // 最初のクラスターヒット

func init() {
	burstFrames = frames.InitAbilSlice(80) // Q -> CA
	burstFrames[action.ActionAttack] = 78  // Q -> N1
	burstFrames[action.ActionSkill] = 57   // Q -> E
	burstFrames[action.ActionDash] = 58    // Q -> D
	burstFrames[action.ActionJump] = 58    // Q -> J
	burstFrames[action.ActionSwap] = 56    // Q -> Swap
}

// 天狗呉雷・螢重を落とし、範囲雷元素ダメージを与える。その後、天狗呉雷・螢重は4回の
// 天狗呉雷・雷砤に分散し、範囲雷元素ダメージを与える。
// 天狗呉雷・螢重と天狗呉雷・雷砤は、元素スキルと同じ攻撃力バフを範囲内のアクティブキャラクターに付与する。
// 各種天狗呉雷の攻撃力バフは重複せず、効果と持続時間は最後に発動した天狗呉雷により決定される。
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 元素爆発全体が発動後、初撃前のどこかでスナップショットされる。
	// 現状、CD遅延時にスナップショットされると仮定
	// ICDなしと設定（雷砤は主撃とICDを共有しないため）
	// ICDなしは、これが1回しかヒットしないため実質的に影響なし

	// 螢重
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Tengu Juurai: Titanbreaker",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burstMain[c.TalentLvlBurst()],
	}

	var c1cb combat.AttackCBFunc
	if c.Base.Cons >= 1 {
		c1cb = func(a combat.AttackCB) {
			if a.Target.Type() != targets.TargettableEnemy {
				return
			}
			c.c1()
		}
	}

	burstInitialDirection := c.Core.Combat.Player().Direction()
	burstInitialPos := c.Core.Combat.PrimaryTarget().Pos()
	initialAp := combat.NewCircleHitOnTarget(burstInitialPos, nil, 6)

	c.Core.QueueAttack(ai, initialAp, burstStart, burstInitialHitmark, c1cb)
	c.attackBuff(initialAp, burstInitialHitmark)

	// stormcluster
	ai.Abil = "Tengu Juurai: Stormcluster"
	ai.ICDTag = attacks.ICDTagElementalBurst
	ai.Mult = burstCluster[c.TalentLvlBurst()]

	stormClusterRadius := 3.0
	var stormClusterCount float64
	if c.Base.Cons >= 4 {
		// 征服・號令による天狗呉雷・雷砤の発生回数が6回に増加する。
		stormClusterCount = 6
	} else {
		stormClusterCount = 4
	}
	stepSize := 360 / stormClusterCount

	for i := 0.0; i < stormClusterCount; i++ {
		// 各雷砤はそれぞれ独自の方向を持つ
		direction := geometry.DegreesToDirection(i * stepSize).Rotate(burstInitialDirection)
		// 1雷砤あたり6Tick
		for j := 0; j < 6; j++ {
			// 3.6mのオフセットから開始、Tickごとに1.35m移動
			stormClusterPos := geometry.CalcOffsetPoint(burstInitialPos, geometry.Point{Y: 3.6 + 1.35*float64(j)}, direction)
			stormClusterAp := combat.NewCircleHitOnTarget(stormClusterPos, nil, stormClusterRadius)

			c.Core.QueueAttack(ai, stormClusterAp, burstStart, burstClusterHitmark+18*j, c1cb)
			c.attackBuff(stormClusterAp, burstClusterHitmark+18*j)
		}
	}

	c.SetCDWithDelay(action.ActionBurst, 20*60, burstStart)
	c.ConsumeEnergy(50)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
