package model

import "testing"

func TestSlotsConfigValidateStandard(t *testing.T) {
	cfg := SlotsConfig{Mode: SlotsStandard, Rows: 3, Cols: 5}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid standard config: %v", err)
	}
}

func TestSlotsConfigValidateStandardBad(t *testing.T) {
	tests := []struct {
		name string
		cfg  SlotsConfig
	}{
		{"rows too low", SlotsConfig{Mode: SlotsStandard, Rows: 0, Cols: 5}},
		{"rows too high", SlotsConfig{Mode: SlotsStandard, Rows: 6, Cols: 5}},
		{"cols too low", SlotsConfig{Mode: SlotsStandard, Rows: 3, Cols: 2}},
		{"cols too high", SlotsConfig{Mode: SlotsStandard, Rows: 3, Cols: 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestSlotsConfigValidateMinMax(t *testing.T) {
	good := SlotsConfig{Mode: SlotsMinMax, Cols: 3, Min: 1, Max: 100}
	if err := good.Validate(); err != nil {
		t.Fatalf("valid minmax config: %v", err)
	}

	bad := SlotsConfig{Mode: SlotsMinMax, Cols: 3, Min: 100, Max: 1}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected validation error for min >= max")
	}
}

func TestSlotsConfigValidateCustom(t *testing.T) {
	good := SlotsConfig{Mode: SlotsCustom, ReelValues: [][]string{{"a", "b"}, {"x", "y"}}}
	if err := good.Validate(); err != nil {
		t.Fatalf("valid custom config: %v", err)
	}

	bad := SlotsConfig{Mode: SlotsCustom, ReelValues: [][]string{{"a"}}}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected validation error for single-column custom")
	}
}

func TestParseSlotsMode(t *testing.T) {
	tests := []struct {
		input string
		want  SlotsMode
		err   bool
	}{
		{"", SlotsStandard, false},
		{"standard", SlotsStandard, false},
		{"minmax", SlotsMinMax, false},
		{"custom", SlotsCustom, false},
		{"bad", 0, true},
	}
	for _, tt := range tests {
		m, err := ParseSlotsMode(tt.input)
		if (err != nil) != tt.err {
			t.Fatalf("ParseSlotsMode(%q) error = %v", tt.input, err)
		}
		if err == nil && m != tt.want {
			t.Fatalf("ParseSlotsMode(%q) = %v, want %v", tt.input, m, tt.want)
		}
	}
}

func TestStandardSymbolCount(t *testing.T) {
	if len(StandardSymbols) != 9 {
		t.Fatalf("expected 9 standard symbols, got %d", len(StandardSymbols))
	}
}

func TestStandardPaylineCount(t *testing.T) {
	if len(StandardPaylines) != 20 {
		t.Fatalf("expected 20 paylines, got %d", len(StandardPaylines))
	}
	for i, pl := range StandardPaylines {
		if len(pl) != 5 {
			t.Fatalf("payline %d has %d columns, want 5", i, len(pl))
		}
	}
}
