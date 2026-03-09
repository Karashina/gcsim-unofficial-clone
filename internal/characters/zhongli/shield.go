package zhongli

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) addJadeShield() {
	shield := shieldBase[c.TalentLvlSkill()] + shieldPer[c.TalentLvlSkill()]*c.MaxHP()

	c.Core.Player.Shields.Add(c.newShield(shield, 1200))
	c.Tags["shielded"] = 1

	// シールド獲得時に耐性バフを追加
	res := []attributes.Element{attributes.Pyro, attributes.Hydro, attributes.Cryo, attributes.Electro, attributes.Geo, attributes.Anemo, attributes.Physical, attributes.Dendro}

	// シールドはプレイヤー周囲の一定範囲内の全敵に0.3秒ごとに1秒間の耐性デバフを付与
	for i := 0; i <= 1200; i += 18 {
		c.Core.Tasks.Add(func() {
			// シールドがない場合は付与を停止
			if c.Tags["shielded"] != 1 {
				return
			}
			enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7.5), nil)
			for _, v := range res {
				key := fmt.Sprintf("zhongli-%v", v.String())
				for _, e := range enemies {
					e.AddResistMod(combat.ResistMod{
						Base:  modifier.NewBaseWithHitlag(key, 60),
						Ele:   v,
						Value: -0.2,
					})
				}
			}
		}, i)
	}
}

func (c *char) removeJadeShield() {
	c.Tags["shielded"] = 0
	c.Tags["a1"] = 0
}

func (c *char) newShield(base float64, dur int) *shd {
	n := &shd{}
	n.Tmpl = &shield.Tmpl{}
	n.Tmpl.ActorIndex = c.Index
	n.Tmpl.Target = -1
	n.Tmpl.Src = c.Core.F
	n.Tmpl.ShieldType = shield.ZhongliJadeShield
	n.Tmpl.Ele = attributes.Geo
	n.Tmpl.HP = base
	n.Tmpl.Name = "Zhongli Skill"
	n.Tmpl.Expires = c.Core.F + dur
	n.c = c
	return n
}

type shd struct {
	*shield.Tmpl
	c *char
}

func (s *shd) OnExpire() {
	s.c.removeJadeShield()
}

func (s *shd) OnDamage(dmg float64, ele attributes.Element, bonus float64) (float64, bool) {
	taken, ok := s.Tmpl.OnDamage(dmg, ele, bonus)
	// まず回復を試行
	if s.c.Base.Cons >= 6 {
		// ダメージの40%を回復に変換、ただし各キャラのHP上限の8%を超えない
		// そのため各キャラクターを個別に処理する必要がある...

		active := s.c.Core.Player.ActiveChar()
		heal := 0.4 * dmg
		maxhp := s.c.MaxHP()
		if heal > 0.08*maxhp {
			heal = 0.08 * maxhp
		}
		s.c.Core.Player.Heal(info.HealInfo{
			Caller:  s.c.Index,
			Target:  active.Index,
			Message: "Chrysos, Bounty of Dominator",
			Src:     heal,
			Bonus:   s.c.Stat(attributes.Heal),
		})
	}
	if !ok {
		s.c.removeJadeShield()
	}
	if s.c.Tags["a1"] < 5 {
		s.c.Tags["a1"]++
	}
	return taken, ok
}
