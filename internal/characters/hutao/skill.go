package hutao

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	skillStart        = 14
	paramitaBuff      = "paramita"
	paramitaEnergyICD = "paramita-ball-icd"
	bbDebuff          = "blood-blossom"
)

func init() {
	skillFrames = frames.InitAbilSlice(52)
	skillFrames[action.ActionAttack] = 29
	skillFrames[action.ActionBurst] = 28
	skillFrames[action.ActionDash] = 37
	skillFrames[action.ActionJump] = 37
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	bonus := ppatk[c.TalentLvlSkill()] * c.MaxHP()
	maxBonus := c.Stat(attributes.BaseATK) * 4
	if bonus > maxBonus {
		bonus = maxBonus
	}
	c.ppbuff[attributes.ATK] = bonus
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(paramitaBuff, 540+skillStart),
		AffectedStat: attributes.ATK,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return c.ppbuff, true
		},
	})
	//TODO: a1 は paramita 終了時に適用されるが、"PP延長"のチェックはしていない
	c.applyA1 = true
	c.QueueCharTask(c.a1, 540+skillStart)

	// HPを一部削除
	c.Core.Player.Drain(info.DrainInfo{
		ActorIndex: c.Index,
		Abil:       "Paramita Papilio",
		Amount:     0.30 * c.CurrentHP(),
	})

	// 0ダメージ攻撃をトリガー。凍結を解除するために重要
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Paramita (0 dmg)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Physical,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3), skillStart, skillStart)

	c.SetCDWithDelay(action.ActionSkill, 960, 14)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionBurst], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !c.StatModIsActive(paramitaBuff) {
		return
	}
	if c.StatusIsActive(paramitaEnergyICD) {
		return
	}
	c.AddStatus(paramitaEnergyICD, 5*60, true)

	count := 2.0
	if c.Core.Rand.Float64() < 0.5 {
		count = 3
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Pyro, c.ParticleDelay) // TODO: this used to be 80
}

func (c *char) applyBB(a combat.AttackCB) {
	trg, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	if !trg.StatusIsActive(bbDebuff) {
		// ティックを開始
		trg.QueueEnemyTask(c.bbtickfunc(c.Core.F, trg), 240)
		trg.SetTag(bbDebuff, c.Core.F) // 現在のbbソースを追跡
	}

	trg.AddStatus(bbDebuff, 570, true) // 8秒 + 1.5秒持続
}

func (c *char) bbtickfunc(src int, trg *enemy.Enemy) func() {
	return func() {
		// ソースが変更された場合は何もしない
		if trg.Tags[bbDebuff] != src {
			return
		}
		if !trg.StatusIsActive(bbDebuff) {
			return
		}
		c.Core.Log.NewEvent("Blood Blossom checking for tick", glog.LogCharacterEvent, c.Index).
			Write("src", src)

		// 1回分のダメージをキューに追加
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Blood Blossom",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Pyro,
			Durability: 25,
			Mult:       bb[c.TalentLvlSkill()],
		}
		// 2命ノ星座なら固定ダメージを追加
		if c.Base.Cons >= 2 {
			ai.FlatDmg += c.MaxHP() * 0.1
		}
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(trg.Key()), 0, 0)

		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("Blood Blossom ticked", glog.LogCharacterEvent, c.Index).
				Write("next expected tick", c.Core.F+240).
				Write("dur", trg.StatusExpiry(bbDebuff)).
				Write("src", src)
		}
		// 次のインスタンスをキューに追加
		trg.QueueEnemyTask(c.bbtickfunc(src, trg), 240)
	}
}
