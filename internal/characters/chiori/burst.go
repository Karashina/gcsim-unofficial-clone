package chiori

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var burstFrames []int

const (
	burstHitmark        = 92
	burstSnapshotTiming = burstHitmark - 1 // TODO: スナップショットタイミング?
	burstEnergyFrame    = 10
)

func init() {
	burstFrames = frames.InitAbilSlice(101)
	burstFrames[action.ActionSkill] = 100
	burstFrames[action.ActionDash] = 100
	burstFrames[action.ActionSwap] = 99
}

// 双剣が鞘から離れ、千織が一流の仕立屋のように綺麗な斬撃で斜る。
// 攻撃力と防御力に基づく岩元素範囲ダメージを与える。
func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Hiyoku: Twin Blades",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   200,
		Element:    attributes.Geo,
		Durability: 50,
		Mult:       burstAtkScaling[c.TalentLvlBurst()],
	}

	c.Core.Tasks.Add(func() {
		snap := c.Snapshot(&ai)

		// 防御力スケーリング部分の固定ダメージ
		ai.FlatDmg = snap.Stats.TotalDEF()
		ai.FlatDmg *= burstDefScaling[c.TalentLvlBurst()]

		// 2凸は実際のダメージ発生の少し前に呼ぶ必要がある
		c.c2()

		// TODO: ヒットボックス、chiori mainsが間違っていたら彼らのせい
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 12),
			burstHitmark-burstSnapshotTiming,
		)
	}, burstSnapshotTiming)

	c.ConsumeEnergy(burstEnergyFrame)
	c.SetCD(action.ActionBurst, 13.5*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}
