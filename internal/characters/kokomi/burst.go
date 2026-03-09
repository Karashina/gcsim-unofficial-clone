package kokomi

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstHitmark = 49
	burstKey     = "kokomiburst"
)

func init() {
	burstFrames = frames.InitAbilSlice(78) // Q -> D/J
	burstFrames[action.ActionAttack] = 77  // Q -> N1
	burstFrames[action.ActionCharge] = 77  // Q -> CA
	burstFrames[action.ActionSkill] = 77   // Q -> E
	burstFrames[action.ActionWalk] = 77    // Q -> W
	burstFrames[action.ActionSwap] = 76    // Q -> Swap
}

// 元素爆発 - この関数は初回ダメージとステータス設定のみ処理
// ダメージボーナスの修正はステータスに基づき別関数で処理
// 海祇の力が降臨し、周囲の敵に水元素ダメージを与えた後、心海に珊瑚宮の流水から作られた「儀来羽衣」を纏わせる。
// 儀来羽衣:
// - 珊瑚宮心海の通常攻撃・重撃・化海月のダメージがHP上限に基づき増加。
// - 通常攻撃・重撃が敵に命中すると、付近の全パーティメンバーのHPを回復（HP上限に基づく）。
// - 中断耐性が増加し、水面を歩行可能になる。
// これらの効果は珊瑚宮心海がフィールドを離れると解除される。
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// TODO: スナップショットのタイミングは不明。現時点ではダイナミックと仮定
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Nereid's Ascension",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 50,
		Mult:       0,
	}
	ai.FlatDmg = burstDmg[c.TalentLvlBurst()] * c.MaxHP()

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), burstHitmark, burstHitmark)

	c.Core.Status.Add(burstKey, 10*60)

	// クラゲの固定ダメージを更新
	c.skillFlatDmg = c.burstDmgBonus(attacks.AttackTagElementalArt)

	if c.Base.Ascension >= 1 {
		c.Core.Tasks.Add(c.a1, 46)
	}

	// 4凸の攻撃速度バフ
	if c.Base.Cons >= 4 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.AtkSpd] = 0.1
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("kokomi-c4", 10*60),
			AffectedStat: attributes.AtkSpd,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	// 先行粒子受取不可
	c.ConsumeEnergy(57)
	c.SetCDWithDelay(action.ActionBurst, 18*60, 46)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

// 元素爆発ダメージボーナスが適用されるか判定するヘルパー関数
func (c *char) burstDmgBonus(a attacks.AttackTag) float64 {
	if c.Core.Status.Duration("kokomiburst") == 0 {
		return 0
	}
	switch a {
	case attacks.AttackTagNormal:
		return burstBonusNormal[c.TalentLvlBurst()] * c.MaxHP()
	case attacks.AttackTagExtra:
		return burstBonusCharge[c.TalentLvlBurst()] * c.MaxHP()
	case attacks.AttackTagElementalArt:
		return burstBonusSkill[c.TalentLvlBurst()] * c.MaxHP()
	default:
		return 0
	}
}

// - 元素爆発の回復、2凸・6凸の処理を実装
//
// 通常攻撃・重撃が敵に命中すると、
// 心海は付近の全パーティメンバーのHPを回復する。
// 回復量はHP上限に基づく。
func (c *char) makeBurstHealCB() combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.Core.Status.Duration("kokomiburst") == 0 {
			return
		}
		if done {
			return
		}
		done = true

		heal := burstHealPct[c.TalentLvlBurst()]*c.MaxHP() + burstHealFlat[c.TalentLvlBurst()]
		for _, char := range c.Core.Player.Chars() {
			src := heal

			// 2凸: HP50%以下のキャラクターに対する回復ボーナス
			// 海人の羽衣の通常攻撃・重撃: 心海のHP上限の0.6%分。
			if c.Base.Cons >= 2 && char.CurrentHPRatio() <= 0.5 {
				bonus := 0.006 * c.MaxHP()
				src += bonus
				c.Core.Log.NewEvent("kokomi c2 proc'd", glog.LogCharacterEvent, char.Index).
					Write("bonus", bonus)
			}
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  char.Index,
				Message: "Ceremonial Garment",
				Src:     src,
				Bonus:   c.Stat(attributes.Heal),
			})
		}

		if c.Base.Cons >= 6 {
			c.c6()
		}
	}
}

// 心海がフィールドを離れると元素爆発を解除する
func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		// 海月の固定ダメージを更新（爆発がアクティブかどうかに関わらず）
		if prev == c.Index {
			c.swapEarlyF = c.Core.F
			c.skillFlatDmg = c.burstDmgBonus(attacks.AttackTagElementalArt)
		}
		c.Core.Status.Delete("kokomiburst")
		return false
	}, "kokomi-exit")
}
