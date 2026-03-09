package barbara

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/avatar"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

// バーバラのスキル - ベネットの元素爆発からコピー
const skillDuration = 15*60 + 1
const barbSkillKey = "barbara-e"
const skillCDStart = 3

var (
	skillHitmarks = []int{42, 78}
	skillFrames   []int
)

func init() {
	skillFrames = frames.InitAbilSlice(55)
	skillFrames[action.ActionWalk] = 54
	skillFrames[action.ActionDash] = 4
	skillFrames[action.ActionJump] = 5
	skillFrames[action.ActionSwap] = 53
	skillFrames[action.ActionSkill] = 54
	skillFrames[action.ActionAttack] = 54
	skillFrames[action.ActionCharge] = 54
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// a4カウンターをリセット
	c.a4extendCount = 0

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Let the Show Begin♪ (Droplet)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	// 2つの水滴
	for _, hitmark := range skillHitmarks {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3),
			5,
			hitmark,
		) // スナップショットタイミングの確認が必要
	}

	c.skillInitF = c.Core.F // Tick処理に必要

	// 回復と湿潤Tickの設定（最初のTickはskillCDStart時、以降5秒ごと）
	stats, _ := c.Stats()
	hpplus := stats[attributes.Heal]
	heal := skillhp[c.TalentLvlSkill()] + skillhpp[c.TalentLvlSkill()]*c.MaxHP()

	// メロディーループTickの設定（最初のTickはskillCDStart時、1.5秒ごと）
	ai.Abil = "Let the Show Begin♪ (Melody Loop)"
	ai.AttackTag = attacks.AttackTagNone
	ai.Mult = 0
	ai.HitlagFactor = 0.05
	ai.HitlagHaltFrames = 0.05 * 60
	ai.CanBeDefenseHalted = true
	ai.IsDeployable = true

	// スキルステータスを追加しTickをキューに入れる
	c.Core.Tasks.Add(func() {
		c.Core.Status.Add(barbSkillKey, skillDuration)
		c.a1()
		c.barbaraSelfTick(heal, hpplus, c.skillInitF)()
		c.barbaraMelodyTick(ai, c.skillInitF)()
	}, skillCDStart)

	if c.Base.Cons >= 2 {
		c.c2() // 2凸水元素バフ
		c.SetCDWithDelay(action.ActionSkill, 32*60*0.85, skillCDStart)
	} else {
		c.SetCDWithDelay(action.ActionSkill, 32*60, skillCDStart)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

func (c *char) barbaraSelfTick(healAmt, hpplus float64, skillInitF int) func() {
	return func() {
		// 上書きされていないことを確認
		if c.skillInitF != skillInitF {
			return
		}
		// バフが期限切れなら何もしない
		if c.Core.Status.Duration(barbSkillKey) == 0 {
			return
		}

		c.Core.Log.NewEvent("barbara heal and wet ticking", glog.LogCharacterEvent, c.Index)

		// 回復
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.Active(),
			Message: "Melody Loop (Tick)",
			Src:     healAmt,
			Bonus:   hpplus,
		})

		// 湿潤: 0.3秒間自己付与を適用
		p, ok := c.Core.Combat.Player().(*avatar.Player)
		if !ok {
			panic("target 0 should be Player but is not!!")
		}
		p.ApplySelfInfusion(attributes.Hydro, 25, 0.3*60)

		// 5秒ごとにTick
		c.Core.Tasks.Add(c.barbaraSelfTick(healAmt, hpplus, skillInitF), 5*60)
	}
}

func (c *char) barbaraMelodyTick(ai combat.AttackInfo, skillInitF int) func() {
	return func() {
		// 上書きされていないことを確認
		if c.skillInitF != skillInitF {
			return
		}
		// バフが期限切れなら何もしない
		if c.Core.Status.Duration(barbSkillKey) == 0 {
			return
		}

		c.Core.Log.NewEvent("barbara melody loop ticking", glog.LogCharacterEvent, c.Index)

		// 0ダメージ攻撃で敵にのみヒットラグを発生させる
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1), -1, 0)

		// 1.5秒ごとにTick
		c.Core.Tasks.Add(c.barbaraMelodyTick(ai, skillInitF), 1.5*60)
	}
}
