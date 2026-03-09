package electro

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames [][]int

const skillHitmark = 21

func init() {
	skillFrames = make([][]int, 2)

	// 男性
	skillFrames[0] = frames.InitAbilSlice(57) // E -> N1
	skillFrames[0][action.ActionBurst] = 56   // E -> Q
	skillFrames[0][action.ActionDash] = 42    // E -> D
	skillFrames[0][action.ActionJump] = 42    // E -> J
	skillFrames[0][action.ActionSwap] = 56    // E -> Swap

	// 女性
	skillFrames[1] = frames.InitAbilSlice(57) // E -> N1/Q
	skillFrames[1][action.ActionDash] = 42    // E -> D
	skillFrames[1][action.ActionJump] = 42    // E -> J
	skillFrames[1][action.ActionSwap] = 55    // E -> Swap
}

func (c *Traveler) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lightning Blade",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	snap := c.Snapshot(&ai)

	hits, ok := p["hits"]
	if !ok {
		hits = 1
	} else if hits > 3 {
		hits = 3
	}

	maxAmulets := 2
	if c.Base.Cons >= 1 {
		maxAmulets = 3
	}

	// 既存のアミュレットをクリア
	c.abundanceAmulets = 0

	// 生成されるアミュレット数を制限するパラメータを受け入れ
	pMaxAmulets, ok := p["max_amulets"]
	if ok && pMaxAmulets < maxAmulets {
		maxAmulets = pMaxAmulets
	}

	// 元素スキルが押されたフレームから、キャラクターがアミュレットを拾えるようになるまで平均1.79秒
	// https://library.keqingmains.com/evidence/characters/electro/traveler-electro#amulets-delay
	amuletDelay := p["amulet_delay"]
	// 1.79秒より早くならないようにする
	if amuletDelay < 107 {
		amuletDelay = 107 // ~1.79s
	}

	amuletCB := func(a combat.AttackCB) {
		// 生成数が上限未満の場合にアミュレットを生成
		if c.abundanceAmulets >= maxAmulets {
			return
		}
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}

		// 攻撃1回につきアミュレット1個
		c.abundanceAmulets++
		c.SetTag("generated", c.abundanceAmulets)

		c.Core.Log.NewEvent("travelerelectro abundance amulet generated", glog.LogCharacterEvent, c.Index).
			Write("amulets", c.abundanceAmulets)
	}

	for i := 0; i < hits; i++ {
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewBoxHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				nil,
				0.1,
				0.6,
			),
			skillHitmark,
			c.makeParticleCB(), // 各刃が命中すると1粒子を生成
			amuletCB,
		)
	}

	// アミュレットの回収を試みる
	c.Core.Tasks.Add(func() {
		active := c.Core.Player.ActiveChar()
		c.collectAmulets(active)
	}, amuletDelay)

	c.SetCDWithDelay(action.ActionSkill, 810, 20) // 13.5s, starts 20 frames in

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames[c.gender]),
		AnimationLength: skillFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   skillFrames[c.gender][action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *Traveler) makeParticleCB() combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Electro, c.ParticleDelay)
	}
}

func (c *Traveler) collectAmulets(collector *character.CharWrapper) bool {
	// 回収するアミュレットがない場合はreturn
	if c.abundanceAmulets <= 0 {
		return false
	}

	// 利用可能な全アミュレットが同時に回収されると仮定

	mER := make([]float64, attributes.EndStatType)

	mER[attributes.ER] = 0.20

	// 固有天賦2:
	// 雷影剣のアバンダンスアミュレットによるER効果を旅人のERの10%分増加させる。
	// - この効果は旅人の元のERのみを参照する。
	// - アミュレット取得によるER増加はResounding Roarの他のアミュレット回収のER共有量に影響しない。
	// - TODO バフなしのER%をどう取得する？初期化時に保存？
	if c.Base.Ascension >= 4 {
		mER[attributes.ER] += c.NonExtraStat(attributes.ER) * .1
	}

	// 固定エネルギーを適用
	buffEnergy := skillRegen[c.Talents.Skill] * float64(c.abundanceAmulets)

	// 4凸 - 雷影剣で生成されたアバンダンスアミュレットを獲得した時、キャラクターのエネルギーが
	//   35%未満の場合、アミュレットのエネルギー回復量が100%増加。
	buffEnergy = c.c4(buffEnergy)

	collector.AddEnergy("abundance-amulet", buffEnergy)

	// 固有天賦1:
	// パーティの他の近くのキャラクターが雷影剣で生成されたアバンダンスアミュレットを獲得すると、
	// 雷影剣のCTが1.5秒短縮される。
	if c.Base.Ascension >= 1 && collector.Index != c.Index {
		c.ReduceActionCooldown(action.ActionSkill, 90*c.abundanceAmulets)
	}

	// ER modを適用
	collector.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("abundance-amulet", 360),
		AffectedStat: attributes.ER,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return mER, true
		},
	})

	// アミュレットをリセット
	c.abundanceAmulets = 0

	return true
}
