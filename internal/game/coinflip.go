package game

import "github.com/dhenkes/luck-os-rng/internal/model"

// FlipCoin flips a coin. Pure function.
func FlipCoin(cfg model.CoinFlipConfig) model.CoinFlipResult {
	isHeads := RandomInt(0, 1) == 0
	value := cfg.Tails
	if isHeads {
		value = cfg.Heads
	}
	return model.CoinFlipResult{
		Value:   value,
		IsHeads: isHeads,
		Config:  cfg,
	}
}
