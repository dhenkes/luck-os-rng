package game

import (
	"strconv"
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestSpinRouletteStandard(t *testing.T) {
	for range 100 {
		r := SpinRoulette(model.RouletteConfig{Mode: model.RouletteStandard})
		if r.Number < 0 || r.Number > 36 {
			t.Fatalf("standard spin number %d out of range [0, 36]", r.Number)
		}
		if r.Value != strconv.Itoa(r.Number) {
			t.Fatalf("value %q != number %d", r.Value, r.Number)
		}
		if r.Number == 0 {
			if r.Color != model.RouletteGreen || !r.IsZero {
				t.Fatalf("zero should be green and IsZero")
			}
		} else {
			if r.IsZero {
				t.Fatalf("non-zero number %d has IsZero=true", r.Number)
			}
		}
		if r.Mode != model.RouletteStandard {
			t.Fatalf("mode = %v, want standard", r.Mode)
		}
	}
}

func TestSpinRouletteMinMax(t *testing.T) {
	cfg := model.RouletteConfig{Mode: model.RouletteMinMax, Min: 10, Max: 20}
	for range 100 {
		r := SpinRoulette(cfg)
		if r.Number < 10 || r.Number > 20 {
			t.Fatalf("minmax spin number %d out of range [10, 20]", r.Number)
		}
		if r.Mode != model.RouletteMinMax {
			t.Fatalf("mode = %v, want minmax", r.Mode)
		}
	}
}

func TestSpinRouletteCustomValues(t *testing.T) {
	values := []string{"pizza", "sushi", "tacos"}
	cfg := model.RouletteConfig{Mode: model.RouletteCustomValues, Values: values}
	seen := map[string]bool{}
	for range 1000 {
		r := SpinRoulette(cfg)
		seen[r.Value] = true
		if r.Number != -1 {
			t.Fatalf("custom mode number should be -1, got %d", r.Number)
		}
	}
	for _, v := range values {
		if !seen[v] {
			t.Fatalf("custom mode never returned %q", v)
		}
	}
}
