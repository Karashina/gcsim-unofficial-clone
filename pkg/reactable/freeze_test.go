package reactable

import (
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

func TestFreezePlusCryoHydro(t *testing.T) {
	c := testCore()
	trg := addTargetToCore(c)
	c.Init()

	trg.AttachOrRefill(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Cryo,
			Durability: 25,
		},
	})
	trg.React(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Hydro,
			Durability: 25,
		},
	})
	// Tick処理なしの場合、ここで凍結50があるはず
	if !durApproxEqual(40, trg.Durability[Frozen], 0.00001) {
		t.Errorf("frozen expected to be 40, got %v", trg.Durability[Frozen])
		t.FailNow()
	}

	trg.AttachOrRefill(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Cryo,
			Durability: 25,
		},
	})

	// ここで凍結+氷元素があるはず
	if !durApproxEqual(20, trg.Durability[Cryo], 0.00001) {
		t.Errorf("expecting 20 cryo attached, got %v", trg.Durability[Cryo])
	}
}

func TestFreezePlusAddFreeze(t *testing.T) {
	c := testCore()
	trg := addTargetToCore(c)
	c.Init()

	trg.AttachOrRefill(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Cryo,
			Durability: 25,
		},
	})
	trg.React(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Hydro,
			Durability: 25,
		},
	})
	// Tick処理なしの場合、ここで凍結50があるはず
	if !durApproxEqual(40, trg.Durability[Frozen], 0.00001) {
		t.Errorf("frozen expected to be 40, got %v", trg.Durability[Frozen])
		t.FailNow()
	}

	trg.AttachOrRefill(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Cryo,
			Durability: 50, // 付与後は40になる
		},
	})
	trg.React(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Hydro,
			Durability: 50,
		},
	})

	// ここで凍結+氷元素があるはず
	if !durApproxEqual(80, trg.Durability[Frozen], 0.00001) {
		t.Errorf("expecting 80 frozen attached, got %v", trg.Durability[Frozen])
	}
}
