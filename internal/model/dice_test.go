package model

import "testing"

func TestDiceConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     DiceConfig
		wantErr bool
	}{
		{"valid d6", DiceConfig{Count: 2, Sides: 6}, false},
		{"valid d20", DiceConfig{Count: 1, Sides: 20}, false},
		{"count too low", DiceConfig{Count: 0, Sides: 6}, true},
		{"count too high", DiceConfig{Count: 11, Sides: 6}, true},
		{"invalid sides", DiceConfig{Count: 1, Sides: 7}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiceConfigDefaults(t *testing.T) {
	cfg := DiceConfig{}
	cfg.ApplyDefaults()
	if cfg.Count != 1 {
		t.Fatalf("default count = %d, want 1", cfg.Count)
	}
	if cfg.Sides != 6 {
		t.Fatalf("default sides = %d, want 6", cfg.Sides)
	}
}
