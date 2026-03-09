package kachina

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

const (
	c1IcdKey = "kachina-c1-icd"
	c6IcdKey = "kachina-c6-icd"
)

func (c *char) c1shard() {
	if c.Base.Cons < 1 {
		return
	}
	Area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	// 全ての結晶シャードを吸引
	for _, g := range c.Core.Combat.Gadgets() {
		cs, ok := g.(*reactable.CrystallizeShard)
		// 結晶がなければスキップ
		if !ok {
			continue
		}
		// シャードが範囲外ならスキップ
		if !cs.IsWithinArea(Area) {
			continue
		}
		// 吸引をフレームあたり0.4mと近似
		distance := cs.Pos().Distance(Area.Shape.Pos())
		travel := int(math.Ceil(distance / 0.4))
		// 結晶が生成直後で拾える前に到着するエッジケースのための特別チェック
		if c.Core.F+travel < cs.EarliestPickup {
			continue
		}
		c.Core.Tasks.Add(func() {
			cs.AddShieldKillShard()
		}, travel)
	}
}

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		// シールドをチェック
		shd := args[0].(shield.Shield)
		if shd.Type() != shield.Crystallize {
			return false
		}
		if c.StatusIsActive(c1IcdKey) {
			return false
		}
		c.AddEnergy("kachina-c1-energy", 3)
		c.AddStatus(c1IcdKey, 5*60, true)
		c.Core.Log.NewEvent("Energy gained from Crystallise", glog.LogCharacterEvent, c.Index)
		return false
	}, "shrapnel-gain")
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	if !c.StatusIsActive(skillKey) {
		c.AddStatus(skillKey, -1, true)
		c.nightsoulState.HasBlessing()
		c.nightsoulState.GeneratePoints(20)
		c.newTwirly()
	} else {
		c.nightsoulState.GeneratePoints(20)
	}
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	area := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 10)
	count := len(c.Core.Combat.EnemiesWithinArea(area, nil))
	amt := 0.08

	switch min(4, count) {
	case 1:
		amt = 0.08
	case 2:
		amt = 0.12
	case 3:
		amt = 0.16
	case 4:
		amt = 0.20
	default:
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DEFP] = amt

	for i, char := range c.Core.Player.Chars() {
		idx := i
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("kachina-c4", -1),
			AffectedStat: attributes.DEFP,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				if c.StatusIsActive(burstKey) && c.Core.Player.Active() == idx {
					return m, true
				}
				return nil, false
			},
		})
	}
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "This Time, I've Gotta Win (C6)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       2,
		UseDef:     true,
	}

	c.Core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		c.AddStatus(c6IcdKey, 5*60, true)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 3.5), 0, 1)
		return false
	}, "kachina-shield-gain")

	c.Core.Events.Subscribe(event.OnShieldBreak, func(args ...interface{}) bool {
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		c.AddStatus(c6IcdKey, 5*60, true)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 3.5), 0, 1)
		return false
	}, "kachina-shield-broken")

}
