package thoma

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var burstFrames []int

const (
	burstKey     = "thoma-q"
	burstICDKey  = "thoma-q-icd"
	burstHitmark = 40
)

func init() {
	burstFrames = frames.InitAbilSlice(58)
	burstFrames[action.ActionAttack] = 57
	burstFrames[action.ActionSkill] = 56
	burstFrames[action.ActionDash] = 57
	burstFrames[action.ActionSwap] = 56
}

// 元素爆発のダメージキュー生成
func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Crimson Ooyoroi",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}

	// ダメージコンポーネントは未確定
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4),
		burstHitmark,
		burstHitmark,
	)

	d := 15
	if c.Base.Cons >= 2 {
		d = 18
	}

	c.AddStatus(burstKey, d*60, true)

	// 4凸: エネルギーを15回復
	if c.Base.Cons >= 4 {
		c.Core.Tasks.Add(func() {
			c.AddEnergy("thoma-c4", 15)
		}, 8)
	}

	cd := 20
	if c.Base.Cons >= 1 {
		cd = 17 // CD短縮はトーマのシールドで保護されたキャラが被弾時に発動。シミュレーション上ほぼ確実に発動するため17に設定
	}
	c.SetCD(action.ActionBurst, cd*60)
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSkill],
		State:           action.BurstState,
	}, nil
}

func (c *char) summonFieryCollapse() {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Fiery Collapse",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       burstproc[c.TalentLvlBurst()],
		FlatDmg:    c.a4(),
	}
	done := false
	shieldCb := func(_ combat.AttackCB) {
		if done {
			return
		}
		shieldamt := (burstshieldpp[c.TalentLvlBurst()]*c.MaxHP() + burstshieldflat[c.TalentLvlBurst()])
		c.genShield("Thoma Burst", shieldamt, true)
		done = true
	}
	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 4.5, 8),
		0,
		11,
		shieldCb,
	)

	c.AddStatus(burstICDKey, 60, true)
}
