package lauma

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillInitHitmark     = 30
	skillInitHitmarkHold = 37
	skillTicks           = 8
	skillInterval        = 117
	skillFirstTickDelay  = 65
	skillKey             = "lauma-skill"
	particleICDKey       = "lauma-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(41) // E -> D/J
}

// Tickタイミング用の切り上げヘルパー
func ceil(x float64) int {
	return int(math.Ceil(x))
}

// 元素スキル
// タップまたは長押しに応じて異なる効果の霜林の聖域を召喚する。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	skillPos := c.Core.Combat.Player()
	if p["hold"] == 1 && c.verdantDew > 0 {
		// 長押し
		// 翠露を1つ以上持っている時に発動可能。

		// Laumaは全ての翠露を消費し、永眠の讃歌を唱え、
		// 通常の範囲草元素ダメージ1回と、Lunar-Bloomダメージとみなされる範囲草元素ダメージ1回を与える。
		// 消費した翠露1つにつき、Laumaは月の歌を1スタック獲得する。
		// 長押しで元素スキルを発動するたびに、最大3つの翠露を消費できる。
		dewConsumed := min(c.verdantDew, 3)
		c.verdantDew = 0
		c.moonSong += dewConsumed

		em := c.Stat(attributes.EM)
		ai1 := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Runo: Dawnless Rest of Karsikko (E/Hold/Hit 1)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillHold1[c.TalentLvlSkill()],
		}

		c.Core.QueueAttack(
			ai1,
			combat.NewCircleHitOnTarget(skillPos, geometry.Point{Y: -1.5}, 5),
			skillInitHitmarkHold, skillInitHitmarkHold, c.particleCB, c.shredCB,
		)
		// 「Lunar-Bloomダメージとみなされる」範囲
		ai2 := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Runo: Dawnless Rest of Karsikko (E/Hold/Hit 2)",
			AttackTag:        attacks.AttackTagLBDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Dendro,
			IgnoreDefPercent: 1,
			FlatDmg:          (skillHold2[c.TalentLvlSkill()]*em*(1+c.LBBaseReactBonus(ai1)))*(1+((6*em)/(2000+em))+c.LBReactBonus(ai1)) + c.burstLBBuff,
		}
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		c.Core.QueueAttackWithSnap(
			ai2,
			snap,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget().Pos(), nil, 6),
			skillInitHitmarkHold,
			c.shredCB,
		)
		// 「Lunar-Bloomダメージとみなされる」範囲の終了
	} else {
		// 単押し
		// 範囲草元素ダメージを与える。
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Runo: Dawnless Rest of Karsikko (E/Press)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillPress[c.TalentLvlSkill()],
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(skillPos, geometry.Point{Y: -1.5}, 5),
			skillInitHitmark, skillInitHitmark, c.particleCB, c.shredCB,
		)
	}

	// スキルの持続時間とTickはヒットラグの影響を受けない
	c.skillSrc = c.Core.F

	for i := 0.0; i < skillTicks; i++ {
		c.Core.Tasks.Add(c.skillTick(c.skillSrc), skillFirstTickDelay+ceil(skillInterval*i))
	}
	c.AddStatus(skillKey, skillFirstTickDelay+ceil((skillTicks-1)*skillInterval), false)

	c.SetCD(action.ActionSkill, 12*60)
	c.a1() // A1ムーンサインバフを20秒間適用

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

// スキルの粒子生成コールバック
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	count := 1.0
	if c.Core.Rand.Float64() < 0.3 {
		count = 2.0
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Dendro, c.ParticleDelay)
	c.AddStatus(particleICDKey, 3.3*60, true)
}

// DoTのスキルTickロジック
func (c *char) skillTick(src int) func() {
	return func() {
		if src != c.skillSrc {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Frostgrove Sanctuary Attack DMG (E/DoT)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillDotATK[c.TalentLvlSkill()],
			FlatDmg:    skillDotEM[c.TalentLvlSkill()] * c.Stat(attributes.EM),
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6.5),
			0, 0, c.particleCB, c.shredCB,
		)
		c.c4()

		if c.Base.Cons >= 6 {
			// 元素熟知の185%に等しい追加範囲草元素ダメージを与える
			em := c.Stat(attributes.EM)
			c6ai := combat.AttackInfo{
				ActorIndex:       c.Index,
				Abil:             "Frostgrove Sanctuary (C6 Bonus)",
				AttackTag:        attacks.AttackTagLBDamage,
				StrikeType:       attacks.StrikeTypeDefault,
				Element:          attributes.Dendro,
				IgnoreDefPercent: 1,
			}

			snapc6 := combat.Snapshot{
				CharLvl: c.Base.Level,
			}
			c6ai.FlatDmg = (1.85*em*(1+c.LBBaseReactBonus(c6ai)))*(1+((6*em)/(2000+em))+c.LBReactBonus(c6ai)) + c.burstLBBuff // 元素熟知の185%
			snapc6.Stats[attributes.CR] = c.Stat(attributes.CR)
			snapc6.Stats[attributes.CD] = c.Stat(attributes.CD)

			c.Core.QueueAttackWithSnap(
				c6ai,
				snapc6,
				combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget().Pos(), nil, 6.5),
				1,
				c.shredCB,
			)

			c.paleHymn += 3 // 淡き讃歌を2+1スタック獲得
			c.AddStatus("pale-hymn-window", 15*60, true)
		}
	}
}

func (c *char) shredCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("lauma-e-dendro", 10*60),
		Ele:   attributes.Dendro,
		Value: -1 * skillRESShred[c.TalentLvlSkill()],
	})
}
