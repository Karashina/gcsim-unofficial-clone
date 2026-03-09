package kokomi

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	skillHitmark   = 24
	particleICDKey = "kokomi-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(61)
	skillFrames[action.ActionDash] = 29
	skillFrames[action.ActionJump] = 29
}

// 元素スキル処理 - 初回ダメージインスタンスを処理
// 周囲の敵に水元素ダメージを与え、2秒ごとに付近のアクティブキャラクターを回復する。回復量は心海のHP上限に基づく。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	// スキル持続時間は約12.5秒
	// +1で同一フレーム問題を回避
	c.Core.Status.Add("kokomiskill", 12*60+30+1)

	d := c.createSkillSnapshot()

	// 即座に1Tick、その後2秒ごとに1Tickで合計7Tick
	c.swapEarlyF = -1
	c.skillLastUsed = c.Core.F
	c.skillFlatDmg = c.burstDmgBonus(d.Info.AttackTag)

	c.Core.Tasks.Add(func() { c.skillTick(d) }, skillHitmark)
	c.Core.Tasks.Add(c.skillTickTask(d, c.Core.F), skillHitmark+126)

	c.SetCDWithDelay(action.ActionSkill, 20*60, 20)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillHitmark,
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
	c.AddStatus(particleICDKey, 1*60, false)
	if c.Core.Rand.Float64() < 0.67 {
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Hydro, c.ParticleDelay)
	}
}

// スキル使用時と爆発使用時の両方で作成が必要なためヘルパー関数化
func (c *char) createSkillSnapshot() *combat.AttackEvent {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Bake-Kurage",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       skillDmg[c.TalentLvlSkill()],
	}
	snap := c.Snapshot(&ai)
	ae := combat.AttackEvent{
		Info:        ai,
		Pattern:     combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 3}, 6),
		SourceFrame: c.Core.F,
		Snapshot:    snap,
	}
	ae.Callbacks = append(ae.Callbacks, c.particleCB)
	return &ae
}

// 元素スキルの各Tickでダメージ・回復・粒子を処理するヘルパー関数
func (c *char) skillTick(d *combat.AttackEvent) {
	// スキルに爆発ボーナスのスナップショットがあるか確認
	// スナップショットは1Tick目と2Tick目の間
	if c.swapEarlyF > c.skillLastUsed && c.swapEarlyF < c.skillLastUsed+100 {
		d.Info.FlatDmg = c.skillFlatDmg
	} else {
		d.Info.FlatDmg = c.burstDmgBonus(d.Info.AttackTag)
	}

	// ダメージ処理
	c.Core.QueueAttackEvent(d, 0)

	// 回復処理
	if c.Core.Combat.Player().IsWithinArea(d.Pattern) {
		maxhp := d.Snapshot.Stats.MaxHP()
		src := skillHealPct[c.TalentLvlSkill()]*maxhp + skillHealFlat[c.TalentLvlSkill()]

		// 2凸: HP50%以下のキャラクターに対する化海月の回復ボーナス
		// 化海月: 心海のHP上限の4.5%分。
		if c.Base.Cons >= 2 {
			active := c.Core.Player.ActiveChar()
			if active.CurrentHPRatio() <= 0.5 {
				bonus := 0.045 * maxhp
				src += bonus
				c.Core.Log.NewEvent("kokomi c2 proc'd", glog.LogCharacterEvent, active.Index).
					Write("bonus", bonus)
			}
		}

		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.Active(),
			Message: "Bake-Kurage",
			Src:     src,
			Bonus:   d.Snapshot.Stats[attributes.Heal],
		})
	}
}

// 繰り返しスキルダメージTickを処理。フィールド上に海月は1体のみのため別関数に分離
// スキルはスナップショットするので、元のスナップショットを入力として受け取る
func (c *char) skillTickTask(originalSnapshot *combat.AttackEvent, src int) func() {
	return func() {
		c.Core.Log.NewEvent("Skill Tick Debug", glog.LogCharacterEvent, c.Index).
			Write("current dur", c.Core.Status.Duration("kokomiskill")).
			Write("skilllastused", c.skillLastUsed).
			Write("src", src)
		if c.Core.Status.Duration("kokomiskill") == 0 {
			return
		}

		// 古いスキル発動からのTickを停止し、そのソースからの追加Tickも防止
		if c.skillLastUsed > src {
			return
		}

		c.skillTick(originalSnapshot)

		c.Core.Tasks.Add(c.skillTickTask(originalSnapshot, src), 120)
	}
}
