package hydro

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

const c4ICDKey = "travelerhydro-c4-icd"

// 水紋の剣使用時、旅人の最大HPの10%のダメージを吸収できる水紋の盾を生成し、
// 水元素ダメージを250%の効率で吸収する。元素スキルの使用が終了するまで持続する。
// 2秒ごとに、水珠が敵に命中した後、旅人が水紋の盾で保護されている場合、
// 盾のダメージ吸収量が旅人の最大HPの10%に回復する。盾がない場合は再展開される。
func (c *Traveler) c4() {
	existingShield := c.Core.Player.Shields.Get(shield.TravelerHydroC4)
	if existingShield != nil {
		// HPを更新
		shd, _ := existingShield.(*shield.Tmpl)
		shd.HP = 0.1 * c.MaxHP()
		c.Core.Log.NewEvent("update shield hp", glog.LogCharacterEvent, c.Index).
			Write("hp", shd.HP)
		return
	}

	// シールドを追加
	c.Core.Player.Shields.Add(&shield.Tmpl{
		ActorIndex: c.Index,
		Target:     c.Index,
		Src:        c.Core.F,
		ShieldType: shield.TravelerHydroC4,
		Name:       "Traveler (Hydro) C4",
		HP:         0.1 * c.MaxHP(),
		Ele:        attributes.Hydro,
		Expires:    c.Core.F + 15*60,
	})
}

func (c *Traveler) makeC4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(c4ICDKey) {
			return
		}

		c.c4()
		c.AddStatus(c4ICDKey, 2*60, true)
	}
}

func (c *Traveler) c4Remove() {
	if c.Base.Cons < 4 {
		return
	}

	existingShield := c.Core.Player.Shields.Get(shield.TravelerHydroC4)
	if existingShield == nil {
		return
	}
	shd, _ := existingShield.(*shield.Tmpl)
	shd.Expires = c.Core.F + 1
}

// 旅人が源水の雫を拾うと、他のパーティメンバーの中で残りHP割合が最も低いキャラクターの
// 最大HPの6%に基づいてHPを回復する。
func (c *Traveler) c6() {
	lowest := c.Index
	chars := c.Core.Player.Chars()
	for i, char := range chars {
		if char.CurrentHPRatio() < chars[lowest].CurrentHPRatio() {
			lowest = i
		}
	}

	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  lowest,
		Type:    info.HealTypePercent,
		Message: "Tides of Justice",
		Src:     0.06,
		Bonus:   c.Stat(attributes.Heal),
	})
}
