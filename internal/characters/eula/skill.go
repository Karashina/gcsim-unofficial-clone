package eula

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillPressFrames []int
var skillHoldFrames []int
var icewhirlHitmarks = []int{79, 92}

const (
	skillPressHitmark   = 20
	skillHoldHitmark    = 49
	pressParticleICDKey = "eula-press-particle-icd"
	holdParticleICDKey  = "eula-hold-particle-icd"
	a1Hitmark           = 108
	grimheartICD        = "eula-grimheart-icd"
	grimheartDuration   = "eula-grimheart-duration"
)

func init() {
	// skill (press) -> x
	skillPressFrames = frames.InitAbilSlice(48)
	skillPressFrames[action.ActionAttack] = 31
	skillPressFrames[action.ActionBurst] = 31
	skillPressFrames[action.ActionDash] = 29
	skillPressFrames[action.ActionJump] = 30
	skillPressFrames[action.ActionSwap] = 29

	// skill (hold) -> x
	skillHoldFrames = frames.InitAbilSlice(100)
	skillHoldFrames[action.ActionAttack] = 77
	skillHoldFrames[action.ActionBurst] = 77
	skillHoldFrames[action.ActionDash] = 75
	skillHoldFrames[action.ActionJump] = 75
	skillHoldFrames[action.ActionWalk] = 75
}

func (c *char) addGrimheartStack() {
	if !c.StatusIsActive(grimheartDuration) {
		c.grimheartStacks = 0
	}
	if c.grimheartStacks < 2 {
		c.grimheartStacks++
		c.Core.Log.NewEvent("eula: grimheart stack", glog.LogCharacterEvent, c.Index).
			Write("current count", c.grimheartStacks)
	}
	// 冷酷な心の持続時間を無条件で更新
	c.AddStatus(grimheartDuration, 1080, true) // 18秒
}

func (c *char) currentGrimheartStacks() int {
	if !c.StatusIsActive(grimheartDuration) {
		c.grimheartStacks = 0
		return 0
	}
	if c.grimheartStacks > 2 {
		c.grimheartStacks = 2
	}
	return c.grimheartStacks
}

func (c *char) consumeGrimheartStacks() {
	c.grimheartStacks = 0
	c.DeleteStatus(grimheartDuration)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if p["hold"] != 0 {
		return c.holdSkill(), nil
	}
	return c.pressSkill(), nil
}

func (c *char) pressSkill() action.Info {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Icetide Vortex",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           120,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               skillPress[c.TalentLvlSkill()],
		HitlagHaltFrames:   0.09 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}
	// ICDで制限されていなければ冷酷な心を1追加
	cb := func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(grimheartICD) {
			return
		}
		c.AddStatus(grimheartICD, 18, true)
		c.addGrimheartStack()
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 3.5),
		skillPressHitmark,
		skillPressHitmark,
		cb,
		c.pressParticleCB,
		c.burstStackCB,
	)

	c.SetCDWithDelay(action.ActionSkill, 60*4, 16)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) pressParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(pressParticleICDKey) {
		return
	}
	c.AddStatus(pressParticleICDKey, 0.3*60, true)

	count := 1.0
	if c.Core.Rand.Float64() < 0.5 {
		count = 2
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Cryo, c.ParticleDelay)
}

func (c *char) holdSkill() action.Info {
	// 長押しスキル
	// 296～341フレーム、CDは322で開始
	// 60fps = 108フレーム詠唱、CDは62フレーム目に開始するので+62フレーム必要
	lvl := c.TalentLvlSkill()
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Icetide Vortex (Hold)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               skillHold[lvl],
		HitlagHaltFrames:   0.12 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 5.5),
		skillHoldHitmark,
		skillHoldHitmark,
		c.holdParticleCB,
		c.burstStackCB,
	)

	v := c.currentGrimheartStacks()

	// 耐性ダウン
	var shredCB combat.AttackCBFunc
	if v > 0 {
		shredCB = func(a combat.AttackCB) {
			e, ok := a.Target.(*enemy.Enemy)
			if !ok {
				return
			}
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("eula-icewhirl-shred-cryo", 7*v*60),
				Ele:   attributes.Cryo,
				Value: -resRed[lvl],
			})
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("eula-icewhirl-shred-phys", 7*v*60),
				Ele:   attributes.Physical,
				Value: -resRed[lvl],
			})
		}
	}

	for i := 0; i < v; i++ {
		// 複数のブランドヒット
		//TODO: ヒットラグの影響を受けるか要再確認。設置物の可能性あり
		icewhirlAI := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Icetide Vortex (Icewhirl)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       icewhirl[lvl],
		}
		if i == 0 {
			// shizukaによると最初の渦はヒットラグの影響を受けない？
			c.Core.QueueAttack(
				icewhirlAI,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3.5),
				icewhirlHitmarks[i],
				icewhirlHitmarks[i],
				shredCB,
				c.burstStackCB,
			)
		} else {
			c.QueueCharTask(func() {
				// スタック用に間隔を空ける
				c.Core.QueueAttack(
					icewhirlAI,
					combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3.5),
					0,
					0,
					shredCB,
					c.burstStackCB,
				)
			}, icewhirlHitmarks[i])
		}
	}
	if v == 2 {
		c.a1()
	}

	// 1命ノ星座デバフ追加
	if c.Base.Cons >= 1 && v > 0 {
		//TODO: 持続時間が正しいか要確認
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("eula-c1", (6*v+6)*60),
			AffectedStat: attributes.PhyP,
			Amount: func() ([]float64, bool) {
				return c.c1buff, true
			},
		})
	}

	c.consumeGrimheartStacks()
	cd := 10
	if c.Base.Cons >= 2 {
		cd = 4 // 短押しと長押しのCDが同じ TODO: これが正しいか要確認
	}
	c.SetCDWithDelay(action.ActionSkill, cd*60, 46)

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

	count := 2.0
	if c.Core.Rand.Float64() < 0.5 {
		count = 3
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Cryo, c.ParticleDelay)
}
