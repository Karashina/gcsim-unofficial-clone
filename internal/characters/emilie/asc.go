package emilie

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) a0() {
	// mitigates true dmg
	c.Core.Events.Subscribe(event.OnPlayerPreHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		// only mitigate burning self damage
		if di.Abil != "burning (self damage)" {
			return false
		}
		// no need to mitigate if 0 dmg
		if di.Amount <= 0 {
			return false
		}
		beforeAmount := di.Amount
		// calc mitigation
		mitigation := di.Amount * 0.22727
		// modify hp drain
		di.Amount = max(di.Amount-mitigation, 0)
		// log mitigation
		c.Core.Log.NewEvent("emilie mitigating dmg", glog.LogCharacterEvent, c.Index).
			Write("hurt_before", beforeAmount).
			Write("mitigation", mitigation).
			Write("hurt", di.Amount)
		return false
	}, "emilie-A0-Mitigation")
}

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lingering Fragrance (A1)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupEmilieSkill,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       6,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2.5),
		0,
		0,
	)
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	buff := min(0.36, c.TotalAtk()/1000*0.15)

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = buff
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("emilie-a4", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			r, ok := t.(*enemy.Enemy)
			if !ok {
				return nil, false
			}
			if !r.IsBurning() {
				return nil, false
			}
			return m, true
		},
	})
}
