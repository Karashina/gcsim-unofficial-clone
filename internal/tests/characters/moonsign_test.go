package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/aino"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/flins"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/ineffa"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/lauma"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/nefer"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// Test moonsign party initialization for 0/1/2+ moonsignKey holders.
func TestMoonsignPartyInit(t *testing.T) {
	// 0 holders: use TestCharDoNotUse for all slots
	c0, _ := makeCore(0)
	pTest := defProfile(keys.TestCharDoNotUse)
	for i := 0; i < 4; i++ {
		_, err := c0.AddChar(pTest)
		if err != nil {
			t.Fatalf("failed to add test char: %v", err)
		}
	}
	if err := c0.Init(); err != nil {
		t.Fatalf("core init failed: %v", err)
	}
	for _, ch := range c0.Player.Chars() {
		if ch.MoonsignNascent || ch.MoonsignAscendant {
			t.Fatalf("expected no moonsign flags for test chars, got nascent=%v ascendant=%v", ch.MoonsignNascent, ch.MoonsignAscendant)
		}
	}

	// 1 holder: add one Nefer and three test chars
	c1, _ := makeCore(0)
	pNefer := defProfile(keys.Nefer)
	_, err := c1.AddChar(pNefer)
	if err != nil {
		t.Fatalf("failed to add nefer: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c1.AddChar(pTest)
		if err != nil {
			t.Fatalf("failed to add test char: %v", err)
		}
	}
	if err := c1.Init(); err != nil {
		t.Fatalf("core init failed: %v", err)
	}
	for _, ch := range c1.Player.Chars() {
		if !ch.MoonsignNascent || ch.MoonsignAscendant {
			t.Fatalf("expected nascent=true ascendant=false for 1 holder, got nascent=%v ascendant=%v", ch.MoonsignNascent, ch.MoonsignAscendant)
		}
	}

	// 2+ holders: add two moonsign chars (Nefer, Lauma) and two test chars
	c2, _ := makeCore(0)
	pLauma := defProfile(keys.Lauma)
	_, err = c2.AddChar(pNefer)
	if err != nil {
		t.Fatalf("failed to add nefer: %v", err)
	}
	_, err = c2.AddChar(pLauma)
	if err != nil {
		t.Fatalf("failed to add lauma: %v", err)
	}
	for i := 0; i < 2; i++ {
		_, err := c2.AddChar(pTest)
		if err != nil {
			t.Fatalf("failed to add test char: %v", err)
		}
	}
	if err := c2.Init(); err != nil {
		t.Fatalf("core init failed: %v", err)
	}
	for _, ch := range c2.Player.Chars() {
		if !ch.MoonsignAscendant || ch.MoonsignNascent {
			t.Fatalf("expected ascendant=true nascent=false for 2+ holders, got nascent=%v ascendant=%v", ch.MoonsignNascent, ch.MoonsignAscendant)
		}
	}
}
