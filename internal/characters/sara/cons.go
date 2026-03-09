package sara

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c1ICDKey = "sara-c1-icd"

// 1凸CD短縮の実装。遅延（敵に命中した時）後に効果を発動する
// 元素スキルと元素爆発で発動する
func (c *char) c1() {
	if c.StatusIsActive(c1ICDKey) {
		return
	}
	c.AddStatus(c1ICDKey, 180, true)
	c.ReduceActionCooldown(action.ActionSkill, 60)
	c.Core.Log.NewEvent("c1 reducing skill cooldown", glog.LogCharacterEvent, c.Index).
		Write("new_cooldown", c.Cooldown(action.ActionSkill))
}

// 天狗呉雷により攻撃力が上昇したキャラクターの雷元素ダメージの会心ダメージが60%上昇する。
func (c *char) c6(char *character.CharWrapper) {
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("sara-c6", 360),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.Element != attributes.Electro {
				return nil, false
			}
			return c.c6buff, true
		},
	})
}
