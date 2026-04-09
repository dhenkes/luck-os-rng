package game

import (
	"fmt"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

// DetectSlotsNearMiss checks if the result was close to a bigger win.
func DetectSlotsNearMiss(result model.SlotsResult, cfg model.SlotsConfig) []string {
	if result.Mode != model.SlotsStandard {
		return nil
	}
	if len(result.Paylines) > 0 {
		return nil // already won, no near-miss
	}

	var msgs []string
	grid := result.Grid

	// Check paylines for 2-match (one short of a win).
	for _, pattern := range model.StandardPaylines {
		if len(pattern) > cfg.Cols {
			pattern = pattern[:cfg.Cols]
		}
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

		firstSym := grid[pattern[0]][0]
		if firstSym == model.SymbolScatter {
			continue
		}

		matchSym := firstSym
		if matchSym == model.SymbolWild {
			matchSym = ""
			for c := 1; c < len(pattern); c++ {
				sym := grid[pattern[c]][c]
				if sym != model.SymbolWild && sym != model.SymbolScatter {
					matchSym = sym
					break
				}
			}
			if matchSym == "" {
				continue
			}
		}

		count := 0
		for c := 0; c < len(pattern); c++ {
			sym := grid[pattern[c]][c]
			if sym == matchSym || sym == model.SymbolWild {
				count++
			} else {
				break
			}
		}

		if count == 2 {
			name := symbolDisplayName(matchSym)
			// Only report interesting near-misses (rarer symbols).
			if isRareSymbol(matchSym) {
				msgs = append(msgs, fmt.Sprintf("SO CLOSE! One %s away from TRIPLE %sS", name, name))
				break // one near-miss message is enough
			}
		}
	}

	// Check for near-scatter bonus (2 scatters = one away from free spins).
	scatterCount := 0
	for _, row := range grid {
		for _, cell := range row {
			if cell == model.SymbolScatter {
				scatterCount++
			}
		}
	}
	if scatterCount == 2 {
		msgs = append(msgs, "SO CLOSE! One $$ away from FREE SPINS")
	}

	return msgs
}

// DetectRouletteNearMiss checks if the result was close to notable numbers.
func DetectRouletteNearMiss(result model.RouletteResult) []string {
	if result.Mode != model.RouletteStandard {
		return nil
	}

	var msgs []string

	// Find position in wheel order.
	pos := -1
	for i, n := range model.WheelOrder {
		if n == result.Number {
			pos = i
			break
		}
	}
	if pos < 0 {
		return nil
	}

	// Check neighbors.
	wheel := model.WheelOrder
	prev := wheel[(pos-1+len(wheel))%len(wheel)]
	next := wheel[(pos+1)%len(wheel)]

	if prev == 0 || next == 0 {
		msgs = append(msgs, "Just one pocket from GREEN ZERO")
	}

	// Check if close to 7 or 13 (lucky/unlucky).
	for _, notable := range []int{7, 13, 36} {
		if result.Number != notable && (prev == notable || next == notable) {
			msgs = append(msgs, fmt.Sprintf("One pocket from %d", notable))
			break
		}
	}

	return msgs
}

// DetectDiceNearMiss checks if the result was close to max.
func DetectDiceNearMiss(result model.DiceResult, cfg model.DiceConfig) []string {
	maxPossible := cfg.Count * cfg.Sides
	if result.Sum == maxPossible-1 {
		return []string{"ONE AWAY FROM MAX ROLL"}
	}

	// Check for all-but-one matching.
	if cfg.Count >= 3 {
		counts := make(map[int]int)
		for _, d := range result.Dice {
			counts[d]++
		}
		for _, c := range counts {
			if c == cfg.Count-1 {
				return []string{"SO CLOSE to all matching"}
			}
		}
	}

	return nil
}

func symbolDisplayName(sym string) string {
	for _, s := range model.StandardSymbols {
		if s.Symbol == sym {
			return s.Name
		}
	}
	return sym
}

func isRareSymbol(sym string) bool {
	// BL, DI, 7s, ** are considered rare.
	switch sym {
	case "BL", "DI", "7s", "**":
		return true
	}
	return false
}
