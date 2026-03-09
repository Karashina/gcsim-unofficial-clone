package kuki

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	skillHitmark     = 11 // 初撃
	hpDrainThreshold = 0.2
	ringKey          = "kuki-e"
	particleICDKey   = "kuki-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(52) // E -> Q
	skillFrames[action.ActionAttack] = 50  // E -> N1
	skillFrames[action.ActionDash] = 12    // E -> D
	skillFrames[action.ActionJump] = 11    // E -> J
	skillFrames[action.ActionSwap] = 50    // E -> Swap
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// HPが20%以上の場合のみHPを消費
	if c.CurrentHPRatio() > hpDrainThreshold {
		currentHP := c.CurrentHP()
		maxHP := c.MaxHP()
		hpdrain := 0.3 * currentHP
		// スキルのHP消費はHPを20%以下にすることはできない。
		if (currentHP-hpdrain)/maxHP <= hpDrainThreshold {
			hpdrain = currentHP - hpDrainThreshold*maxHP
		}
		c.Core.Player.Drain(info.DrainInfo{
			ActorIndex: c.Index,
			Abil:       "Sanctifying Ring",
			Amount:     hpdrain,
		})
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Sanctifying Ring",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   30,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
		FlatDmg:    c.Stat(attributes.EM) * 0.25,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), skillHitmark, skillHitmark)

	// 2凸: 草の輪の持続時間が3秒延長。
	skilldur := 720
	if c.Base.Cons >= 2 {
		skilldur = 900 // 12+3s
	}

	// ヒットラグ前に実行されるためキャラキュー不要
	// リング持続時間はヒットマーク後に開始
	c.Core.Tasks.Add(func() {
		// スキルの持続時間とTickはヒットラグの影響を受けない
		c.Core.Status.Add(ringKey, skilldur)
		c.ringSrc = c.Core.F
		c.Core.Tasks.Add(c.bellTick(c.Core.F), 90) // 90フレーム（1.5秒）ごとに実行
		c.Core.Log.NewEvent("Bell activated", glog.LogCharacterEvent, c.Index).
			Write("next expected tick", c.Core.F+90).
			Write("expected end", c.Core.F+skilldur)
	}, 23)

	c.SetCDWithDelay(action.ActionSkill, 15*60, 7)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionJump], // 最速キャンセル
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
	c.AddStatus(particleICDKey, 0.2*60, false)
	if c.Core.Rand.Float64() < .45 {
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Electro, c.ParticleDelay)
	}
}

func (c *char) bellTick(src int) func() {
	return func() {
		if src != c.ringSrc {
			return
		}
		c.Core.Log.NewEvent("Bell ticking", glog.LogCharacterEvent, c.Index)

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Grass Ring of Sanctification",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skilldot[c.TalentLvlSkill()],
			FlatDmg:    c.a4Damage(),
		}
		// ダメージを発動
		//TODO: スナップショット要確認
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 2, 2, c.particleCB)

		// 固有天賦4はここで考慮される
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.Active(),
			Message: "Grass Ring of Sanctification Healing",
			Src:     (skillhealpp[c.TalentLvlSkill()]*c.MaxHP() + skillhealflat[c.TalentLvlSkill()] + c.a4Healing()),
			Bonus:   c.Stat(attributes.Heal),
		})

		if c.Core.Status.Duration(ringKey) == 0 {
			return
		}
		c.Core.Tasks.Add(c.bellTick(src), 90)
	}
}
