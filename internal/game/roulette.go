package game

import (
	"fmt"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

// SpinRoulette executes a single roulette spin. Pure function.
func SpinRoulette(cfg model.RouletteConfig) model.RouletteResult {
	switch cfg.Mode {
	case model.RouletteStandard:
		return spinStandard()
	case model.RouletteMinMax:
		return spinMinMax(cfg)
	case model.RouletteCustomValues:
		return spinCustomValues(cfg)
	default:
		panic(fmt.Sprintf("unhandled roulette mode: %d", cfg.Mode))
	}
}

func spinStandard() model.RouletteResult {
	n := RandomInt(0, 36)
	color := model.RouletteColors[n]
	return model.RouletteResult{
		Value:  fmt.Sprintf("%d", n),
		Number: n,
		Color:  color,
		IsZero: n == 0,
		Mode:   model.RouletteStandard,
	}
}

func spinMinMax(cfg model.RouletteConfig) model.RouletteResult {
	n := RandomInt(cfg.Min, cfg.Max)
	return model.RouletteResult{
		Value:  fmt.Sprintf("%d", n),
		Number: n,
		Mode:   model.RouletteMinMax,
		Config: cfg,
	}
}

func spinCustomValues(cfg model.RouletteConfig) model.RouletteResult {
	val := RandomChoice(cfg.Values)
	return model.RouletteResult{
		Value:  val,
		Number: -1,
		Mode:   model.RouletteCustomValues,
		Config: cfg,
	}
}
