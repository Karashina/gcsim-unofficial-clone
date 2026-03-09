package baizhu

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
)

var burstFrames []int

const (
	burstFirstShield  = 81
	burstFirstRefresh = 142
	burstRefreshRate  = 146
	burstShieldExpiry = 152
	// 変更の可能性あり
	burstTickRelease = 21
	burstTickTravel  = 8
)

func init() {
	burstFrames = frames.InitAbilSlice(105) // Q -> CA/D
	burstFrames[action.ActionAttack] = 104
	burstFrames[action.ActionSkill] = 104
	burstFrames[action.ActionJump] = 104
	burstFrames[action.ActionWalk] = 104
	burstFrames[action.ActionSwap] = 102
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 最初のシールドでは回復しない
	c.Core.Tasks.Add(func() {
		c.summonSeamlessShield()
	}, burstFirstShield)

	// シールドを5回更新
	for i := 0; i <= 4; i += 1 {
		c.Core.Tasks.Add(func() {
			c.summonSeamlessShield()
			c.summonSeamlessShieldHealing()
		}, burstFirstShield+burstFirstRefresh+burstRefreshRate*i)
	}

	if c.Base.Cons >= 4 {
		c.c4()
	}

	c.SetCD(action.ActionBurst, 20*60)
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) summonSeamlessShield() {
	// シールドを追加
	exist := c.Core.Player.Shields.Get(shield.BaizhuBurst)
	shieldamt := (burstShieldPP[c.TalentLvlBurst()]*c.MaxHP() + burstShieldFlat[c.TalentLvlBurst()])
	if exist != nil {
		c.summonSpiritvein()
	}
	c.Core.Player.Shields.Add(c.newShield(shieldamt, burstShieldExpiry))
}

func (c *char) summonSeamlessShieldHealing() {
	// 継ぎ目なきシールドの回復
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  c.Core.Player.Active(),
		Message: "Seamless Shield Healing",
		Src:     burstHealPP[c.TalentLvlBurst()]*c.MaxHP() + burstHealFlat[c.TalentLvlBurst()],
		Bonus:   c.Stat(attributes.Heal),
	})
	c.a4()
}

func (c *char) summonSpiritvein() {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Spiritvein Damage",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       burstAtk[c.TalentLvlBurst()],
	}
	if c.Base.Cons >= 6 {
		ai.FlatDmg = c.MaxHP() * 0.08
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5),
		burstTickRelease,
		burstTickRelease+burstTickTravel,
	)
}
