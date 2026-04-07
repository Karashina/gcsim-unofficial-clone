package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	chargeFrames        []int
	sturmChargeFrames   []int
	azureDevourFrames   []int
	chargeHitmarks      = []int{41, 42}
	sturmChargeHitmarks = []int{36, 37}
	azureDevourHitmarks = []int{39, 40, 62, 63}
)

func init() {
	chargeFrames = frames.InitAbilSlice(42)
	chargeFrames[action.ActionAttack] = 65

	sturmChargeFrames = frames.InitAbilSlice(44)
	sturmChargeFrames[action.ActionAttack] = 65

	azureDevourFrames = frames.InitAbilSlice(62)
	azureDevourFrames[action.ActionAttack] = 65
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.sturmActive {
		// CDタイマーからFWAチャージを更新
		c.updateFWACharges()
		// S&Dモード: FWAチャージがあればAzure Devourを実行
		if c.fwaCharges > 0 {
			return c.azureDevour(p)
		}
		// C6: FWAからのウィンドウが有効なら、チャージなしでAzure Devourを実行
		if c.Base.Cons >= 6 && c.StatusIsActive(c6FWAWindowKey) {
			return c.azureDevour(p)
		}
		// それ以外はS&D重撃を実行
		return c.sturmCharge(p)
	}
	return c.normalCharge(p)
}

// normalCharge は基本重撃を処理する（S&D外）
func (c *char) normalCharge(p map[string]int) (action.Info, error) {
	// 2サブヒット、両方物理
	subHits := [][]float64{charge_a, charge_b}
	for i, mult := range subHits {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Charged Attack (Hit %v)", i+1),
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             attacks.ICDTagExtraAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           120.0,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.1 * 60,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3, 4),
			chargeHitmarks[i], chargeHitmarks[i],
		)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}

// sturmCharge はS&D重撃を処理する（FWAチャージなし）
func (c *char) sturmCharge(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()

	// S&D重撃: 1打目=他, 2打目=風
	type hitInfo struct {
		mult    float64
		element attributes.Element
		icdTag  attacks.ICDTag
	}

	hits := []hitInfo{
		{sturmCA_a[lvl], c.otherElement, attacks.ICDTagVarkaCAOther},
		{sturmCA_b[lvl], attributes.Anemo, attacks.ICDTagVarkaCAAnemo},
	}

	if !c.hasOtherEle {
		hits[0].element = attributes.Anemo
	}

	for i, h := range hits {
		mult := h.mult
		// A1倍率を適用
		if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
			mult *= c.a1MultFactor
		}

		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Sturm und Drang Charged (Hit %v)", i+1),
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             h.icdTag,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           120.0,
			Element:            h.element,
			Durability:         25,
			Mult:               mult,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.1 * 60,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3, 4),
			sturmChargeHitmarks[i], sturmChargeHitmarks[i],
		)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(sturmChargeFrames),
		AnimationLength: sturmChargeFrames[action.InvalidAction],
		CanQueueAfter:   sturmChargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}

// azureDevour はFWAチャージを消費する特殊重撃を処理する
func (c *char) azureDevour(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()

	// C6: check if we should not consume charges
	consumeCharge := true
	if c.Base.Cons >= 6 {
		if c.StatusIsActive(c6FWAWindowKey) {
			// After FWA, tap charged triggers additional Azure Devour without consuming charges
			consumeCharge = false
			c.DeleteStatus(c6FWAWindowKey)
		} else if c.StatusIsActive(c6AzureWindowKey) {
			consumeCharge = false
			c.DeleteStatus(c6AzureWindowKey)
		}
	}

	if consumeCharge {
		c.fwaCharges--
	}

	// C1: Lyrical Libation効果 - S&D突入後の最初のAzure Devourは200%ダメージ
	c1Mult := 1.0
	if c.Base.Cons >= 1 && c.StatusIsActive(c1LyricalKey) {
		c1Mult = 2.0
		c.DeleteStatus(c1LyricalKey)
	}

	// Azure Devour: 4ヒット
	// 1打目: 他, 2打目: 風, 3打目: 他, 4打目: 風
	// 全てICDTagVarkaCAOtherを使用
	type hitInfo struct {
		mult    float64
		element attributes.Element
	}

	hits := []hitInfo{
		{azureOther[lvl], c.otherElement},
		{azureAnemo[lvl], attributes.Anemo},
		{azureOther[lvl], c.otherElement},
		{azureAnemo[lvl], attributes.Anemo},
	}

	if !c.hasOtherEle {
		hits[0].element = attributes.Anemo
		hits[2].element = attributes.Anemo
	}

	for i, h := range hits {
		mult := h.mult
		// A1倍率をAzure Devourに適用
		if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
			mult *= c.a1MultFactor
		}
		// C1 Lyrical Libation倍率を適用
		mult *= c1Mult

		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Azure Devour (Hit %v)", i+1),
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             attacks.ICDTagVarkaCAOther, // 4ヒット全てがこのタグを共有
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           120.0,
			Element:            h.element,
			Durability:         25,
			Mult:               mult,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   0.08 * 60,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3, 4),
			azureDevourHitmarks[i], azureDevourHitmarks[i],
		)
	}

	// C2: ATKの800%に等しい追加風元素攻撃
	if c.Base.Cons >= 2 {
		c.c2Strike(azureDevourHitmarks[3] + 4)
	}

	// C6: after Azure Devour, open window for additional FWA
	// Only set window when this was a normal Azure Devour (not a C6 chain trigger)
	if c.Base.Cons >= 6 && consumeCharge {
		c.AddStatus(c6AzureWindowKey, 60, true) // ~1s window
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(azureDevourFrames),
		AnimationLength: azureDevourFrames[action.InvalidAction],
		CanQueueAfter:   azureDevourHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}
