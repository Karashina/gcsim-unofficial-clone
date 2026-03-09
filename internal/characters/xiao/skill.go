package xiao

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const skillHitmark = 4

func init() {
	skillFrames = frames.InitAbilSlice(37)
	skillFrames[action.ActionAttack] = 24
	skillFrames[action.ActionSkill] = 24
	skillFrames[action.ActionBurst] = 24
	skillFrames[action.ActionDash] = 35
	skillFrames[action.ActionSwap] = 35
}

const a4BuffKey = "xiao-a4"

// 元素スキルのダメージキュー生成
// 追加で固有天賦4を実装
// 風輪両立を使用すると、次の風輪両立のダメージが15%増加する。この効果は7秒間持続し、最大3スタック。新しいスタック取得で持続時間がリフレッシュ。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lemniscatic Wind Cycling",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupXiaoDash,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	// 元素爆発中は元素エネルギーを生成できない
	var particleCB combat.AttackCBFunc
	if !c.StatusIsActive(burstBuffKey) {
		particleCB = c.makeParticleCB()
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			0.8,
		),
		0,
		skillHitmark,
		particleCB,
	)

	if c.Base.Ascension >= 4 {
		// 発動から0.25秒後に固有天賦4を適用
		c.Core.Tasks.Add(c.a4, 15)
	}

	// 6凸処理 - CDを無視でき、チャージを消費せずにスキル使用可能
	// 単に早期リターンでOK
	if c.Base.Cons >= 6 && c.StatusIsActive(c6BuffKey) {
		c.Core.Log.NewEvent("xiao c6 active, Xiao E used, no charge used, no CD", glog.LogCharacterEvent, c.Index).
			Write("c6 remaining duration", c.Core.Status.Duration("xiaoc6"))
	} else {
		c.SetCD(action.ActionSkill, 600)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSkill], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) makeParticleCB() combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true
		c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Anemo, c.ParticleDelay)
	}
}
