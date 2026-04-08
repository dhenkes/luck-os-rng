package model

type DiceConfig struct {
	Count int // number of dice (1-10)
	Sides int // sides per die (4, 6, 8, 10, 12, 20)
}

func (c *DiceConfig) ApplyDefaults() {
	if c.Count <= 0 {
		c.Count = 1
	}
	if c.Sides <= 0 {
		c.Sides = 6
	}
}

func (c DiceConfig) Validate() error {
	v := NewValidationErrors()
	if c.Count < 1 || c.Count > 10 {
		v.Add("count", "must be between 1 and 10")
	}
	switch c.Sides {
	case 4, 6, 8, 10, 12, 20:
		// valid
	default:
		v.Add("sides", "must be 4, 6, 8, 10, 12, or 20")
	}
	return v.OrNil()
}

type DiceResult struct {
	Dice   []int `json:"dice"`
	Sum    int   `json:"sum"`
	Config DiceConfig `json:"-"`
}
