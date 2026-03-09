package wanderer

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c6ICDKey = "wanderer-c6-icd"
)

func (c *char) c1() {
	// 第1命ノ星座: 「風の恵み」状態終了時に手動で削除が必要
	if c.Base.Cons < 1 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.AtkSpd] = 0.1
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag("wanderer-c1-atkspd", 1200),
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

func (c *char) c2() {
	// 第2命ノ星座: バフは元素爆発アニメーション全体で有効
	if c.Base.Cons < 2 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = min(float64(c.maxSkydwellerPoints-c.skydwellerPoints)*0.04, 2)
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag("wanderer-c2-burstbonus", burstFramesE[action.InvalidAction]),
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

func (c *char) makeC6Callback() func(cb combat.AttackCB) {
	if c.Base.Cons < 6 {
		return nil
	}

	done := false

	return func(a combat.AttackCB) {
		if done || !c.StatusIsActive(skillKey) || c.skydwellerPoints <= 0 {
			return
		}

		done = true

		if c.c6Count < 5 && !c.StatusIsActive(c6ICDKey) && c.skydwellerPoints < 40 {
			c.AddStatus(c6ICDKey, 12, true)
			c.c6Count++
			c.skydwellerPoints += 4

			c.Core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index,
				"wanderer c6 added 4 skydweller points",
			)
		}

		// a はコアによりコールバックのパラメータとして渡される
		trg := a.Target

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Shugen: The Curtains’ Melancholic Sway",
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagWandererC6,
			ICDGroup:   attacks.ICDGroupWandererC6,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       a.AttackEvent.Info.Mult * 0.4,
		}

		// TODO: スナップショット遅延?
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(trg, nil, 2), 8, 8,
		)
	}
}
