package itto

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var (
	chargeFrames   [][]int
	chargeHitmarks = []int{89, 51, 24, 71}
	chargeHitboxes = [][][]float64{{{3}, {3.8, 5.5}, {3.8, 5.5}, {3.5}}, {{4}, {5, 7}, {5, 7}, {4.3}}}
	chargeOffsets  = [][]float64{{0, -2, -2, 0.6}, {0, -2.5, -2.5, 0.8}}
)

func init() {
	chargeFrames = make([][]int, EndSlashType)

	// ActionChargeフレームはスラッシュごとに異なる
	// CA1/CA2 -> CAF, CA0 -> CA0 のフレームは ActionInfo.Frames で処理

	// CA0 -> x
	chargeFrames[SaichiSlash] = frames.InitAbilSlice(131) // 通常攻撃/CA1/CAFフレーム
	chargeFrames[SaichiSlash][action.ActionDash] = chargeHitmarks[SaichiSlash]
	chargeFrames[SaichiSlash][action.ActionJump] = chargeHitmarks[SaichiSlash]
	chargeFrames[SaichiSlash][action.ActionSwap] = 130

	// CA1 -> x
	chargeFrames[LeftSlash] = frames.InitAbilSlice(104) // 通常攻撃フレーム
	chargeFrames[LeftSlash][action.ActionCharge] = 57   // CA2フレーム
	chargeFrames[LeftSlash][action.ActionSkill] = chargeHitmarks[LeftSlash]
	chargeFrames[LeftSlash][action.ActionBurst] = chargeHitmarks[LeftSlash]
	chargeFrames[LeftSlash][action.ActionDash] = chargeHitmarks[LeftSlash]
	chargeFrames[LeftSlash][action.ActionJump] = chargeHitmarks[LeftSlash]
	chargeFrames[LeftSlash][action.ActionSwap] = chargeHitmarks[LeftSlash]

	// CA2 -> x
	chargeFrames[RightSlash] = frames.InitAbilSlice(77) // 通常攻撃フレーム
	chargeFrames[RightSlash][action.ActionCharge] = 29  // CA1フレーム
	chargeFrames[RightSlash][action.ActionSkill] = chargeHitmarks[RightSlash]
	chargeFrames[RightSlash][action.ActionBurst] = chargeHitmarks[RightSlash]
	chargeFrames[RightSlash][action.ActionDash] = chargeHitmarks[RightSlash]
	chargeFrames[RightSlash][action.ActionJump] = chargeHitmarks[RightSlash]
	chargeFrames[RightSlash][action.ActionSwap] = chargeHitmarks[RightSlash]

	// CAF -> x
	chargeFrames[FinalSlash] = frames.InitAbilSlice(109) // 通常攻撃/CA0フレーム
	chargeFrames[FinalSlash][action.ActionSkill] = 76
	chargeFrames[FinalSlash][action.ActionBurst] = 76
	chargeFrames[FinalSlash][action.ActionDash] = chargeHitmarks[FinalSlash]
	chargeFrames[FinalSlash][action.ActionJump] = chargeHitmarks[FinalSlash]
	chargeFrames[FinalSlash][action.ActionSwap] = 76
}

type SlashType int

const (
	InvalidSlash SlashType = iota - 1
	SaichiSlash            // CA0
	LeftSlash              // CA1
	RightSlash             // CA2
	FinalSlash             // CAF
	EndSlashType
)

var slashName = []string{
	"Saichimonji Slash",
	"Arataki Kesagiri Combo Slash Left",
	"Arataki Kesagiri Combo Slash Right",
	"Arataki Kesagiri Final Slash",
}

func (s SlashType) String(abil bool) string {
	// スキル名のために Left / Right を除去
	if abil && (s == LeftSlash || s == RightSlash) {
		return "Arataki Kesagiri Combo Slash"
	}
	return slashName[s]
}

func (s SlashType) Next(stacks int, c6Proc bool) SlashType {
	switch s {
	// stacks=1になるまでCA1/CA2をループ
	case LeftSlash:
		if stacks == 1 && !c6Proc {
			return FinalSlash
		}
		return RightSlash
	case RightSlash:
		if stacks == 1 && !c6Proc {
			return FinalSlash
		}
		return LeftSlash

	// idle/CA0/CAF -> 重撃（スタックに基づく）
	default:
		switch {
		case stacks == 1 && !c6Proc:
			return FinalSlash
		case stacks == 1 && c6Proc:
			return LeftSlash
		case stacks > 1:
			return LeftSlash
		}
		return SaichiSlash
	}
}

