package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	skillHitmark          = 25
	gravityRippleDuration = 1500 // 25 seconds
	gravityRippleInterval = 119
	gravityLimit          = 60
)

func init() {
	skillFrames = frames.InitAbilSlice(35)
	skillFrames[action.ActionBurst] = 30
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 初期スキルダメージ（AoE水元素ダメージ）
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Eternal Tides (E)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
	}
	ai.FlatDmg = c.MaxHP() * skillDmg[c.TalentLvlSkill()]

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
	c.Core.QueueAttack(ai, ap, skillHitmark, skillHitmark, c.particleCB)

	// Gravity Rippleを有効化
	c.AddStatus(skillKey, gravityRippleDuration, true)
	c.AddStatus(gravityRippleKey, gravityRippleDuration, true)

	// 重力をリセット
	c.gravity = 0
	c.gravityLC = 0
	c.gravityLB = 0
	c.gravityLCrs = 0

	// Gravity RippleのTickを開始
	c.gravityRippleSrc = c.Core.F
	c.gravityRippleExp = c.Core.F + gravityRippleDuration
	c.Core.Tasks.Add(c.gravityRippleTick(c.Core.F), gravityRippleInterval)

	// 1凸: スキル発動時にGravity Interference効果をトリガー（15sに1回）
	if c.Base.Cons >= 1 {
		c.c1OnSkill()
	}

	// クールダウンを設定（17s）
	c.SetCDWithDelay(action.ActionSkill, 17*60, skillHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// particleCBは粒子生成を処理
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 60, false)

	// 平均1.333個の粒子（0/1/2の重み付き分布）
	count := 1.0
	if c.Core.Rand.Float64() < 0.333 {
		count = 2.0
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Hydro, c.ParticleDelay)
}

// gravityRippleTickは定期的なGravity Rippleダメージを処理
func (c *char) gravityRippleTick(src int) func() {
	return func() {
		if c.gravityRippleSrc != src {
			return
		}
		if !c.StatusIsActive(gravityRippleKey) {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Gravity Ripple (E)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
		}
		ai.FlatDmg = c.MaxHP() * gravityRippleDmg[c.TalentLvlSkill()]

		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)

		// 次のTickをスケジュール
		c.Core.Tasks.Add(c.gravityRippleTick(src), gravityRippleInterval)
	}
}

// subscribeToLunarReactionsはLunar反応イベントを購読してGravity蓄積を行う
func (c *char) subscribeToLunarReactions() {
	// Lunar-Chargedを購読
	c.Core.Events.Subscribe(event.OnLunarCharged, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		c.Core.Tasks.Add(func() {
			c.accumulateGravity("lc")
		}, 1)
		return false
	}, "columbina-gravity-lc")

	// Lunar-Bloomを購読
	c.Core.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		c.Core.Tasks.Add(func() {
			c.accumulateGravity("lb")
		}, 1)
		return false
	}, "columbina-gravity-lb")

	// Lunar-Crystallizeを購読
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		c.Core.Tasks.Add(func() {
			c.accumulateGravity("lcrs")
		}, 1)
		return false
	}, "columbina-gravity-lcrs")

	// Lunarダメージイベントを購読（Lunar反応ダメージ発生時）
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}

		var lunarType string
		switch ae.Info.AttackTag {
		case attacks.AttackTagLCDamage:
			lunarType = "lc"
		case attacks.AttackTagLBDamage:
			lunarType = "lb"
		case attacks.AttackTagLCrsDamage:
			lunarType = "lcrs"
		default:
			return false
		}

		c.Core.Tasks.Add(func() {
			c.accumulateGravity(lunarType)
		}, 1)
		return false
	}, "columbina-gravity-damage")
}

// accumulateGravityはLunar反応からGravityを蓄積する
func (c *char) accumulateGravity(lunarType string) {
	// newMoonOmenKeyが有効な場合、持続時間を更新しアクティブタイプを更新
	// 無効の場合は蓄積プロセスを開始
	if c.StatusIsActive(newMoonOmenKey) {
		// 持続時間を120s最大まで更新
		// アクティブタイプを新しいトリガーに更新
		c.activeGravityType = lunarType
		c.AddStatus(newMoonOmenKey, 120, false)
		return
	}

	// 初回トリガーまたは期限切れ
	c.activeGravityType = lunarType
	c.AddStatus(newMoonOmenKey, 120, false)

	// Tickerを開始
	interval := 6
	if c.Base.Cons >= 2 {
		interval = 4
	}
	c.Core.Tasks.Add(c.gravityTicker(), interval)
}

func (c *char) gravityTicker() func() {
	return func() {
		// ステータスが期限切れの場合は停止
		if !c.StatusIsActive(newMoonOmenKey) {
			return
		}

		// Gravity蓄積
		// Gravity蓄積率: 12フレームごとに2（基本）
		// 2凸: 9フレームごとに2
		addAmount := 2

		c.gravity += addAmount
		if c.gravity > gravityLimit {
			c.gravity = gravityLimit
		}

		// 特定タイプのバケットに追加
		switch c.activeGravityType {
		case "lc":
			c.gravityLC += addAmount
		case "lb":
			c.gravityLB += addAmount
		case "lcrs":
			c.gravityLCrs += addAmount
		}

		// 上限に達したらGravity Interferenceをトリガー
		// 注意: ステータス（New Moon's Omen）は持続するため、リセット/トリガー後も蓄積は継続
		if c.gravity >= gravityLimit {
			c.triggerGravityInterference()
		}

		// 次のTickをスケジュール
		// 基本: 12フレームごとに2
		// 2凸: 9フレームごとに2
		interval := 12
		if c.Base.Cons >= 2 {
			interval = 9
		}
		c.Core.Tasks.Add(c.gravityTicker(), interval)
	}
}

