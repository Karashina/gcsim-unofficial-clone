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

// TestApplyAttackModsLunarWhitelistCRCD はLunarタグがAttackModsからCR/CDのみ受け取ることを検証する
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

	// ATK%, DmgP, CR, CDを提供するAttackModを追加
	testMod := make([]float64, attributes.EndStatType)
	testMod[attributes.ATKP] = 0.50 // Lunarには適用されないべき
	testMod[attributes.DmgP] = 0.30 // Lunarには適用されないべき
	testMod[attributes.CR] = 0.15   // Lunarに適用されるべき
	testMod[attributes.CD] = 0.25   // Lunarに適用されるべき
	ch.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("test-mod", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return testMod, true
		},
	})

	// Lunar-Chargedダメージタグをテスト
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

		// CRとCDが適用されているべき
		if atk.Snapshot.Stats[attributes.CR] < 0.14 {
			t.Fatalf("tag %v: CR should be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.CR])
		}
		if atk.Snapshot.Stats[attributes.CD] < 0.24 {
			t.Fatalf("tag %v: CD should be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.CD])
		}

		// ATK%とDmgPは適用されないべき
		if atk.Snapshot.Stats[attributes.ATKP] > 0.001 {
			t.Fatalf("tag %v: ATKP should NOT be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.ATKP])
		}
		if atk.Snapshot.Stats[attributes.DmgP] > 0.001 {
			t.Fatalf("tag %v: DmgP should NOT be applied to Lunar reactions, got %v", tag, atk.Snapshot.Stats[attributes.DmgP])
		}
	}
}

// TestApplyAttackModsNormalTagGetsAllStats は通常タグが全ステータス補正を受け取ることを検証する
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

	// 通常攻撃タグは全ステータスを受け取るべき
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

		// 全ステータスが適用されているべき
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

// TestApplyAttackModsReactionTagsSkipped は非Lunar元素反応タグがmodを受け取らないことを検証する
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

	// 元素反応タグ（非Lunar）は完全にスキップされるべき
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

		// ApplyAttackModsは元素反応タグに対してnilを返すべき
		if result != nil {
			t.Fatalf("tag %v: ApplyAttackMods should return nil for reaction tags", tag)
		}

		// ステータスが適用されないべき
		if atk.Snapshot.Stats[attributes.CR] > 0.001 {
			t.Fatalf("tag %v: CR should NOT be applied to reaction tags, got %v", tag, atk.Snapshot.Stats[attributes.CR])
		}
	}
}

// TestHexereiPartyDetection はHexereiパーティ検出（2人以上のHexereiキャラ）を検証する
func TestHexereiPartyDetection(t *testing.T) {
	// Hexereiキャラ2人でテスト（Varka + Venti）
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

	// 両方がhexerei = trueを報告するべき
	for i, name := range []string{"Varka", "Venti"} {
		result, err := c1.Player.Chars()[i].Condition([]string{"hexerei"})
		if err != nil {
			t.Fatalf("%s: hexerei condition error: %v", name, err)
		}
		if isHex, ok := result.(bool); !ok || !isHex {
			t.Fatalf("%s should have hexerei=true", name)
		}
	}

	// Hexereiキャラ1人でテスト（Varkaのみ、Ventiはnohex）
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

	// Varkaはhexereiを持つが2人以上いないため、hasHexBonusはfalseであるべき
	// （Hexereiはキャラ単位のフラグ、hasHexBonusはチーム全体の判定）
	varkaResult, _ := c2.Player.Chars()[0].Condition([]string{"hexerei"})
	if isHex, ok := varkaResult.(bool); !ok || !isHex {
		t.Fatal("Varka should still have hexerei=true even without bonus")
	}

	// Ventiはnohex=1でhexerei=falseであるべき
	ventiResult, _ := c2.Player.Chars()[1].Condition([]string{"hexerei"})
	if isHex, ok := ventiResult.(bool); ok && isHex {
		t.Fatal("Venti should have hexerei=false with nohex=1")
	}
}
