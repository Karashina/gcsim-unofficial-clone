package heizou

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var burstFrames []int

func init() {
	burstFrames = frames.InitAbilSlice(72)
	burstFrames[action.ActionAttack] = 71
	burstFrames[action.ActionSkill] = 71
	burstFrames[action.ActionJump] = 70
	burstFrames[action.ActionSwap] = 69
}

const burstHitmark = 34

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.burstTaggedCount = 0
	burstCB := func(a combat.AttackCB) {
		// 敵かチェック
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		// 最大4体までタグ付け
		if c.burstTaggedCount == 4 {
			return
		}
		// 元素を確認して攻撃をキューに追加
		c.burstTaggedCount++
		if c.Base.Cons >= 4 {
			c.c4(c.burstTaggedCount)
		}
		c.irisDmg(a.Target)
	}
	auraCheck := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Windmuster Iris (Aura check)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Physical,
		Durability: 0,
		Mult:       0,
		NoImpulse:  true,
	}
	// 敵のみに命中するべき
	ap := combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6)
	ap.SkipTargets[targets.TargettableGadget] = true
	c.Core.QueueAttack(auraCheck, ap, burstHitmark, burstHitmark, burstCB)

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Fudou Style Vacuum Slugger",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	//TODO: 平蔵の元素爆発はスナップショットするか？
	//TODO: 平蔵の元素爆発の弾道時間パラメータ
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6),
		burstHitmark,
		burstHitmark,
	)

	//TODO: CDの遅延の有無、エネルギー消費フレームを確認
	c.SetCD(action.ActionBurst, 12*60)
	c.ConsumeEnergy(3)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

// 徑座开闘が水/炎/氷/雷元素の影響を受けた敵に命中すると、
// 敵に「風の患い」が付与される。
// 「風の患い」は少し後に爆発して消滅し、
// 対応する元素の範囲ダメージを与える。
func (c *char) irisDmg(t combat.Target) {
	x, ok := t.(combat.TargetWithAura)
	if !ok {
		//TODO: これで正しいか確認が必要。何もしないで良いのか？
		return
	}
	//TODO: 元素爆発の風の患いはスナップショットするか
	aiAbs := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Windmuster Iris",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.NoElement,
		Durability: 25,
		Mult:       burstIris[c.TalentLvlBurst()],
	}
	auraPriority := []attributes.Element{attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo}
	for _, ele := range auraPriority {
		if x.AuraContains(ele) {
			aiAbs.Element = ele
			break
		}
	}
	if aiAbs.Element == attributes.NoElement {
		c.Core.Log.NewEvent(
			"No valid aura detected, omiting iris",
			glog.LogCharacterEvent,
			c.Index,
		).Write("target", t.Key())
		return
	}

	c.Core.QueueAttack(aiAbs, combat.NewCircleHitOnTarget(t, nil, 2.5), 0, 40) // この値が間違っていたらKoliのせい
}
