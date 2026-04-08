package game

import (
	"fmt"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

// SpinSlots executes a single slots spin. Pure function.
func SpinSlots(cfg model.SlotsConfig) model.SlotsResult {
	switch cfg.Mode {
	case model.SlotsStandard:
		return spinSlotsStandard(cfg)
	case model.SlotsMinMax:
		return spinSlotsMinMax(cfg)
	case model.SlotsCustom:
		return spinSlotsCustom(cfg)
	default:
		panic(fmt.Sprintf("unhandled slots mode: %d", cfg.Mode))
	}
}

func spinSlotsStandard(cfg model.SlotsConfig) model.SlotsResult {
	weights := standardWeights()

	var grid [][]string
	switch cfg.Luck {
	case model.SlotsLuckHigh:
		grid = generateLuckyGrid(cfg.Rows, cfg.Cols, weights)
	case model.SlotsLuckInsane:
		grid = generateInsaneGrid(cfg.Rows, cfg.Cols)
	default:
		grid = generateGrid(cfg.Rows, cfg.Cols, weights)
	}

	paylines, cascadeSteps, multiplier := cascade(grid, cfg, weights)

	scatterCount := countSymbol(grid, model.SymbolScatter)
	freeSpins := 0
	bonusRound := false
	if scatterCount >= 3 {
		freeSpins = 5
		if scatterCount >= 4 {
			bonusRound = true
		}
	}

	return model.SlotsResult{
		Grid:         grid,
		Paylines:     paylines,
		Multiplier:   multiplier,
		FreeSpins:    freeSpins,
		BonusRound:   bonusRound,
		CascadeSteps: cascadeSteps,
		Mode:         model.SlotsStandard,
		Config:       cfg,
	}
}

func standardWeights() []int {
	weights := make([]int, len(model.StandardSymbols))
	for i, s := range model.StandardSymbols {
		weights[i] = s.Weight
	}
	return weights
}

func generateGrid(rows, cols int, weights []int) [][]string {
	grid := make([][]string, rows)
	for r := range rows {
		grid[r] = make([]string, cols)
		for c := range cols {
			idx := RandomWeighted(weights)
			grid[r][c] = model.StandardSymbols[idx].Symbol
		}
	}
	return grid
}

// generateLuckyGrid biases toward wins by picking fewer distinct symbols.
// Each row picks 1-2 dominant symbols and fills most columns with them.
func generateLuckyGrid(rows, cols int, weights []int) [][]string {
	grid := make([][]string, rows)
	for r := range rows {
		grid[r] = make([]string, cols)

		// Pick a dominant symbol for this row (skip wild/scatter).
		domIdx := RandomWeighted(weights[:len(weights)-2])
		dominant := model.StandardSymbols[domIdx].Symbol

		for c := range cols {
			// 70% chance of the dominant symbol, 30% random.
			if RandomInt(0, 99) < 70 {
				grid[r][c] = dominant
			} else {
				idx := RandomWeighted(weights)
				grid[r][c] = model.StandardSymbols[idx].Symbol
			}
		}
	}
	return grid
}

// generateInsaneGrid guarantees multiple wins and cascades.
// Forces matching runs on several paylines.
func generateInsaneGrid(rows, cols int) [][]string {
	grid := make([][]string, rows)
	for r := range rows {
		grid[r] = make([]string, cols)
	}

	// Pick 2-3 symbols to dominate the grid.
	syms := []string{
		model.StandardSymbols[RandomInt(0, 5)].Symbol,
		model.StandardSymbols[RandomInt(0, 5)].Symbol,
	}

	// Fill the grid mostly with these symbols.
	for r := range rows {
		for c := range cols {
			grid[r][c] = syms[RandomInt(0, len(syms)-1)]
		}
	}

	// Sprinkle a couple of wilds for extra wins.
	if rows > 0 && cols > 2 {
		grid[RandomInt(0, rows-1)][RandomInt(0, cols-1)] = model.SymbolWild
	}

	// Add one random symbol to prevent it from being too uniform.
	weights := standardWeights()
	for range 2 {
		r := RandomInt(0, rows-1)
		c := RandomInt(0, cols-1)
		idx := RandomWeighted(weights)
		grid[r][c] = model.StandardSymbols[idx].Symbol
	}

	return grid
}

// cascade repeatedly checks for wins, removes them, drops symbols, and refills.
// Returns all cascade steps so the renderer can animate each one.
func cascade(grid [][]string, cfg model.SlotsConfig, weights []int) ([]model.Payline, []model.CascadeStep, int) {
	var allPaylines []model.Payline
	var steps []model.CascadeStep
	multiplier := 1

	for {
		wins := evaluatePaylines(grid, cfg)
		if len(wins) == 0 {
			break
		}
		for i := range wins {
			allPaylines = append(allPaylines, wins[i])
		}

		// Snapshot the grid before removal.
		gridBefore := copyGrid(grid)

		// Remove winning symbols.
		removed := make(map[model.Position]bool)
		for _, pl := range wins {
			for _, pos := range pl.Positions {
				removed[pos] = true
			}
		}

		// Drop and refill.
		for c := range cfg.Cols {
			var kept []string
			for r := cfg.Rows - 1; r >= 0; r-- {
				if !removed[model.Position{Row: r, Col: c}] {
					kept = append(kept, grid[r][c])
				}
			}
			for r := cfg.Rows - 1; r >= 0; r-- {
				idx := cfg.Rows - 1 - r
				if idx < len(kept) {
					grid[r][c] = kept[idx]
				} else {
					symIdx := RandomWeighted(weights)
					grid[r][c] = model.StandardSymbols[symIdx].Symbol
				}
			}
		}

		steps = append(steps, model.CascadeStep{
			Grid:      gridBefore,
			Wins:      wins,
			Removed:   removed,
			GridAfter: copyGrid(grid),
		})

		multiplier++
	}

	return allPaylines, steps, multiplier
}

func copyGrid(grid [][]string) [][]string {
	cp := make([][]string, len(grid))
	for r := range grid {
		cp[r] = make([]string, len(grid[r]))
		copy(cp[r], grid[r])
	}
	return cp
}

func evaluatePaylines(grid [][]string, cfg model.SlotsConfig) []model.Payline {
	var wins []model.Payline

	paylinePatterns := model.StandardPaylines
	// Trim paylines to match actual grid dimensions.
	for i, pattern := range paylinePatterns {
		if len(pattern) > cfg.Cols {
			pattern = pattern[:cfg.Cols]
		}
		// Skip patterns that reference rows outside our grid.
		valid := true
		for _, row := range pattern {
			if row >= cfg.Rows {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		// Check this payline for a win (3+ matching from left).
		firstSym := grid[pattern[0]][0]
		if firstSym == model.SymbolScatter {
			continue // scatters don't count on paylines
		}

		matchSym := firstSym
		if matchSym == model.SymbolWild {
			// If first is wild, find the first non-wild to determine the matching symbol.
			matchSym = ""
			for c := 1; c < len(pattern); c++ {
				sym := grid[pattern[c]][c]
				if sym != model.SymbolWild && sym != model.SymbolScatter {
					matchSym = sym
					break
				}
			}
			if matchSym == "" {
				matchSym = model.SymbolWild // all wilds
			}
		}

		count := 0
		var positions []model.Position
		for c := 0; c < len(pattern); c++ {
			sym := grid[pattern[c]][c]
			if sym == matchSym || sym == model.SymbolWild {
				count++
				positions = append(positions, model.Position{Row: pattern[c], Col: c})
			} else {
				break
			}
		}

		if count >= 3 {
			wins = append(wins, model.Payline{
				LineIndex: i,
				Symbol:    matchSym,
				Count:     count,
				Positions: positions,
			})
		}
	}

	return wins
}

func countSymbol(grid [][]string, symbol string) int {
	count := 0
	for _, row := range grid {
		for _, cell := range row {
			if cell == symbol {
				count++
			}
		}
	}
	return count
}

func spinSlotsMinMax(cfg model.SlotsConfig) model.SlotsResult {
	target := RandomInt(cfg.Min, cfg.Max)
	grid := make([][]string, 1)
	grid[0] = make([]string, cfg.Cols)

	switch cfg.Operation {
	case model.SlotsAdd:
		values := partitionSum(target, cfg.Cols)
		for c, v := range values {
			grid[0][c] = fmt.Sprintf("%d", v)
		}
	case model.SlotsMultiply:
		values := partitionProduct(target, cfg.Cols)
		for c, v := range values {
			grid[0][c] = fmt.Sprintf("%d", v)
		}
	}

	return model.SlotsResult{
		Grid:        grid,
		FinalNumber: target,
		Mode:        model.SlotsMinMax,
		Multiplier:  1,
		Operation:   cfg.Operation,
		Config:      cfg,
	}
}

// partitionSum splits n into cols random positive parts that sum to n.
func partitionSum(n, cols int) []int {
	if cols <= 0 {
		return nil
	}
	if cols == 1 {
		return []int{n}
	}

	parts := make([]int, cols)
	remaining := n

	for i := range cols - 1 {
		// Ensure each remaining part can be at least 0.
		maxVal := remaining - (cols - 1 - i) * 0
		if maxVal < 0 {
			maxVal = 0
		}
		if maxVal > remaining {
			maxVal = remaining
		}
		parts[i] = RandomInt(0, maxVal)
		remaining -= parts[i]
	}
	parts[cols-1] = remaining

	return parts
}

// partitionProduct splits n into cols factors that multiply to approximately n.
func partitionProduct(n, cols int) []int {
	if cols <= 0 {
		return nil
	}
	if cols == 1 {
		return []int{n}
	}
	if n == 0 {
		parts := make([]int, cols)
		parts[0] = 0
		for i := 1; i < cols; i++ {
			parts[i] = RandomInt(1, 9)
		}
		return parts
	}

	parts := make([]int, cols)
	remaining := n
	if remaining < 0 {
		remaining = -remaining
	}

	for i := range cols - 1 {
		if remaining <= 1 {
			parts[i] = 1
			continue
		}
		// Pick a factor of remaining, or a small number.
		factors := findFactors(remaining)
		if len(factors) > 0 {
			parts[i] = RandomChoice(factors)
		} else {
			parts[i] = 1
		}
		remaining /= parts[i]
	}
	parts[cols-1] = remaining

	if n < 0 {
		parts[0] = -parts[0]
	}

	return parts
}

func findFactors(n int) []int {
	if n <= 1 {
		return []int{1}
	}
	var factors []int
	for i := 2; i*i <= n && i <= 20; i++ {
		if n%i == 0 {
			factors = append(factors, i)
		}
	}
	if len(factors) == 0 {
		factors = []int{1}
	}
	return factors
}

func spinSlotsCustom(cfg model.SlotsConfig) model.SlotsResult {
	grid := make([][]string, 1)
	grid[0] = make([]string, len(cfg.ReelValues))

	for c, values := range cfg.ReelValues {
		grid[0][c] = RandomChoice(values)
	}

	return model.SlotsResult{
		Grid:       grid,
		Mode:       model.SlotsCustom,
		Multiplier: 1,
		Config:     cfg,
	}
}
