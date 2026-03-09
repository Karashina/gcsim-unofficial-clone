package reactable

import (
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

func TestBurning(t *testing.T) {
	c := testCore()
	trg := addTargetToCore(c)

	c.Init()

	//TODO: 燃焼のテストを作成する（現状は激化からコピペ）
	trg.AttachOrRefill(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Dendro,
			Durability: 25,
		},
	})
	trg.React(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Electro,
			Durability: 25,
		},
	})
	// 草と雷が消滅、激化20が残る
	if !durApproxEqual(20, trg.Durability[Quicken], 0.00001) {
		t.Errorf("expecting 20 cryo attached, got %v", trg.Durability[Quicken])
	}
	if trg.AuraContains(attributes.Dendro, attributes.Electro) {
		t.Error("expecting target to not contain any remaining dendro or electro aura")
	}
}
