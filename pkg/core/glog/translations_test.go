package glog

import (
	"testing"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		english  string
		japanese string
	}{
		{"chongyun adding infusion", "重雲元素付与追加"},
		{"barbara heal and wet ticking", "バーバラの回復と水元素付与ティック"},
		{"oz activated", "オズが発動"},
		{"unknown message", "unknown message"}, // Should return original for unknown messages
	}

	for _, test := range tests {
		result := Translate(test.english)
		if result != test.japanese {
			t.Errorf("Translate(%q) = %q, want %q", test.english, result, test.japanese)
		}
	}
}

func TestLoggerWithTranslation(t *testing.T) {
	frame := 0
	logger := New(&frame, 100)
	
	event := logger.NewEvent("chongyun adding infusion", LogCharacterEvent, 0)
	
	// Check that the event was created with translated message
	logEvent, ok := event.(*LogEvent)
	if !ok {
		t.Fatal("Expected LogEvent type")
	}
	
	expected := "重雲元素付与追加"
	if logEvent.Msg != expected {
		t.Errorf("Expected translated message %q, got %q", expected, logEvent.Msg)
	}
}