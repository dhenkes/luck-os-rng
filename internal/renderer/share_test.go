package renderer

import (
	"strings"
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestShareBlockSlots(t *testing.T) {
	result := model.SlotsResult{
		Mode:       model.SlotsStandard,
		Grid:       [][]string{{"CH", "CH", "CH", "LM", "OR"}, {"LM", "OR", "GR", "BL", "DI"}, {"OR", "GR", "BL", "DI", "7s"}},
		Paylines:   []model.Payline{{LineIndex: 0, Symbol: "CH", Count: 3}},
		Multiplier: 1,
	}
	block := ShareBlockSlots(result)
	if len(block) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(block))
	}
	content := strings.Join(block, "\n")
	if !strings.Contains(content, "LUCK SLOTS") {
		t.Fatal("missing header")
	}
	if !strings.Contains(content, "WIN") {
		t.Fatal("missing win status")
	}
}

func TestShareBlockSlotsNonStandard(t *testing.T) {
	result := model.SlotsResult{Mode: model.SlotsMinMax}
	block := ShareBlockSlots(result)
	if block != nil {
		t.Fatal("non-standard mode should not produce share block")
	}
}

func TestShareBlockRoulette(t *testing.T) {
	result := model.RouletteResult{Value: "17", Color: model.RouletteRed}
	block := ShareBlockRoulette(result)
	if len(block) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(block))
	}
	content := strings.Join(block, "\n")
	if !strings.Contains(content, "17") {
		t.Fatal("missing result value")
	}
}

func TestShareBlockCoinFlip(t *testing.T) {
	result := model.CoinFlipResult{Value: "Heads"}
	block := ShareBlockCoinFlip(result)
	content := strings.Join(block, "\n")
	if !strings.Contains(content, "Heads") {
		t.Fatal("missing result")
	}
}

func TestShareBlockDice(t *testing.T) {
	result := model.DiceResult{Dice: []int{3, 5}, Sum: 8}
	block := ShareBlockDice(result)
	content := strings.Join(block, "\n")
	if !strings.Contains(content, "8") {
		t.Fatal("missing sum")
	}
}

func TestCashOutFrames(t *testing.T) {
	eng := model.EngagementState{Score: 1000}
	urls := map[string]string{
		"Slots":    "localhost/slots",
		"Roulette": "localhost/roulette",
	}
	frames := CashOutFrames(500, eng, urls)
	if len(frames) == 0 {
		t.Fatal("expected frames")
	}

	// Last frame should have game links.
	last := frames[len(frames)-1]
	content := strings.Join(last.Lines, "\n")
	if !strings.Contains(content, "CASHED OUT") {
		t.Fatal("missing cash out message")
	}
	if !strings.Contains(content, "500") {
		t.Fatal("missing amount")
	}
}

func TestSpinHeaderFrame(t *testing.T) {
	frame := SpinHeaderFrame(2, 5, 300)
	content := strings.Join(frame.Lines, "\n")
	if !strings.Contains(content, "SPIN 2/5") {
		t.Fatal("missing spin counter")
	}
	if !strings.Contains(content, "300") {
		t.Fatal("missing running score")
	}
}

func TestD6DiceLines(t *testing.T) {
	lines := buildD6Lines([]int{3, 6}, 9)
	if len(lines) == 0 {
		t.Fatal("expected lines")
	}
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "o") {
		t.Fatal("d6 faces should contain dot characters")
	}
	if !strings.Contains(content, "= 9") {
		t.Fatal("missing sum")
	}
}

func TestNumericDiceLines(t *testing.T) {
	lines := buildNumericDiceLines([]int{15}, 20, 15)
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "15") {
		t.Fatal("missing dice value")
	}
}

func TestRouletteWheelFrames(t *testing.T) {
	result := model.RouletteResult{Value: "0", Number: 0, Color: model.RouletteGreen, Mode: model.RouletteStandard}
	cfg := model.RouletteConfig{Mode: model.RouletteStandard}
	frames := RouletteFrames(result, cfg)

	// Final frame should contain the result.
	last := frames[len(frames)-1]
	content := strings.Join(last.Lines, "\n")
	if !strings.Contains(content, "GREEN") {
		t.Fatal("final frame should show GREEN for zero")
	}
}
