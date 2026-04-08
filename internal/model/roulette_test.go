package model

import "testing"

func TestRouletteConfigValidateStandard(t *testing.T) {
	cfg := RouletteConfig{Mode: RouletteStandard}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("standard config should be valid: %v", err)
	}
}

func TestRouletteConfigValidateMinMax(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RouletteConfig
		wantErr bool
	}{
		{"valid", RouletteConfig{Mode: RouletteMinMax, Min: 1, Max: 100}, false},
		{"min >= max", RouletteConfig{Mode: RouletteMinMax, Min: 100, Max: 100}, true},
		{"negative min", RouletteConfig{Mode: RouletteMinMax, Min: -1, Max: 10}, true},
		{"range too large", RouletteConfig{Mode: RouletteMinMax, Min: 0, Max: 20000}, true},
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

func TestRouletteConfigValidateCustom(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RouletteConfig
		wantErr bool
	}{
		{"valid", RouletteConfig{Mode: RouletteCustomValues, Values: []string{"a", "b"}}, false},
		{"too few", RouletteConfig{Mode: RouletteCustomValues, Values: []string{"a"}}, true},
		{"empty value", RouletteConfig{Mode: RouletteCustomValues, Values: []string{"a", ""}}, true},
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

func TestParseRouletteMode(t *testing.T) {
	tests := []struct {
		input string
		want  RouletteMode
		err   bool
	}{
		{"", RouletteStandard, false},
		{"standard", RouletteStandard, false},
		{"minmax", RouletteMinMax, false},
		{"custom", RouletteCustomValues, false},
		{"invalid", 0, true},
	}
	for _, tt := range tests {
		m, err := ParseRouletteMode(tt.input)
		if (err != nil) != tt.err {
			t.Fatalf("ParseRouletteMode(%q) error = %v, wantErr = %v", tt.input, err, tt.err)
		}
		if err == nil && m != tt.want {
			t.Fatalf("ParseRouletteMode(%q) = %v, want %v", tt.input, m, tt.want)
		}
	}
}

func TestRouletteColors(t *testing.T) {
	if RouletteColors[0] != RouletteGreen {
		t.Fatal("0 should be green")
	}
	if RouletteColors[1] != RouletteRed {
		t.Fatal("1 should be red")
	}
	if RouletteColors[2] != RouletteBlack {
		t.Fatal("2 should be black")
	}
	// Check all 37 numbers are mapped.
	for i := 0; i <= 36; i++ {
		if _, ok := RouletteColors[i]; !ok {
			t.Fatalf("number %d not in color map", i)
		}
	}
}