// triggerGravityInterferenceは支配的なLunarタイプに基づいてGravity Interferenceをトリガー
func (c *char) triggerGravityInterference() {
	dominantType := c.getDominantLunarType()

	c.Core.Log.NewEvent("gravity interference triggered", glog.LogCharacterEvent, c.Index).
		Write("dominant_type", dominantType).
		Write("gravity", c.gravity)

	// 固有天賦1: Lunacy効果を獲得
	if c.Base.Ascension >= 1 {
		c.a1OnGravityInterference()
	}

	// 2凸: Lunar Brillianceを獲得
	if c.Base.Cons >= 2 {
		c.c2OnGravityInterference(dominantType)
	}

	// 4凸: エネルギー回復とダメージボーナス
	if c.Base.Cons >= 4 {
		c.c4OnGravityInterference(dominantType)
	}

	switch dominantType {
	case "lc":
		c.gravityInterferenceLC()
	case "lb":
		c.gravityInterferenceLB()
	case "lcrs":
		c.gravityInterferenceLCrs()
	}

	c.c1OnGravityInterference()

	if !c.StatusIsActive(c1GravitySkipKey) {
		// 重力をリセット
		c.gravity = 0
		c.gravityLC = 0
		c.gravityLB = 0
		c.gravityLCrs = 0
	} else {
		c.DeleteStatus(c1GravitySkipKey)
	}
}

// gravityInterferenceLCは雷元素のAoEダメージを与える（Lunar-Chargedダメージとして扱われる）
func (c *char) gravityInterferenceLC() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Gravity Interference (Lunar-Charged / E)",
		AttackTag:        attacks.AttackTagLCDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// HP係数 + Lunar-Charged計算式
	em := c.Stat(attributes.EM)
	baseDmg := c.MaxHP() * gravityInterfLC[c.TalentLvlSkill()] * 3
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + c.LCBaseReactBonus(ai)) * (1 + emBonus + c.LCReactBonus(ai))

	// 4凸ボーナス
	if c.Base.Cons >= 4 && !c.StatusIsActive(c4IcdKey) {
		ai.FlatDmg += c.MaxHP() * c4HPBonusLC
		c.AddStatus(c4IcdKey, 15*60, false)
	}

	ai.FlatDmg *= (1 + c.ElevationBonus(ai))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.Core.QueueAttackWithSnap(ai, snap, ap, 10)

	// イベントを発行
	enemies := c.Core.Combat.EnemiesWithinArea(ap, nil)
	if len(enemies) > 0 {
		ae := &combat.AttackEvent{Info: ai}
		c.Core.Events.Emit(event.OnLunarCharged, enemies[0], ae)
	}
}

// gravityInterferenceLBは5つのMoondew Sigilを発射し草元素ダメージを与える（Lunar-Bloomダメージとして扱われる）
func (c *char) gravityInterferenceLB() {
	for i := 0; i < 5; i++ {
		delay := 10 + i*6
		c.Core.Tasks.Add(func() {
			ai := combat.AttackInfo{
				ActorIndex:       c.Index,
				Abil:             "Gravity Interference (Lunar-Bloom / E)",
				AttackTag:        attacks.AttackTagLBDamage,
				ICDTag:           attacks.ICDTagNone,
				ICDGroup:         attacks.ICDGroupDefault,
				StrikeType:       attacks.StrikeTypeDefault,
				Element:          attributes.Dendro,
				Durability:       0,
				IgnoreDefPercent: 1,
			}

			// HP係数 + Lunar-Bloom計算式
			em := c.Stat(attributes.EM)
			baseDmg := c.MaxHP() * gravityInterfLB[c.TalentLvlSkill()]
			emBonus := (6 * em) / (2000 + em)
			ai.FlatDmg = baseDmg * (1 + c.LBBaseReactBonus(ai)) * (1 + emBonus + c.LBReactBonus(ai))

			// 4凸ボーナス
			if c.Base.Cons >= 4 && !c.StatusIsActive(c4IcdKey) {
				ai.FlatDmg += c.MaxHP() * c4HPBonusLB
				c.AddStatus(c4IcdKey, 15*60, false)
			}

			ai.FlatDmg *= (1 + c.ElevationBonus(ai))

			snap := combat.Snapshot{CharLvl: c.Base.Level}
			snap.Stats[attributes.CR] = c.Stat(attributes.CR)
			snap.Stats[attributes.CD] = c.Stat(attributes.CD)

			closest := c.Core.Combat.ClosestEnemy(c.Core.Combat.Player().Pos())
			if closest == nil {
				return
			}
			ap := combat.NewCircleHitOnTarget(closest.Pos(), nil, 2)
			c.Core.QueueAttackWithSnap(ai, snap, ap, 0)

		}, delay)
	}
}

// gravityInterferenceLCrsは岩元素のAoEダメージを与える（Lunar-Crystallizeダメージとして扱われる）
func (c *char) gravityInterferenceLCrs() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Gravity Interference (Lunar-Crystallize / E)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// HP係数 + Lunar-Crystallize計算式
	em := c.Stat(attributes.EM)
	baseDmg := c.MaxHP() * gravityInterfLCrs[c.TalentLvlSkill()] * 1.6
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai)) * (1 + emBonus + c.LCrsReactBonus(ai))

	// 4凸ボーナス
	if c.Base.Cons >= 4 && !c.StatusIsActive(c4IcdKey) {
		ai.FlatDmg += c.MaxHP() * c4HPBonusLCrs
		c.AddStatus(c4IcdKey, 15*60, false)
	}

	ai.FlatDmg *= (1 + c.ElevationBonus(ai))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.Core.QueueAttackWithSnap(ai, snap, ap, 10)
}
