package game

import (
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestFlipCoin(t *testing.T) {
	cfg := model.CoinFlipConfig{Heads: "H", Tails: "T"}
	heads, tails := 0, 0
	for range 1000 {
		r := FlipCoin(cfg)
		if r.IsHeads {
			heads++
			if r.Value != "H" {
				t.Fatalf("heads but value = %q", r.Value)
			}
		} else {
			tails++
			if r.Value != "T" {
				t.Fatalf("tails but value = %q", r.Value)
			}
		}
	}
	if heads < 300 || tails < 300 {
		t.Fatalf("distribution looks off: heads=%d tails=%d", heads, tails)
	}
}

func TestFlipCoinDefaults(t *testing.T) {
	cfg := model.CoinFlipConfig{}
	cfg.ApplyDefaults()
	r := FlipCoin(cfg)
	if r.Value != "Heads" && r.Value != "Tails" {
		t.Fatalf("unexpected value with defaults: %q", r.Value)
	}
}
