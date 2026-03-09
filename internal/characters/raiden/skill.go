package raiden

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

/**
雷電将軍が涄善の一片を展開し、周囲の敵に雷元素ダメージを与え、近くのパーティメンバーに雷牠の眼を付与する。
雷牠の眼
**/

var skillFrames []int

const (
	skillHitmark   = 51
	skillKey       = "raiden-e"
	particleICDKey = "raiden-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(37)
	skillFrames[action.ActionDash] = 17
	skillFrames[action.ActionJump] = 17
	skillFrames[action.ActionSwap] = 36
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Eye of Stormy Judgement",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5),
		skillHitmark,
		skillHitmark,
	)

	// ダメージ前修飾子を追加
	mult := skillBurstBonus[c.TalentLvlSkill()]
	m := make([]float64, attributes.EndStatType)
	for _, char := range c.Core.Player.Chars() {
		this := char
		// CDディレイの1秒後に開始
		c.Core.Tasks.Add(func() {
			this.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag(skillKey, 1500),
				Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
					if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
						return nil, false
					}

					m[attributes.DmgP] = mult * this.EnergyMax
					return m, true
				},
			})
		}, 6+60)
	}

	c.SetCDWithDelay(action.ActionSkill, 600, 6)

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
	c.AddStatus(particleICDKey, 0.8*60, true)
	if c.Core.Rand.Float64() < 0.5 {
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Electro, c.ParticleDelay)
	}
}

/*
*
このバフを持つキャラクターが攻撃して敵に命中すると、雷牠の眼が連携攻撃を発動し、敵の位置に雷元素範囲ダメージを与える。
雷牠の眼はパーティごとに0.9秒に1回連携攻撃を行える。
*
*/
func (c *char) eyeOnDamage() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		trg := args[0].(combat.Target)
		ae := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		// 眼のICD中なら無視
		if c.eyeICD > c.Core.F {
			return false
		}
		// ダメージを与えたキャラに眼のステータスがない場合は無視
		if !c.Core.Player.ByIndex(ae.Info.ActorIndex).StatusIsActive(skillKey) {
			return false
		}
		// 感電・水拡散・燃焧ダメージを無視
		// これらのダメージタイプはキャラではなくターゲットがソースのため
		if ae.Info.AttackTag == attacks.AttackTagECDamage || ae.Info.AttackTag == attacks.AttackTagBurningDamage ||
			ae.Info.AttackTag == attacks.AttackTagSwirlHydro {
			return false
		}
		// 自身のダメージを無視
		if ae.Info.ActorIndex == c.Index &&
			ae.Info.AttackTag == attacks.AttackTagElementalArt &&
			ae.Info.StrikeType == attacks.StrikeTypeSlash {
			return false
		}
		// 0ダメージを無視
		if dmg == 0 {
			return false
		}

		// ヒットマーク857、眼着弾862
		// 雷元素は即座に付与される模様
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Eye of Stormy Judgement (Strike)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skillTick[c.TalentLvlSkill()],
		}
		if c.Base.Cons >= 2 && c.StatusIsActive(BurstKey) {
			ai.IgnoreDefPercent = 0.6
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 4), 5, 5, c.particleCB)

		c.eyeICD = c.Core.F + 54 // 0.9 sec icd
		return false
	}, "raiden-eye")
}
