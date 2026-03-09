package sayu

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const burstHitmark = 12
const tickTaskDelay = 20

func init() {
	burstFrames = frames.InitAbilSlice(65) // Q -> N1/E/J
	burstFrames[action.ActionDash] = 64    // Q -> D
	burstFrames[action.ActionSwap] = 64    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// dmg
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Yoohoo Art: Mujina Flurry",
		AttackTag:        attacks.AttackTagElementalBurst,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Anemo,
		Durability:       25,
		Mult:             burst[c.TalentLvlBurst()],
		HitlagFactor:     0.05,
		HitlagHaltFrames: 0.02 * 60,
	}
	snap := c.Snapshot(&ai)
	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.5}, 10)
	c.Core.QueueAttackWithSnap(
		ai,
		snap,
		combat.NewCircleHitOnTarget(burstArea.Shape.Pos(), nil, 4.5),
		burstHitmark,
	)

	// 回復
	atk := snap.Stats.TotalATK()
	heal := initHealFlat[c.TalentLvlBurst()] + atk*initHealPP[c.TalentLvlBurst()]
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  -1,
		Message: "Yoohoo Art: Mujina Flurry",
		Src:     heal,
		Bonus:   snap.Stats[attributes.Heal],
	})

	// Tick処理
	d := c.createBurstSnapshot()
	atk = d.Snapshot.Stats.TotalATK()
	heal = burstHealFlat[c.TalentLvlBurst()] + atk*burstHealPP[c.TalentLvlBurst()]

	if c.Base.Cons >= 6 {
		// TODO: スナップショットか？
		d.Info.FlatDmg += atk * min(d.Snapshot.Stats[attributes.EM]*0.002, 4.0)
		heal += min(d.Snapshot.Stats[attributes.EM]*3, 6000)
	}

	// このタスクが以下の条件で実行されるようにする:
	// - 元素爆発初撃のヒットラグが発生した後
	// - 早柚がさらなるヒットラグの影響を受ける前
	c.QueueCharTask(func() {
		// 最初のTickは145フレーム目
		for i := 145 - tickTaskDelay; i < 145+540-tickTaskDelay; i += 90 {
			c.Core.Tasks.Add(func() {
				// プレイヤーをチェック
				// アクティブキャラクターのHPのみ確認
				char := c.Core.Player.ActiveChar()
				hasC1 := c.Base.Cons >= 1
				// 1凸はHP制限を無視
				needHeal := c.Core.Combat.Player().IsWithinArea(burstArea) && (char.CurrentHPRatio() <= 0.7 || hasC1)

				// 敵をチェック
				enemy := c.Core.Combat.ClosestEnemyWithinArea(burstArea, nil)

				// 攻撃か回復かを判断
				// - 1凸は元素爆発が敵への攻撃とプレイヤーの回復を同時に行えるようになる
				needAttack := enemy != nil && (!needHeal || hasC1)
				if needAttack {
					d.Pattern = combat.NewCircleHitOnTarget(enemy, nil, c.qTickRadius)
					c.Core.QueueAttackEvent(d, 0)
				}
				if needHeal {
					c.Core.Player.Heal(info.HealInfo{
						Caller:  c.Index,
						Target:  char.Index,
						Message: "Muji-Muji Daruma",
						Src:     heal,
						Bonus:   d.Snapshot.Stats[attributes.Heal],
					})
				}
			}, i)
		}
	}, tickTaskDelay)

	c.SetCD(action.ActionBurst, 20*60)
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

// TODO: このヘルパー関数は必要か？
func (c *char) createBurstSnapshot() *combat.AttackEvent {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Muji-Muji Daruma",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burstSkill[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)
	ae := combat.AttackEvent{
		Info:        ai,
		SourceFrame: c.Core.F,
		Snapshot:    snap,
	}
	return &ae
}
