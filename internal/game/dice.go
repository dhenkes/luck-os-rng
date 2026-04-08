package game

import "github.com/dhenkes/luck-os-rng/internal/model"

// RollDice rolls the configured dice. Pure function.
func RollDice(cfg model.DiceConfig) model.DiceResult {
	dice := make([]int, cfg.Count)
	sum := 0
	for i := range cfg.Count {
		dice[i] = RandomInt(1, cfg.Sides)
		sum += dice[i]
	}
	return model.DiceResult{
		Dice:   dice,
		Sum:    sum,
		Config: cfg,
	}
}
