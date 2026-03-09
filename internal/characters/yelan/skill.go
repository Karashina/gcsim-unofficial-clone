package yelan

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillHitmark        = 35
	particleICDKey      = "yelan-particle-icd"
	skillTargetCountTag = "marked"
	skillHoldDuration   = "hold_length" // 未実装
	skillMarkedTag      = "yelan-skill-marked"
)

func init() {
	skillFrames = frames.InitAbilSlice(42)
	skillFrames[action.ActionBurst] = 41
	skillFrames[action.ActionDash] = 41
	skillFrames[action.ActionJump] = 41
	skillFrames[action.ActionSwap] = 40
}

/*
*
命の綱を射出して素早く自身を引き寄せ、経路上の敵を絡め取りマーキングする。
高速移動が終了すると、命の綱が爆発し、マーキングされた敵に夜蘭のHP上限に基づく水元素ダメージを与える。
*
*/
func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lingering Lifeline",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       0,
		FlatDmg:    skill[c.TalentLvlSkill()] * c.MaxHP(),
	}

	// 既存のタグを全て消去
	for _, t := range c.Core.Combat.Enemies() {
		if e, ok := t.(*enemy.Enemy); ok {
			e.SetTag(skillMarkedTag, 0)
		}
	}

	if !c.StatusIsActive("yelanc4") {
		c.c4count = 0
		c.Core.Log.NewEvent("c4 stacks set to 0", glog.LogCharacterEvent, c.Index)
	}

	// ターゲットをループしてマーキングするタスクを追加
	marked, ok := p[skillTargetCountTag]
	// デフォルト1
	if !ok {
		marked = 1
	}
	c.Core.Tasks.Add(func() {
		for _, t := range c.Core.Combat.Enemies() {
			if marked == 0 {
				break
			}
			e, ok := t.(*enemy.Enemy)
			if !ok {
				continue
			}
			e.SetTag(skillMarkedTag, 1)
			c.Core.Log.NewEvent("marked by Lifeline", glog.LogCharacterEvent, c.Index).
				Write("target", e.Key())
			marked--
			c.c4count++
			if c.Base.Cons >= 4 {
				c.AddStatus("yelanc4", 25*60, true)
			}
		}
	}, skillHitmark) //TODO: 長押しスキルのフレーム

	// hold := p["hold"]

	cb := func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		// 破局の矢の状態を確認
		if c.Core.Rand.Float64() < 0.34 {
			c.breakthrough = true
			c.Core.Log.NewEvent("breakthrough state added", glog.LogCharacterEvent, c.Index)
		}
		//TODO: このICDは？
		if c.StatusIsActive(burstKey) {
			c.summonExquisiteThrow()
			c.Core.Log.NewEvent("yelan burst on skill", glog.LogCharacterEvent, c.Index)
		}
	}

	// マーキングされたターゲットにダメージを与えるタスクを追加
	c.Core.Tasks.Add(func() {
		for _, t := range c.Core.Combat.Enemies() {
			e, ok := t.(*enemy.Enemy)
			if !ok {
				continue
			}
			if e.GetTag(skillMarkedTag) == 0 {
				continue
			}
			e.SetTag(skillMarkedTag, 0)
			c.Core.Log.NewEvent("damaging marked target", glog.LogCharacterEvent, c.Index).
				Write("target", e.Key())
			marked--
			// 1フレーム後に攻撃をキュー
			//TODO: 長押しで攻撃範囲は変わる？変わらないと思うが？
			c.Core.QueueAttack(ai, combat.NewSingleTargetHit(e.Key()), 1, 1, c.particleCB, cb)
		}

		// 該当する場合4凸を発動
		//TODO: これが正確か確認
		if c.Base.Cons >= 4 && c.c4count > 0 {
			m := make([]float64, attributes.EndStatType)
			m[attributes.HPP] = float64(c.c4count) * 0.1
			if m[attributes.HPP] > 0.4 {
				m[attributes.HPP] = 0.4
			}
			c.Core.Log.NewEvent("c4 activated", glog.LogCharacterEvent, c.Index).
				Write("enemies count", c.c4count)
			for _, char := range c.Core.Player.Chars() {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("yelan-c4", 25*60),
					AffectedStat: attributes.HPP,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}
		}
	}, skillHitmark) //TODO: スキルダメージのフレーム？付着後5秒？

	c.SetCDWithDelay(action.ActionSkill, 600, skillHitmark-2)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // 最速キャンセル
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
	c.AddStatus(particleICDKey, 0.3*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Hydro, c.ParticleDelay) // TODO: 以前は82だった？
}
