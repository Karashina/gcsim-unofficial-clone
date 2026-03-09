package shenhe

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var (
	skillPressFrames []int
	skillHoldFrames  []int
)

const (
	skillPressCDStart  = 2
	skillPressHitmark  = 4
	skillHoldCDStart   = 31
	skillHoldHitmark   = 33
	holdParticleICDKey = "shenhe-hold-particle-icd"
	quillKey           = "shenhe-quill"
)

func init() {
	// スキル（単押し） -> x
	skillPressFrames = frames.InitAbilSlice(38) // 歩行
	skillPressFrames[action.ActionAttack] = 27
	skillPressFrames[action.ActionSkill] = 27
	skillPressFrames[action.ActionBurst] = 27
	skillPressFrames[action.ActionDash] = 21
	skillPressFrames[action.ActionJump] = 21
	skillPressFrames[action.ActionSwap] = 27

	// スキル（長押し） -> x
	// TODO: スキル（長押し） -> スキル（長押し）は52フレーム
	skillHoldFrames = frames.InitAbilSlice(78) // 歩行
	skillHoldFrames[action.ActionAttack] = 45
	skillHoldFrames[action.ActionSkill] = 45 // スキル（単押し）と仮定
	skillHoldFrames[action.ActionBurst] = 45
	skillHoldFrames[action.ActionDash] = 38
	skillHoldFrames[action.ActionJump] = 39
	skillHoldFrames[action.ActionSwap] = 44
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if p["hold"] != 0 {
		return c.skillHold(), nil
	}
	return c.skillPress(), nil
}

func (c *char) skillPress() action.Info {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Spring Spirit Summoning (Press)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               skillPress[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}

	c.Core.Tasks.Add(func() {
		snap := c.Snapshot(&ai)
		snap.Stats[attributes.DmgP] += c.c4()
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				nil,
				0.8,
			),
			0,
			c.makePressParticleCB(),
		)
	}, skillPressHitmark)

	if c.Base.Ascension >= 4 {
		c.Core.Tasks.Add(c.skillPressBuff, skillPressCDStart+1)
	}
	c.SetCDWithDelay(action.ActionSkill, 10*60, skillPressCDStart)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) makePressParticleCB() combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		// スキルがゲーム内で実際にキャラを移動させる - キャッチは90-110フレームの範囲、平均100として扱う
		c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Cryo, c.ParticleDelay)
	}
}

func (c *char) skillHold() action.Info {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Spring Spirit Summoning (Hold)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Cryo,
		Durability: 50,
		Mult:       skillHold[c.TalentLvlSkill()],
	}

	c.Core.Tasks.Add(func() {
		snap := c.Snapshot(&ai)
		snap.Stats[attributes.DmgP] += c.c4()
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.5}, 4),
			0,
			c.holdParticleCB,
		)
	}, skillHoldHitmark)

	if c.Base.Ascension >= 4 {
		c.Core.Tasks.Add(c.skillHoldBuff, skillHoldCDStart+1)
	}
	c.SetCDWithDelay(action.ActionSkill, 15*60, skillHoldCDStart+1)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) holdParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(holdParticleICDKey) {
		return
	}
	c.AddStatus(holdParticleICDKey, 0.5*60, true)
	// 粒子の生成タイミングは単押しEより少し遅い
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Cryo, c.ParticleDelay)
}

// 固有天賦2:
// 申鶴が仕組みの春神を使用後、近くのパーティーメンバー全員に以下の効果を付与する:
//
// - 短押し: 元素スキルと元素爆発のダメージが10秒間15%増加。
func (c *char) skillPressBuff() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(quillKey, 10*60, true) // 10秒間の持続時間
		char.SetTag(quillKey, 5)              // 単押し時は氷翎5回
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("shenhe-a4-press", 10*60),
			Amount: func(a *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				switch a.Info.AttackTag {
				case attacks.AttackTagElementalArt:
				case attacks.AttackTagElementalArtHold:
				case attacks.AttackTagElementalBurst:
				default:
					return nil, false
				}
				return c.skillBuff, true
			},
		})
	}
}

// 固有天賦2:
// 申鶴が仕組みの春神を使用後、近くのパーティーメンバー全員に以下の効果を付与する:
//
// - 長押し: 通常攻撃、重撃、落下攻撃のダメージが15秒間15%増加。
func (c *char) skillHoldBuff() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(quillKey, 15*60, true) // 15秒間の持続時間
		char.SetTag(quillKey, 7)              // 長押し時は氷翎5回
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("shenhe-a4-hold", 15*60),
			Amount: func(a *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				switch a.Info.AttackTag {
				case attacks.AttackTagNormal:
				case attacks.AttackTagExtra:
				case attacks.AttackTagPlunge:
				default:
					return nil, false
				}
				return c.skillBuff, true
			},
		})
	}
}

func (c *char) quillDamageMod() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		consumeStack := true
		if atk.Info.Element != attributes.Cryo {
			return false
		}

		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalBurst:
		case attacks.AttackTagElementalArt:
		case attacks.AttackTagElementalArtHold:
		case attacks.AttackTagNormal:
			consumeStack = c.Base.Cons < 6
		case attacks.AttackTagExtra:
			consumeStack = c.Base.Cons < 6
		case attacks.AttackTagPlunge:
		default:
			return false
		}

		char := c.Core.Player.ByIndex(atk.Info.ActorIndex)

		if !char.StatusIsActive(quillKey) {
			return false
		}

		if char.Tags[quillKey] > 0 {
			amt := skillpp[c.TalentLvlSkill()] * c.TotalAtk()
			if consumeStack { // 6凸
				char.Tags[quillKey]--
			}

			if c.Core.Flags.LogDebug {
				c.Core.Log.NewEvent("Shenhe Quill proc dmg add", glog.LogPreDamageMod, atk.Info.ActorIndex).
					Write("before", atk.Info.FlatDmg).
					Write("addition", amt).
					Write("effect_ends_at", char.StatusExpiry(quillKey)).
					Write("quill_left", char.Tags[quillKey])
			}

			atk.Info.FlatDmg += amt
			if c.Base.Cons >= 4 {
				atk.Callbacks = append(atk.Callbacks, c.c4CB)
			}
		}

		return false
	}, "shenhe-quill-hook")
}
