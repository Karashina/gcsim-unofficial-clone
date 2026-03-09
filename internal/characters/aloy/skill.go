package aloy

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const skillRelease = 20 // ボムのリリースフレーム、travelが上乗せ、その後bomb_delayが上乗せ

func init() {
	skillFrames = frames.InitAbilSlice(49) // E -> Dash
	skillFrames[action.ActionAttack] = 47  // E -> N1
	skillFrames[action.ActionBurst] = 48   // E -> Q
	skillFrames[action.ActionJump] = 47    // E -> J
	skillFrames[action.ActionSwap] = 66    // E -> Swap
}

const (
	rushingIceKey      = "rushingice"
	rushingIceDuration = 600
)

// 元素スキル - メインダメージ、ボムレット、コイル効果を処理
//
// 3つのパラメータを持つ：
//
// - "travel" = メインダメージまでのフレーム遅延、ボムレットはメインダメージ時に生成
//
// - "bomblets" = ヒットするボムレットの数
//
// - "bomb_delay" = ボムレットが爆発しコイルスタックが追加されるまでのフレーム遅延
//
// - ボムレットヒットのバリエーションが多すぎるため構文を短くできない。ここで簡略化して処理
func (c *char) Skill(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 5
	}

	bomblets, ok := p["bomblets"]
	if !ok {
		bomblets = 2
	}

	delay, ok := p["bomb_delay"]
	if !ok {
		delay = 0
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Freeze Bomb",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       skillMain[c.TalentLvlSkill()],
	}
	// TODO: 正確なスナップショットタイミング、ヒット/ボム生成時ではなく発動時にスナップショットと仮定
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 4),
		skillRelease,
		skillRelease+travel,
		c.makeParticleCB(),
	)

	// ボムレットは発動時にスナップショット
	ai = combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Chillwater Bomblets",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagElementalArt,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               skillBomblets[c.TalentLvlSkill()],
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}

	// ボムレットをキューに追加
	for i := 0; i < bomblets; i++ {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2),
			skillRelease+travel,
			skillRelease+travel+delay+((i+1)*6),
			c.coilStacks,
		)
	}

	c.SetCDWithDelay(action.ActionSkill, 20*60, 19)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillRelease,
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
		c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Cryo, c.ParticleDelay)
	}
}

// コイルスタックと関連効果（ラッシュアイスのトリガーを含む）を処理
func (c *char) coilStacks(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.coilICDExpiry > c.Core.F {
		return
	}
	// ラッシュアイス中はコイルスタックを獲得できない
	if c.StatusIsActive(rushingIceKey) {
		return
	}
	c.coils++
	c.coilICDExpiry = c.Core.F + 6

	c.Core.Log.NewEvent("coil stack gained", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.coils)

	c.a1()

	if c.coils == 4 {
		c.coils = 0
		c.rushingIce()
	}
}

// ラッシュアイス状態を処理
func (c *char) rushingIce() {
	c.AddStatus(rushingIceKey, rushingIceDuration, true)
	c.Core.Player.AddWeaponInfuse(c.Index, "aloy-rushing-ice", attributes.Cryo, 600, true, attacks.AttackTagNormal)

	// ラッシュアイス通常攻撃ボーナス
	val := make([]float64, attributes.EndStatType)
	val[attributes.DmgP] = skillRushingIceNABonus[c.TalentLvlSkill()]
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("aloy-rushing-ice", rushingIceDuration),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagNormal {
				return val, true
			}
			return nil, false
		},
	})

	c.a4()
}

// シミュレーション開始時にコイルModを追加
// アーロイのフィールド離脱後30秒までコイルが持続するため、動的にするのは容易ではない
func (c *char) coilMod() {
	val := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("aloy-coil-stacks", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagNormal && c.coils > 0 {
				val[attributes.DmgP] = skillCoilNABonus[c.coils-1][c.TalentLvlSkill()]
				return val, true
			}
			return nil, false
		},
	})
}

// コイルスタックをクリアするタイマーを開始するフィールド離脱フック
func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		if prev != c.Index {
			return false
		}
		c.lastFieldExit = c.Core.F

		c.Core.Tasks.Add(func() {
			if c.lastFieldExit != (c.Core.F - 30*60) {
				return
			}
			c.coils = 0
		}, 30*60)

		return false
	}, "aloy-exit")
}
