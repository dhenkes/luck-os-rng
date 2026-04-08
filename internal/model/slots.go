package model

import "fmt"

type SlotsMode int

const (
	SlotsStandard SlotsMode = iota // classic casino slots
	SlotsMinMax                    // numeric reels, combine result
	SlotsCustom                    // custom values per column
)

func (m SlotsMode) String() string {
	switch m {
	case SlotsStandard:
		return "standard"
	case SlotsMinMax:
		return "minmax"
	case SlotsCustom:
		return "custom"
	default:
		return "unknown"
	}
}

func ParseSlotsMode(s string) (SlotsMode, error) {
	switch s {
	case "", "standard":
		return SlotsStandard, nil
	case "minmax":
		return SlotsMinMax, nil
	case "custom":
		return SlotsCustom, nil
	default:
		return 0, fmt.Errorf("unknown slots mode: %q", s)
	}
}

type SlotsOperation int

const (
	SlotsAdd SlotsOperation = iota
	SlotsMultiply
)

func (o SlotsOperation) String() string {
	switch o {
	case SlotsAdd:
		return "add"
	case SlotsMultiply:
		return "multiply"
	default:
		return "add"
	}
}

func ParseSlotsOperation(s string) SlotsOperation {
	if s == "multiply" {
		return SlotsMultiply
	}
	return SlotsAdd
}

type SlotsLuck int

const (
	SlotsLuckNormal SlotsLuck = iota
	SlotsLuckHigh
	SlotsLuckInsane
)

func ParseSlotsLuck(s string) SlotsLuck {
	switch s {
	case "high":
		return SlotsLuckHigh
	case "insane":
		return SlotsLuckInsane
	default:
		return SlotsLuckNormal
	}
}

type SlotsConfig struct {
	Mode       SlotsMode
	Rows       int            // number of visible rows (standard: 3)
	Cols       int            // number of reels/columns (standard: 5)
	Min        int            // for MinMax mode
	Max        int            // for MinMax mode
	Operation  SlotsOperation // for MinMax: add or multiply
	ReelValues [][]string     // for Custom mode: values per column
	Luck       SlotsLuck      // bias toward wins
}

func DefaultSlotsConfig() SlotsConfig {
	return SlotsConfig{
		Mode: SlotsStandard,
		Rows: 3,
		Cols: 5,
	}
}

func (c SlotsConfig) Validate() error {
	v := NewValidationErrors()
	switch c.Mode {
	case SlotsStandard:
		if c.Rows < 1 || c.Rows > 5 {
			v.Add("rows", "must be between 1 and 5")
		}
		if c.Cols < 3 || c.Cols > 7 {
			v.Add("cols", "must be between 3 and 7")
		}
	case SlotsMinMax:
		if c.Min >= c.Max {
			v.Add("min", "must be less than max")
		}
		if c.Cols < 2 || c.Cols > 7 {
			v.Add("cols", "must be between 2 and 7")
		}
	case SlotsCustom:
		if len(c.ReelValues) < 2 {
			v.Add("reel_values", "must have at least 2 columns")
		}
		for i, col := range c.ReelValues {
			if len(col) < 2 {
				v.Add(fmt.Sprintf("reel_values[%d]", i), "must have at least 2 values")
			}
			for j, val := range col {
				if val == "" {
					v.Add(fmt.Sprintf("reel_values[%d][%d]", i, j), "must not be empty")
				}
			}
		}
	}
	return v.OrNil()
}

// Position represents a row, col coordinate in the slots grid.
type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// Payline represents a winning line in the grid.
type Payline struct {
	LineIndex int        `json:"line_index"`
	Symbol    string     `json:"symbol"`
	Count     int        `json:"count"`
	Positions []Position `json:"positions"`
}

// CascadeStep records one round of a cascade animation.
type CascadeStep struct {
	Grid      [][]string          // grid when wins were found
	Wins      []Payline           // paylines that won
	Removed   map[Position]bool   // positions removed
	GridAfter [][]string          // grid after drop + refill
}

// SlotsResult is the outcome of a slots spin.
type SlotsResult struct {
	Grid         [][]string    `json:"grid"`
	Paylines     []Payline     `json:"paylines,omitempty"`
	Multiplier   int           `json:"multiplier"`
	FreeSpins    int           `json:"free_spins"`
	BonusRound   bool          `json:"bonus_round"`
	FinalNumber  int           `json:"final_number,omitempty"`
	CascadeSteps []CascadeStep `json:"-"`
	Mode         SlotsMode     `json:"mode"`
	Operation    SlotsOperation `json:"-"`
	Config       SlotsConfig   `json:"-"`
}

// Standard mode symbols, ordered by rarity (rarest last).
// Two-char codes for consistent fixed-width display.
type SlotSymbol struct {
	Symbol string
	Name   string
	Weight int
}

var StandardSymbols = []SlotSymbol{
	{Symbol: "CH", Name: "Cherry", Weight: 20},
	{Symbol: "LM", Name: "Lemon", Weight: 18},
	{Symbol: "OR", Name: "Orange", Weight: 15},
	{Symbol: "GR", Name: "Grape", Weight: 12},
	{Symbol: "BL", Name: "Bell", Weight: 8},
	{Symbol: "DI", Name: "Diamond", Weight: 5},
	{Symbol: "7s", Name: "Seven", Weight: 3},
	{Symbol: "**", Name: "Wild", Weight: 2},
	{Symbol: "$$", Name: "Bonus", Weight: 1}, // scatter
}

const (
	SymbolWild    = "**"
	SymbolScatter = "$$"
)

// StandardPaylines defines payline patterns for a 5-col grid.
// Each pattern is an array of row indices, one per column.
var StandardPaylines = [][]int{
	{1, 1, 1, 1, 1}, // middle row
	{0, 0, 0, 0, 0}, // top row
	{2, 2, 2, 2, 2}, // bottom row
	{0, 1, 2, 1, 0}, // V shape
	{2, 1, 0, 1, 2}, // inverted V
	{0, 0, 1, 2, 2}, // diagonal down
	{2, 2, 1, 0, 0}, // diagonal up
	{1, 0, 0, 0, 1}, // U shape top
	{1, 2, 2, 2, 1}, // U shape bottom
	{0, 1, 0, 1, 0}, // zigzag top
	{2, 1, 2, 1, 2}, // zigzag bottom
	{1, 0, 1, 0, 1}, // W shape
	{1, 2, 1, 2, 1}, // M shape
	{0, 1, 1, 1, 0}, // mild V
	{2, 1, 1, 1, 2}, // mild inverted V
	{0, 0, 1, 0, 0}, // bump
	{2, 2, 1, 2, 2}, // inverted bump
	{1, 0, 1, 2, 1}, // S shape
	{1, 2, 1, 0, 1}, // inverted S
	{0, 2, 0, 2, 0}, // extreme zigzag
}
