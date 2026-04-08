package game

import (
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestSpinSlotsStandard(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsStandard, Rows: 3, Cols: 5}
	r := SpinSlots(cfg)

	if len(r.Grid) != 3 {
		t.Fatalf("grid rows = %d, want 3", len(r.Grid))
	}
	for i, row := range r.Grid {
		if len(row) != 5 {
			t.Fatalf("grid row %d cols = %d, want 5", i, len(row))
		}
		for j, sym := range row {
			if sym == "" {
				t.Fatalf("grid[%d][%d] is empty", i, j)
			}
		}
	}
	if r.Multiplier < 1 {
		t.Fatalf("multiplier = %d, want >= 1", r.Multiplier)
	}
	if r.Mode != model.SlotsStandard {
		t.Fatalf("mode = %v, want standard", r.Mode)
	}
}

func TestSpinSlotsStandardSmall(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsStandard, Rows: 1, Cols: 3}
	r := SpinSlots(cfg)
	if len(r.Grid) != 1 || len(r.Grid[0]) != 3 {
		t.Fatalf("grid dimensions wrong: %dx%d", len(r.Grid), len(r.Grid[0]))
	}
}

func TestSpinSlotsLuckInsane(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsStandard, Rows: 3, Cols: 5, Luck: model.SlotsLuckInsane}

	// Insane mode should almost always produce wins. Run a few times.
	wins := 0
	for range 20 {
		r := SpinSlots(cfg)
		if len(r.Paylines) > 0 {
			wins++
		}
	}
	if wins < 10 {
		t.Fatalf("insane mode only produced wins %d/20 times, expected nearly all", wins)
	}
}

func TestSpinSlotsLuckInsaneCascades(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsStandard, Rows: 3, Cols: 5, Luck: model.SlotsLuckInsane}

	cascaded := 0
	for range 20 {
		r := SpinSlots(cfg)
		if r.Multiplier > 1 {
			cascaded++
		}
	}
	if cascaded < 5 {
		t.Fatalf("insane mode only cascaded %d/20 times, expected more", cascaded)
	}
}

func TestSpinSlotsMinMax(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsMinMax, Cols: 3, Min: 10, Max: 50, Operation: model.SlotsAdd}
	for range 50 {
		r := SpinSlots(cfg)
		if r.FinalNumber < 10 || r.FinalNumber > 50 {
			t.Fatalf("minmax result %d out of range [10, 50]", r.FinalNumber)
		}
		if len(r.Grid) != 1 || len(r.Grid[0]) != 3 {
			t.Fatalf("minmax grid dimensions wrong")
		}
	}
}

func TestSpinSlotsMinMaxMultiply(t *testing.T) {
	cfg := model.SlotsConfig{Mode: model.SlotsMinMax, Cols: 2, Min: 1, Max: 100, Operation: model.SlotsMultiply}
	for range 50 {
		r := SpinSlots(cfg)
		if r.FinalNumber < 1 || r.FinalNumber > 100 {
			t.Fatalf("minmax multiply result %d out of range [1, 100]", r.FinalNumber)
		}
	}
}

func TestSpinSlotsCustom(t *testing.T) {
	cfg := model.SlotsConfig{
		Mode: model.SlotsCustom,
		ReelValues: [][]string{
			{"a", "b"},
			{"x", "y"},
			{"1", "2"},
		},
	}
	for range 100 {
		r := SpinSlots(cfg)
		if len(r.Grid) != 1 || len(r.Grid[0]) != 3 {
			t.Fatalf("custom grid dimensions wrong")
		}
		for c, val := range r.Grid[0] {
			found := false
			for _, v := range cfg.ReelValues[c] {
				if v == val {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("grid[0][%d] = %q, not in reel values", c, val)
			}
		}
	}
}
