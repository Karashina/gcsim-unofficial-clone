package nahida

import (
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a1BuffKey = "nahida-a1"
const a4BuffKey = "nahida-a4"

// 幻心を解き放つとき、マヤの祠殿は以下の効果を得る:
// フィールド内のアクティブキャラクターの元素熔研が、チーム内で最も元素熔研が高いメンバーの25%分増加する。
// この方法で得られる元素熔研の最大値は250。
func (c *char) calcA1Buff() {
	if c.Base.Ascension < 1 {
		return
	}
	var maxEM float64
	team := c.Core.Player.Chars()
	for _, char := range team {
		em := char.NonExtraStat(attributes.EM)
		if em > maxEM {
			maxEM = em
		}
	}
	maxEM = 0.25 * maxEM

	if maxEM > 250 {
		maxEM = 250
	}

	c.a1Buff[attributes.EM] = maxEM
}

func (c *char) applyA1(dur int) {
	if c.Base.Ascension < 1 {
		return
	}
	for i, char := range c.Core.Player.Chars() {
		idx := i
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(a1BuffKey, dur),
			AffectedStat: attributes.EM,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return c.a1Buff, c.Core.Player.Active() == idx
			},
		})
	}
}

// ナヒーダの元素熔研が200を超えるごとに、三浄浄化に0.1%のダメージバフと
// 0.03%の会心率が付与される。
// 三浄浄化に付与できる最大のダメージバフは80%、会心率は24%。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a4BuffKey, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			if !strings.HasPrefix(atk.Info.Abil, "Tri-Karma") {
				return nil, false
			}
			return c.a4Buff, true
		},
	})
}

func (c *char) a4Tick() {
	if c.Base.Ascension < 4 {
		return
	}
	em := c.Stat(attributes.EM)
	var dmgBuff, crBuff float64
	if em > 200 {
		em -= 200
		dmgBuff = em * 0.001
		if dmgBuff > 0.8 {
			dmgBuff = 0.8
		}
		crBuff = em * 0.0003
		if crBuff > .24 {
			crBuff = .24
		}
	}
	c.a4Buff[attributes.DmgP] = dmgBuff
	c.a4Buff[attributes.CR] = crBuff

	c.Core.Tasks.Add(c.a4Tick, 30)
}
