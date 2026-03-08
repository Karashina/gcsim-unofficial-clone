package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/varka"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/venti"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// TestApplyAttackModsLunarWhitelistCRCD verifies Lunar tags only receive CR/CD from AttackMods
func TestApplyAttackModsLunarWhitelistCRCD(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.TestCharDoNotUse)
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	ch := c.Player.Chars()[idx]

	// Add an AttackMod that provides ATK%, DmgP, CR, CD
	testMod := make([]float64, attributes.EndStatType)
	testMod[attributes.ATKP] = 0.50 // should NOT apply to Lunar
	testMod[attributes.DmgP] = 0.30 // should NOT apply to Lunar
	testMod[attributes.CR] = 0.15   // should apply to Lunar
	testMod[attributes.CD] = 0.25   // should apply to Lunar
	ch.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("test-mod", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return testMod, true
		},
	})

	// Test Lunar-Charged damage tag
	lunarTags := []attacks.AttackTag{
		attacks.AttackTagLCDamage,
		attacks.AttackTagLBDamage,
		attacks.AttackTagLCrsDamage,
	}

	for _, tag := range lunarTags {
		atk := &combat.AttackEvent{
			Info: combat.AttackInfo{
				ActorIndex: idx,
				AttackTag:  tag,
			},
		}
		atk.Snapshot.Stats = [attributes.EndStatType]float64{}
		ch.ApplyAttackMods(atk, nil)

		// CR and CD should be applied
		if atk.Snapshot.Stats[attributes.CR] < 0.14 {
			t.Fatalf("tag %v: CR should be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.CR])
		}
		if atk.Snapshot.Stats[attributes.CD] < 0.24 {
			t.Fatalf("tag %v: CD should be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.CD])
		}

		// ATK% and DmgP should NOT be applied
		if atk.Snapshot.Stats[attributes.ATKP] > 0.001 {
			t.Fatalf("tag %v: ATKP should NOT be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.ATKP])
		}
		if atk.Snapshot.Stats[attributes.DmgP] > 0.001 {
			t.Fatalf("tag %v: DmgP should NOT be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.DmgP])
		}
	}
}

// TestApplyAttackModsNormalTagGetsAllStats verifies normal tags receive all stat mods
func TestApplyAttackModsNormalTagGetsAllStats(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.TestCharDoNotUse)
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	ch := c.Player.Chars()[idx]

	testMod := make([]float64, attributes.EndStatType)
	testMod[attributes.ATKP] = 0.50
	testMod[attributes.DmgP] = 0.30
	testMod[attributes.CR] = 0.15
	testMod[attributes.CD] = 0.25
	ch.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("test-mod", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return testMod, true
		},
	})

	// Normal attack tag should get ALL stats
	normalTags := []attacks.AttackTag{
		attacks.AttackTagNormal,
		attacks.AttackTagExtra,
		attacks.AttackTagElementalArt,
		attacks.AttackTagElementalBurst,
	}

	for _, tag := range normalTags {
		atk := &combat.AttackEvent{
			Info: combat.AttackInfo{
				ActorIndex: idx,
				AttackTag:  tag,
			},
		}
		atk.Snapshot.Stats = [attributes.EndStatType]float64{}
		ch.ApplyAttackMods(atk, nil)

		// All stats should be applied
		if atk.Snapshot.Stats[attributes.ATKP] < 0.49 {
			t.Fatalf("tag %v: ATKP should be applied to normal attacks, got %v", tag, atk.Snapshot.Stats[attributes.ATKP])
		}
		if atk.Snapshot.Stats[attributes.DmgP] < 0.29 {
			t.Fatalf("tag %v: DmgP should be applied to normal attacks, got %v", tag, atk.Snapshot.Stats[attributes.DmgP])
		}
		if atk.Snapshot.Stats[attributes.CR] < 0.14 {
			t.Fatalf("tag %v: CR should be applied to normal attacks, got %v", tag, atk.Snapshot.Stats[attributes.CR])
		}
		if atk.Snapshot.Stats[attributes.CD] < 0.24 {
			t.Fatalf("tag %v: CD should be applied to normal attacks, got %v", tag, atk.Snapshot.Stats[attributes.CD])
		}
	}
}

