package model

import "fmt"

type RouletteMode int

const (
	RouletteStandard     RouletteMode = iota // 0-36, standard European wheel
	RouletteMinMax                           // custom min/max range
	RouletteCustomValues                     // custom list of values
)

func (m RouletteMode) String() string {
	switch m {
	case RouletteStandard:
		return "standard"
	case RouletteMinMax:
		return "minmax"
	case RouletteCustomValues:
		return "custom"
	default:
		return "unknown"
	}
}

func ParseRouletteMode(s string) (RouletteMode, error) {
	switch s {
	case "", "standard":
		return RouletteStandard, nil
	case "minmax":
		return RouletteMinMax, nil
	case "custom":
		return RouletteCustomValues, nil
	default:
		return 0, fmt.Errorf("unknown roulette mode: %q", s)
	}
}

type RouletteConfig struct {
	Mode   RouletteMode
	Min    int      // for MinMax mode
	Max    int      // for MinMax mode
	Values []string // for CustomValues mode
}

func (c RouletteConfig) Validate() error {
	v := NewValidationErrors()
	switch c.Mode {
	case RouletteStandard:
		// no config needed
	case RouletteMinMax:
		if c.Min >= c.Max {
			v.Add("min", "must be less than max")
		}
		if c.Min < 0 {
			v.Add("min", "must be non-negative")
		}
		if c.Max-c.Min > 10000 {
			v.Add("max", "range must not exceed 10000")
		}
	case RouletteCustomValues:
		if len(c.Values) < 2 {
			v.Add("values", "must have at least 2 values")
		}
		for i, val := range c.Values {
			if val == "" {
				v.Add(fmt.Sprintf("values[%d]", i), "must not be empty")
			}
		}
	}
	return v.OrNil()
}

type RouletteColor int

const (
	RouletteRed RouletteColor = iota
	RouletteBlack
	RouletteGreen
)

func (c RouletteColor) String() string {
	switch c {
	case RouletteRed:
		return "RED"
	case RouletteBlack:
		return "BLACK"
	case RouletteGreen:
		return "GREEN"
	default:
		return ""
	}
}

type RouletteResult struct {
	Value  string        `json:"value"`
	Number int           `json:"number"`
	Color  RouletteColor `json:"color"`
	IsZero bool          `json:"is_zero"`
	Mode   RouletteMode  `json:"mode"`
	Config RouletteConfig `json:"-"`
}

// Standard European roulette: number-to-color mapping.
// 0 = green, then alternating red/black per the real wheel.
var RouletteColors = map[int]RouletteColor{
	0: RouletteGreen,
	1: RouletteRed, 2: RouletteBlack, 3: RouletteRed, 4: RouletteBlack,
	5: RouletteRed, 6: RouletteBlack, 7: RouletteRed, 8: RouletteBlack,
	9: RouletteRed, 10: RouletteBlack, 11: RouletteBlack, 12: RouletteRed,
	13: RouletteBlack, 14: RouletteRed, 15: RouletteBlack, 16: RouletteRed,
	17: RouletteBlack, 18: RouletteRed, 19: RouletteRed, 20: RouletteBlack,
	21: RouletteRed, 22: RouletteBlack, 23: RouletteRed, 24: RouletteBlack,
	25: RouletteRed, 26: RouletteBlack, 27: RouletteRed, 28: RouletteBlack,
	29: RouletteBlack, 30: RouletteRed, 31: RouletteBlack, 32: RouletteRed,
	33: RouletteBlack, 34: RouletteRed, 35: RouletteBlack, 36: RouletteRed,
}

// European wheel order (clockwise).
var WheelOrder = []int{
	0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8,
	23, 10, 5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7, 28, 12,
	35, 3, 26,
}
