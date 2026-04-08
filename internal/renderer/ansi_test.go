package renderer

import (
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestRouletteFrameCount(t *testing.T) {
	cfg := model.RouletteConfig{Mode: model.RouletteStandard}
	result := model.RouletteResult{Value: "17", Number: 17, Color: model.RouletteRed, Mode: model.RouletteStandard}
	frames := RouletteFrames(result, cfg)

	// 28 animation frames + 1 final = 29
	if len(frames) != 29 {
		t.Fatalf("frame count = %d, want 29", len(frames))
	}

	// Final frame should have zero delay.
	last := frames[len(frames)-1]
	if last.Delay != 0 {
		t.Fatalf("final frame delay = %v, want 0", last.Delay)
	}

	// Final frame should end with newline.
	if last.Content[len(last.Content)-1] != '\n' {
		t.Fatal("final frame should end with newline")
	}

	// Every frame should have Lines populated.
	for i, f := range frames {
		if len(f.Lines) == 0 {
			t.Fatalf("frame %d has no Lines", i)
		}
	}
}

func TestCoinFlipFrameCount(t *testing.T) {
	cfg := model.CoinFlipConfig{Heads: "Heads", Tails: "Tails"}
	result := model.CoinFlipResult{Value: "Heads", IsHeads: true, Config: cfg}
	frames := CoinFlipFrames(result)

	// 20 animation + 1 final = 21
	if len(frames) != 21 {
		t.Fatalf("frame count = %d, want 21", len(frames))
	}
}

func TestDiceFrameCount(t *testing.T) {
	result := model.DiceResult{
		Dice:   []int{3, 5},
		Sum:    8,
		Config: model.DiceConfig{Count: 2, Sides: 6},
	}
	frames := DiceFrames(result)

	// 22 animation + 1 final = 23
	if len(frames) != 23 {
		t.Fatalf("frame count = %d, want 23", len(frames))
	}
}

func TestSlotsFramesStandard(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsStandard, Rows: 3, Cols: 3}
	result := model.SlotsResult{
		Grid:       [][]string{{"CH", "LM", "OR"}, {"GR", "BL", "DI"}, {"7s", "**", "$$"}},
		Multiplier: 1,
		Mode:       model.SlotsStandard,
		Config:     cfg,
	}
	frames := SlotsFrames(result, cfg)

	if len(frames) == 0 {
		t.Fatal("expected frames")
	}

	last := frames[len(frames)-1]
	if last.Delay != 0 {
		t.Fatalf("final frame delay = %v, want 0", last.Delay)
	}
	if len(last.Lines) == 0 {
		t.Fatal("final frame has no lines")
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"", 0},
		{Bold + "hi" + Reset, 2},
		{Red + "OK" + Reset, 2},
	}
	for _, tt := range tests {
		got := displayWidth(tt.input)
		if got != tt.want {
			t.Fatalf("displayWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCenterPad(t *testing.T) {
	got := centerPad("AB", 6)
	if len(got) != 6 {
		t.Fatalf("centerPad(\"AB\", 6) len = %d, want 6", len(got))
	}
	if got != "  AB  " {
		t.Fatalf("centerPad(\"AB\", 6) = %q, want %q", got, "  AB  ")
	}
}

func TestCenterPadWithANSI(t *testing.T) {
	s := Bold + "X" + Reset
	got := centerPad(s, 5)
	// Display width of "X" is 1, padded to 5 = 2 left + X + 2 right
	dw := displayWidth(got)
	if dw != 5 {
		t.Fatalf("displayWidth of padded = %d, want 5", dw)
	}
}
