package bennett

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/avatar"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstStartFrame   = 34
	burstBuffDuration = 126
	burstKey          = "bennettburst"
	burstFieldKey     = "bennett-field"
)

func init() {
	burstFrames = frames.InitAbilSlice(53)
	burstFrames[action.ActionDash] = 49
	burstFrames[action.ActionJump] = 50
	burstFrames[action.ActionSwap] = 51
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// フィールド効果タイマーを追加
	// 設置物のためヒットラグなし
	c.Core.Status.Add(burstKey, 720+burstStartFrame)
	// バフ用フック。発動直後に有効

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Fantastic Voyage",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	const radius = 6.0
	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, radius)
	c.Core.QueueAttack(ai, burstArea, 37, 37)

	// t=0sからt=12sまで13回のティックを追加
	// バフはヒット直前（t=0s）からティック開始
	// https://discord.com/channels/845087716541595668/869210750596554772/936507730779308032
	stats, _ := c.Stats()

	// 最初のティックは攻撃力バフのみで回復しない
	c.Core.Tasks.Add(func() {
		if c.Core.Combat.Player().IsWithinArea(burstArea) {
			c.applyBennettField(stats, true)()
		}
	}, burstStartFrame)
	// その他のティックは回復する
	for i := 60; i <= 12*60; i += 60 {
		c.Core.Tasks.Add(func() {
			if c.Core.Combat.Player().IsWithinArea(burstArea) {
				c.applyBennettField(stats, false)()
			}
		}, i+burstStartFrame)
	}

	c.ConsumeEnergy(36)
	c.SetCDWithDelay(action.ActionBurst, 900, burstStartFrame)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) applyBennettField(stats [attributes.EndStatType]float64, firstTick bool) func() {
	hpplus := stats[attributes.Heal]
	heal := bursthp[c.TalentLvlBurst()] + bursthpp[c.TalentLvlBurst()]*c.MaxHP()
	pc := burstatk[c.TalentLvlBurst()]
	if c.Base.Cons >= 1 {
		pc += 0.2
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATK] = pc * c.Stat(attributes.BaseATK)
	if c.Base.Cons >= 6 {
		m[attributes.PyroP] = 0.15
	}

	return func() {
		c.Core.Log.NewEvent("bennett field ticking", glog.LogCharacterEvent, -1)

		// 自己元素付与
		p, ok := c.Core.Combat.Player().(*avatar.Player)
		if !ok {
			panic("target 0 should be Player but is not!!")
		}
		p.ApplySelfInfusion(attributes.Pyro, 25, burstBuffDuration)

		active := c.Core.Player.ActiveChar()
		// 最初のティックでなくHP70%未満なら回復
		if !firstTick && active.CurrentHPRatio() < 0.7 {
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  active.Index,
				Message: "Inspiration Field",
				Src:     heal,
				Bonus:   hpplus,
			})
		}

		// HP70%超の場合、攻撃力を追加
		threshold := .7
		if c.Base.Cons >= 1 {
			threshold = 0
		}
		// 攻撃力バフを有効化
		if active.CurrentHPRatio() > threshold {
			// 武器元素付与を追加
			if c.Base.Cons >= 6 {
				switch active.Weapon.Class {
				case info.WeaponClassClaymore:
					fallthrough
				case info.WeaponClassSpear:
					fallthrough
				case info.WeaponClassSword:
					c.Core.Player.AddWeaponInfuse(
						active.Index,
						"bennett-fire-weapon",
						attributes.Pyro,
						burstBuffDuration,
						true,
						attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
					)
				}
				c.Core.Events.Emit(event.OnInfusion, active.Index, attributes.Pyro, burstBuffDuration)
			}

			active.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(burstFieldKey, burstBuffDuration),
				AffectedStat: attributes.NoStat,
				Extra:        true,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})

			c.Core.Log.NewEvent("bennett field - adding attack", glog.LogCharacterEvent, c.Index).
				Write("threshold", threshold)
		}
	}
}
