package athame

import (
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/testhelper"
)

func TestAthameBurstTriggers(t *testing.T) {
	core.RegisterCharFunc(keys.TestCharDoNotUse, testhelper.NewChar)

	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Geo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.AthameArtis, Refine: 1, Level: 1}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	before := char.Stat(attributes.ATKP)

	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagElementalBurst}}
	c.Events.Emit(event.OnEnemyHit, nil, atk)

	after := char.Stat(attributes.ATKP)
	if after <= before {
		t.Fatalf("expected ATKP to increase after burst hit, before=%v after=%v", before, after)
	}
}