func (c *char) windupFrames(prevSlash, curSlash SlashType) int {
	switch c.Core.Player.CurrentState() {
	// 通常攻撃 -> x
	case action.NormalAttackState:
		switch curSlash {
		// NA -> CA0
		case SaichiSlash:
			switch c.NormalCounter - 1 {
			case 0:
				return 14
			case 1, 2:
				return 21
			}
		// NA -> CA1/CAF
		case LeftSlash, FinalSlash:
			return 10
		}

	// charge -> x
	case action.ChargeAttackState:
		switch curSlash {
		// CAF->CA0
		case SaichiSlash:
			if prevSlash == FinalSlash {
				return 14
			}
		// CA0/CA2/CAF -> CA1
		case LeftSlash:
			switch prevSlash {
			// CA0/CAF -> CA1
			case SaichiSlash, FinalSlash:
				return 17
			// CA2 -> CA1
			case RightSlash:
				return 28
			}
		// CA0/CA1/CA2/CAF -> CAF
		case FinalSlash:
			switch prevSlash {
			// CA0/CAF -> CAF
			case SaichiSlash, FinalSlash:
				return 17
			// CA1/CA2 -> CAF
			case LeftSlash, RightSlash:
				return 25
			}
		}

	// スキル -> x
	case action.SkillState:
		switch curSlash {
		// E -> CA0
		case SaichiSlash:
			return 14
		// E -> CA1/CAF
		case LeftSlash, FinalSlash:
			return 17
		}

	// 低空/高空落下攻撃 -> x
	case action.PlungeAttackState:
		switch curSlash {
		// 落下攻撃 -> CA0
		case SaichiSlash:
			return 11
		// 落下攻撃 -> CA1/CAF
		case LeftSlash, FinalSlash:
			return 10
		}
	}

	return 0
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   0.10 * 60, // デフォルトはCA0/CAFのヒットラグ
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	prevSlash := c.slashState
	stacks := c.Tags[strStackKey]
	c.slashState = prevSlash.Next(stacks, c.c6Proc)

	// スキップするフレーム数を算出
	windup := c.windupFrames(prevSlash, c.slashState)

	// ヒットラグと天賦%を処理
	ai.Abil = c.slashState.String(true)
	c.Core.Log.NewEvent("performing CA", glog.LogCharacterEvent, c.Index).
		Write("slash", c.slashState.String(false)).
		Write("stacks", stacks)

	switch c.slashState {
	case SaichiSlash:
		ai.PoiseDMG = 120
		ai.Mult = saichiSlash[c.TalentLvlAttack()]
	case LeftSlash, RightSlash:
		ai.PoiseDMG = 81.7
		ai.Mult = akCombo[c.TalentLvlAttack()]
		haltFrames := 0.03 // 消費スタック >= 3
		switch c.stacksConsumed {
		case 0:
			ai.PoiseDMG = 143.4
			haltFrames = 0.07
		case 1:
			haltFrames = 0.05
		}
		ai.HitlagHaltFrames = haltFrames * 60
	case FinalSlash:
		ai.PoiseDMG = 143.4
		ai.Mult = akFinal[c.TalentLvlAttack()]
	}

	// A4を適用
	if c.slashState != SaichiSlash {
		c.a4(&ai)
	}

	// ヒットボックスのインデックス用
	burstIndex := 0
	if c.StatusIsActive(burstBuffKey) {
		burstIndex = 1
	}
	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: chargeOffsets[burstIndex][c.slashState]},
		chargeHitboxes[burstIndex][c.slashState][0],
	)
	if c.slashState == LeftSlash || c.slashState == RightSlash {
		ap = combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: chargeOffsets[burstIndex][c.slashState]},
			chargeHitboxes[burstIndex][c.slashState][0],
			chargeHitboxes[burstIndex][c.slashState][1],
		)
	}
	// TODO: ヒットマークが攻撃速度で調整されていない
	// TODO: 一斗の重撃は重撃開始時にスナップショットされるか？（現在は開始時と仮定）
	c.Core.QueueAttack(ai, ap, 0, chargeHitmarks[c.slashState]-windup)

	// 6凸: 50%の確率で怒髪衝天スタックを消費しない。
	if !c.c6Proc {
		c.addStrStack("charge", -1)
	}

	// 攻撃速度を増加
	c.a1Update(c.slashState)

	// フレーム関数に必要
	curSlash := c.slashState
	c.c6Proc = c.Base.Cons >= 6 && c.Core.Rand.Float64() < 0.5
	atkspd := c.Stat(attributes.AtkSpd)

	return action.Info{
		Frames: func(next action.Action) int {
			f := chargeFrames[curSlash][next]

			if next == action.ActionCharge {
				nextSlash := curSlash.Next(c.Tags[strStackKey], c.c6Proc)
				switch nextSlash {
				// CA1/CA2 -> CAF フレームを処理
				case FinalSlash:
					switch curSlash {
					case LeftSlash: // CA1 -> CAF
						f = 60
					case RightSlash: // CA2 -> CAF
						f = 32
					}
				// CA0 -> CA0 フレームを処理
				case SaichiSlash:
					if curSlash == SaichiSlash {
						f = 500
					}
				}
			}

			return frames.AtkSpdAdjust(f-windup, atkspd)
		},
		AnimationLength: chargeFrames[curSlash][action.InvalidAction] - windup,
		CanQueueAfter:   chargeHitmarks[curSlash] - windup,
		State:           action.ChargeAttackState,
	}, nil
}
