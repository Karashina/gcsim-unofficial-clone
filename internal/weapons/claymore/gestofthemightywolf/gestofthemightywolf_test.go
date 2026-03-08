package gestofthemightywolf

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

func init() {
	core.RegisterCharFunc(keys.TestCharDoNotUse, testhelper.NewChar)
}

func TestGestStacksOnNormalHit(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// Before hit: DmgP should be 0 (no stacks)
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	if atk.Snapshot.Stats[attributes.DmgP] != 0 {
		t.Fatalf("expected DmgP=0 before first hit, got %v", atk.Snapshot.Stats[attributes.DmgP])
	}

	// Advance past ICD
	c.F = stackICD + 2

	// Normal Attack hit → 1 stack
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// Check DmgP is applied
	atk2 := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk2, nil)
	if atk2.Snapshot.Stats[attributes.DmgP] <= 0 {
		t.Fatalf("expected DmgP > 0 after NA hit (1 stack), got %v", atk2.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestStacksOnSkillHit(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// Advance past ICD
	c.F = stackICD + 2

	// Skill hit → 2 stacks
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagElementalArt}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// Check DmgP (should be 2 × 7.5% = 15% at R1)
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	expectedDmg := 0.075 * 2 // 2 stacks × 7.5% per stack at R1
	if atk.Snapshot.Stats[attributes.DmgP] < expectedDmg-0.001 || atk.Snapshot.Stats[attributes.DmgP] > expectedDmg+0.001 {
		t.Fatalf("expected DmgP ≈ %v after skill hit (2 stacks), got %v", expectedDmg, atk.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestMaxStacks(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// Apply 6 stacks worth of hits (should cap at 4)
	for i := 0; i < 6; i++ {
		c.F += stackICD + 2 // advance past ICD each time
		hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
		c.Events.Emit(event.OnEnemyHit, nil, hit)
	}

	// Check DmgP capped at 4 × 7.5% = 30%
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	maxDmg := 0.075 * 4 // 4 stacks max × 7.5%
	if atk.Snapshot.Stats[attributes.DmgP] > maxDmg+0.001 {
		t.Fatalf("DmgP should be capped at %v (4 stacks), got %v", maxDmg, atk.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestChargedAttackGives2Stacks(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// Advance past ICD
	c.F = stackICD + 2

	// Charged Attack hit → 2 stacks
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagExtra}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// Check DmgP (should be 2 × 7.5% = 15%)
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	expectedDmg := 0.075 * 2
	if atk.Snapshot.Stats[attributes.DmgP] < expectedDmg-0.001 || atk.Snapshot.Stats[attributes.DmgP] > expectedDmg+0.001 {
		t.Fatalf("expected DmgP ≈ %v after CA hit (2 stacks), got %v", expectedDmg, atk.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestBuffExpiresAfter4Seconds(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// Add a stack
	c.F = stackICD + 2
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// Verify buff is active
	if !char.StatusIsActive(buffKey) {
		t.Fatal("buff should be active after hit")
	}

	// Advance 5 seconds (300 frames) — past 4s buff duration
	c.F += 300

	// Buff should have expired
	if char.StatusIsActive(buffKey) {
		t.Fatal("buff should expire after 4 seconds")
	}
}

func TestGestNoCritDMGWithoutHexerei(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// Add stacks
	c.F = stackICD + 2
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// Without Hexerei party: CD bonus should not be applied
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	if atk.Snapshot.Stats[attributes.CD] != 0 {
		t.Fatalf("expected no CRIT DMG bonus without Hexerei party, got %v", atk.Snapshot.Stats[attributes.CD])
	}
}

func TestGestAtkSpdPassive(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
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

	// ATK SPD should be permanently +10%
	atkSpd := char.Stat(attributes.AtkSpd)
	if atkSpd < 0.09 || atkSpd > 0.11 {
		t.Fatalf("expected ATK SPD ≈ 0.10, got %v", atkSpd)
	}
}