// TestApplyAttackModsReactionTagsSkipped verifies non-Lunar reaction tags get no mods
func TestApplyAttackModsReactionTagsSkipped(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.TestCharDoNotUse)
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	ch := c.Player.Chars()[idx]

	testMod := make([]float64, attributes.EndStatType)
	testMod[attributes.CR] = 0.15
	testMod[attributes.CD] = 0.25
	ch.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("test-mod", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return testMod, true
		},
	})

	// Reaction tags (non-Lunar) should be skipped entirely
	reactionTags := []attacks.AttackTag{
		attacks.AttackTagOverloadDamage,
		attacks.AttackTagSuperconductDamage,
		attacks.AttackTagSwirlPyro,
		attacks.AttackTagBloom,
	}

	for _, tag := range reactionTags {
		atk := &combat.AttackEvent{
			Info: combat.AttackInfo{
				ActorIndex: idx,
				AttackTag:  tag,
			},
		}
		atk.Snapshot.Stats = [attributes.EndStatType]float64{}
		result := ch.ApplyAttackMods(atk, nil)

		// ApplyAttackMods should return nil for reaction tags
		if result != nil {
			t.Fatalf("tag %v: ApplyAttackMods should return nil for reaction tags", tag)
		}

		// No stats should be applied
		if atk.Snapshot.Stats[attributes.CR] > 0.001 {
			t.Fatalf("tag %v: CR should NOT be applied to reaction tags, got %v", tag, atk.Snapshot.Stats[attributes.CR])
		}
	}
}

// TestHexereiPartyDetection verifies Hexerei party detection (2+ Hexerei chars)
func TestHexereiPartyDetection(t *testing.T) {
	// Test with 2 Hexerei chars (Varka + Venti)
	c1, _ := makeCore(1)
	profVarka := defProfile(keys.Varka)
	profVarka.Base.Ascension = 6
	_, err := c1.AddChar(profVarka)
	if err != nil {
		t.Fatalf("error adding Varka: %v", err)
	}
	profVenti := defProfile(keys.Venti)
	profVenti.Base.Ascension = 6
	_, err = c1.AddChar(profVenti)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
	}
	for i := 0; i < 2; i++ {
		_, err := c1.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c1.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	// Both should report hexerei = true
	for i, name := range []string{"Varka", "Venti"} {
		result, err := c1.Player.Chars()[i].Condition([]string{"hexerei"})
		if err != nil {
			t.Fatalf("%s: hexerei condition error: %v", name, err)
		}
		if isHex, ok := result.(bool); !ok || !isHex {
			t.Fatalf("%s should have hexerei=true", name)
		}
	}

	// Test with 1 Hexerei char (Varka only, nohex on Venti)
	c2, _ := makeCore(1)
	profVarka2 := defProfile(keys.Varka)
	profVarka2.Base.Ascension = 6
	_, err = c2.AddChar(profVarka2)
	if err != nil {
		t.Fatalf("error adding Varka: %v", err)
	}
	profVenti2 := defProfile(keys.Venti)
	profVenti2.Base.Ascension = 6
	profVenti2.Params["nohex"] = 1
	_, err = c2.AddChar(profVenti2)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
	}
	for i := 0; i < 2; i++ {
		_, err := c2.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c2.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	// Varka has hexerei but without 2+ chars, hasHexBonus should be false
	// (Hexerei is a per-character flag, hasHexBonus is a team-wide check)
	varkaResult, _ := c2.Player.Chars()[0].Condition([]string{"hexerei"})
	if isHex, ok := varkaResult.(bool); !ok || !isHex {
		t.Fatal("Varka should still have hexerei=true even without bonus")
	}

	// Venti should have hexerei=false with nohex=1
	ventiResult, _ := c2.Player.Chars()[1].Condition([]string{"hexerei"})
	if isHex, ok := ventiResult.(bool); ok && isHex {
		t.Fatal("Venti should have hexerei=false with nohex=1")
	}
}
