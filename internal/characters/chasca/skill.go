package chasca

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
var skillCancelFrames []int

const (
	skillHitmarks      = 3
	plungeAvailableKey = "chasca-plunge-available"
)

func init() {
	skillFrames = frames.InitAbilSlice(27) // E -> E
	skillFrames[action.ActionAttack] = 5
	skillFrames[action.ActionAim] = 17
	skillFrames[action.ActionBurst] = 6
	skillFrames[action.ActionDash] = 6
	skillFrames[action.ActionSwap] = 586 + 37 // ナイトソウルが尽きて地面に落下するのを待つ

	skillCancelFrames = frames.InitAbilSlice(40) // E -> Dash/Jump
	skillCancelFrames[action.ActionAttack] = 38
	skillCancelFrames[action.ActionAim] = 38
	skillCancelFrames[action.ActionLowPlunge] = 2
	skillCancelFrames[action.ActionBurst] = 39
	skillCancelFrames[action.ActionWalk] = 38
	skillCancelFrames[action.ActionSwap] = 37
}

func (c *char) reduceNightsoulPoints(val float64) {
	c.nightsoulState.ConsumePoints(val)
	c.checkNS()
}

// 現在のナイトソウルポイント数を確認し、不足している場合はナイトソウルを抜ける。チェック後のNSステータスを返す
func (c *char) checkNS() {
	if c.nightsoulState.Points() < 0.001 {
		c.exitNightsoul()
	}
}

// NSが切れた場合はskillCancelFramesを、それ以外は入力されたフレームを返す
func (c *char) skillNextFrames(f func(next action.Action) int, extraDelay int) func(next action.Action) int {
	// これはアクション開始からのヒットラグ効果経過時間を計算するために使用
	actionStart := c.TimePassed
	actionEnd := -1
	return func(next action.Action) int {
		if c.nightsoulState.HasBlessing() {
			return f(next)
		}
		if actionEnd < 0 {
			actionEnd = c.TimePassed
		}
		// TODO: この場合に落下アニメーションを「落下/待機」に設定する？
		return actionEnd - actionStart + skillCancelFrames[next] + extraDelay
	}
}

func (c *char) enterNightsoul() {
	c.nightsoulState.EnterBlessing(80)
	c.nightsoulSrc = c.Core.F
	c.Core.Tasks.Add(c.nightsoulPointReduceFunc(c.nightsoulSrc), 6)
	c.NormalHitNum = 1
	c.NormalCounter = 0
	c.skillParticleICD = false
	c.c6Used = false
}

func (c *char) nigthsoulFallingMsg() {
	c.Core.Log.NewEvent("nightsoul ended, falling", glog.LogCharacterEvent, c.Index)
}
func (c *char) exitNightsoul() {
	if !c.nightsoulState.HasBlessing() {
		return
	}

	switch c.Core.Player.CurrentState() {
	case action.AimState:
		// NS終了後も最大10フレームまで弾丸の充填を継続
		c.QueueCharTask(c.fireBullets, skillAimChargeDelay)
		c.QueueCharTask(c.nigthsoulFallingMsg, skillAimFallDelay)
	case action.Idle:
		c.Core.Player.SwapCD = 37
		c.nigthsoulFallingMsg()
	case action.DashState, action.NormalAttackState:
		c.nigthsoulFallingMsg()
	}

	c.nightsoulState.ExitBlessing()
	c.nightsoulState.ClearPoints()
	c.nightsoulSrc = -1
	c.SetCD(action.ActionSkill, 6.5*60)
	c.NormalHitNum = normalHitNum
	c.NormalCounter = 0
	c.AddStatus(plungeAvailableKey, 26, true)
}

func (c *char) nightsoulPointReduceFunc(src int) func() {
	return func() {
		if c.nightsoulSrc != src {
			return
		}
		c.reduceNightsoulPoints(0.8)
		// 6フレームごとに0.8ポイント減少、つまり1秒あたり8
		c.Core.Tasks.Add(c.nightsoulPointReduceFunc(src), 6)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.exitNightsoul()
		return action.Info{
			Frames:          frames.NewAbilFunc(skillCancelFrames),
			AnimationLength: skillFrames[action.InvalidAction],
			CanQueueAfter:   skillFrames[action.ActionLowPlunge], // 最速キャンセル
			State:           action.SkillState,
		}, nil
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Spirit Reins, Shadow Hunt",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           skillResonance[c.TalentLvlSkill()],
	}
	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		nil,
		5.5,
	)
	c.Core.QueueAttack(ai, ap, skillHitmarks, skillHitmarks)
	c.enterNightsoul()
	return action.Info{
		Frames:          c.skillNextFrames(frames.NewAbilFunc(skillFrames), 0),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.skillParticleICD {
		return
	}
	c.skillParticleICD = true
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Anemo, c.ParticleDelay)
}
