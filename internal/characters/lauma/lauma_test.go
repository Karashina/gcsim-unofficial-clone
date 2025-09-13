package lauma

import (
	"testing"
)

func TestLaumaScalingValues(t *testing.T) {
	// Test that scaling arrays are properly populated
	if len(attack_1) != 15 {
		t.Errorf("Expected attack_1 to have 15 values, got %d", len(attack_1))
	}
	
	if len(attack_2) != 15 {
		t.Errorf("Expected attack_2 to have 15 values, got %d", len(attack_2))
	}
	
	if len(attack_3) != 15 {
		t.Errorf("Expected attack_3 to have 15 values, got %d", len(attack_3))
	}
	
	if len(skillPress) != 15 {
		t.Errorf("Expected skillPress to have 15 values, got %d", len(skillPress))
	}
	
	if len(skillHold1) != 15 {
		t.Errorf("Expected skillHold1 to have 15 values, got %d", len(skillHold1))
	}
	
	if len(skillHold2) != 15 {
		t.Errorf("Expected skillHold2 to have 15 values, got %d", len(skillHold2))
	}
	
	if len(skillDotATK) != 15 {
		t.Errorf("Expected skillDotATK to have 15 values, got %d", len(skillDotATK))
	}
	
	if len(skillDotEM) != 15 {
		t.Errorf("Expected skillDotEM to have 15 values, got %d", len(skillDotEM))
	}
	
	if len(burstBuffBloom) != 15 {
		t.Errorf("Expected burstBuffBloom to have 15 values, got %d", len(burstBuffBloom))
	}
	
	if len(burstBuffLBloom) != 15 {
		t.Errorf("Expected burstBuffLBloom to have 15 values, got %d", len(burstBuffLBloom))
	}
	
	// Test that values are reasonable (not zero or negative)
	for i, val := range attack_1 {
		if val <= 0 {
			t.Errorf("attack_1[%d] should be positive, got %f", i, val)
		}
	}
	
	for i, val := range skillPress {
		if val <= 0 {
			t.Errorf("skillPress[%d] should be positive, got %f", i, val)
		}
	}
	
	// Test that values are generally increasing (typical for talent scaling)
	if attack_1[0] >= attack_1[14] {
		t.Errorf("Expected attack_1 scaling to increase from talent level 1 to 15")
	}
	
	if skillPress[0] >= skillPress[14] {
		t.Errorf("Expected skillPress scaling to increase from talent level 1 to 15")
	}
}