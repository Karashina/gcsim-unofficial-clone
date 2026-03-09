package venti

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const burstStart = 94

func init() {
	burstFrames = frames.InitAbilSlice(95) // Q -> N1/CA/E/D
	burstFrames[action.ActionJump] = 94    // Q -> J
	burstFrames[action.ActionSwap] = 93    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 位置をリセット
	c.qAbsorb = attributes.NoElement
	player := c.Core.Combat.Player()
	c.qPos = geometry.CalcOffsetPoint(player.Pos(), geometry.Point{Y: 5}, player.Direction())
	c.absorbCheckLocation = combat.NewBoxHitOnTarget(c.qPos, geometry.Point{Y: -1}, 2.5, 2.5)

	// Hexereiパッシブ用に元素爆発の眼の持続時間を追跡：
	// 最後Tickは発動からフレーム (106 + 24*19) = 562で発射。眼の持続用にバッファを追加。
	c.burstEnd = c.Core.F + 570
	// この元素爆発のHex通常攻撃トリガーカウンターをリセット
	c.normalHexCount = 0

	// 8秒間の持続、0.4秒ごとにTick
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Wind's Grand Ode",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurstAnemo,
		ICDGroup:   attacks.ICDGroupVenti,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burstDot[c.TalentLvlBurst()],
	}
	ap := combat.NewCircleHitOnTarget(c.qPos, nil, 4)

	c.aiAbsorb = ai
	c.aiAbsorb.Abil = "Wind's Grand Ode (Absorbed)"
	c.aiAbsorb.Mult = burstAbsorbDot[c.TalentLvlBurst()]
	c.aiAbsorb.Element = attributes.NoElement

	// スナップショットはCDフレームと第1Tick前後？
	var snap combat.Snapshot
	c.Core.Tasks.Add(func() {
		snap = c.Snapshot(&ai)
		c.snapAbsorb = c.Snapshot(&c.aiAbsorb)
	}, 104)

	var cb combat.AttackCBFunc
	if c.Base.Cons >= 6 {
		cb = c.c6(attributes.Anemo)
	}

	// 106フレーム目から開始、24f間隔でTick。合計20回
	for i := 0; i < 20; i++ {
		c.Core.Tasks.Add(func() {
			c.Core.QueueAttackWithSnap(ai, snap, ap, 0, cb)
		}, 106+24*i)
	}
	// KQMライブラリによると、通常風4Tick後に元素変化が発生
	c.Core.Tasks.Add(c.absorbCheckQ(c.Core.F, 0, int((480-24*4)/18)), 106+24*3)

	if c.Base.Ascension >= 4 {
		c.Core.Tasks.Add(c.a4, 480+burstStart)
	}

	c.SetCDWithDelay(action.ActionBurst, 15*60, 81)
	c.ConsumeEnergy(84)

	// 4凸（Hexerei）：Venti + チームが元素爆発使用後10秒間風元素ダメージ+25%
	if c.Base.Cons >= 4 {
		c.c4New()
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) burstAbsorbedTicks() {
	var cb combat.AttackCBFunc
	if c.Base.Cons >= 6 {
		cb = c.c6(c.qAbsorb)
	}

	ap := combat.NewCircleHitOnTarget(c.qPos, nil, 6)
	// 24f間隔でTick。合計15回
	for i := 0; i < 15; i++ {
		c.Core.QueueAttackWithSnap(c.aiAbsorb, c.snapAbsorb, ap, i*24, cb)
	}
}

func (c *char) absorbCheckQ(src, count, maxcount int) func() {
	return func() {
		if count == maxcount {
			return
		}
		c.qAbsorb = c.Core.Combat.AbsorbCheck(c.Index, c.absorbCheckLocation, attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)
		if c.qAbsorb != attributes.NoElement {
			c.aiAbsorb.Element = c.qAbsorb
			switch c.qAbsorb {
			case attributes.Pyro:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstPyro
			case attributes.Hydro:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstHydro
			case attributes.Electro:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstElectro
			case attributes.Cryo:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstCryo
			}
			// ここでダメージTickを発動
			c.burstAbsorbedTicks()
			return
		}
		// それ以外はキューに追加
		c.Core.Tasks.Add(c.absorbCheckQ(src, count+1, maxcount), 18)
	}
}
