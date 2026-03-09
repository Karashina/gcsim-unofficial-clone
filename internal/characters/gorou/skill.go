package gorou

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillHitmark   = 34
	particleICDKey = "gorou-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(47) // E -> N1/Q
	skillFrames[action.ActionDash] = 33    // E -> D
	skillFrames[action.ActionJump] = 33    // E -> J
	skillFrames[action.ActionSwap] = 46    // E -> Swap
}

/*
*
パーティ内の岩元素キャラクターの数に応じて、スキルAoE内のアクティブキャラクターに
最大3つのバフを付与（発動時の人数で決定）:
• 岩元素キャラ1人: 「積石」追加 - 防御力ボーナス。
• 岩元素キャラ2人: 「集岩」追加 - 中断耐性向上。
• 岩元素キャラ3人: 「碎岩」追加 - 岩元素ダメージボーナス。
ゴローは「大将旗」をフィールド上に1つしか配置できない。キャラクターは1つの「大将旗」の恩恵しか受けられない。
パーティメンバーがフィールドを離れると、アクティブなバフは2秒間持続する。
*
*/
func (c *char) Skill(p map[string]int) (action.Info, error) {
	c.Core.Tasks.Add(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Inuzaka All-Round Defense",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   80,
			Element:    attributes.Geo,
			Durability: 25,
			Mult:       skill[c.TalentLvlSkill()],
			FlatDmg:    c.a4Skill(),
		}
		c.eFieldArea = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 8)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.eFieldArea.Shape.Pos(), nil, 5), 0, 0, c.particleCB)

		// 元素スキル
		// ゴローのフィールドはベネットのフィールドと同様に動作する
		// ただし元素爆発フィールドがアクティブの場合、元素スキルフィールドは配置できない
		if c.Core.Status.Duration(generalGloryKey) == 0 {
			c.eFieldSrc = c.Core.F
			c.Core.Tasks.Add(c.gorouSkillBuffField(c.Core.F), 17) // 17にして最後のTickを取得

			// 大将旗のステータスを追加、10秒
			c.Core.Status.Add(generalWarBannerKey, 600)
		}

		// 6凸
		if c.Base.Cons == 6 {
			c.c6()
		}
	}, skillHitmark)

	// 10秒CD
	c.SetCDWithDelay(action.ActionSkill, 600, 32)

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
	c.AddStatus(particleICDKey, 0.2*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Geo, c.ParticleDelay)
}

// Tickをキューに追加する再帰関数
func (c *char) gorouSkillBuffField(src int) func() {
	return func() {
		// 上書きされている場合は何もしない
		if c.eFieldSrc != src {
			return
		}
		// 両方のフィールドが期限切れなら何もしない
		eActive := c.Core.Status.Duration(generalWarBannerKey) > 0
		qActive := c.Core.Status.Duration(generalGloryKey) > 0
		if !eActive && !qActive {
			return
		}
		// 元素スキルのみアクティブでプレイヤーがフィールド範囲外の場合は何もしない
		// 元素爆発がアクティブならプレイヤーは常にフィールド内
		if eActive && !qActive && !c.Core.Combat.Player().IsWithinArea(c.eFieldArea) {
			return
		}

		// 岩元素キャラ数に応じたバフをアクティブキャラに追加
		// 既存のモディファイアを上書きして問題ない
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(defenseBuffKey, 120), // 2秒間持続
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return c.gorouBuff, true
			},
		})

		// 0.3秒ごとにTick
		c.Core.Tasks.Add(c.gorouSkillBuffField(src), 18)
	}
}
