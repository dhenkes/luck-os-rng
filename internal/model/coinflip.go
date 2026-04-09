package model

const maxLabelLen = 50

type CoinFlipConfig struct {
	Heads string // custom label for heads (default: "Heads")
	Tails string // custom label for tails (default: "Tails")
}

func (c *CoinFlipConfig) ApplyDefaults() {
	if c.Heads == "" {
		c.Heads = "Heads"
	}
	if c.Tails == "" {
		c.Tails = "Tails"
	}
}

func (c CoinFlipConfig) Validate() error {
	v := NewValidationErrors()
	if len(c.Heads) > maxLabelLen {
		v.Add("heads", "must be 50 characters or fewer")
	}
	if len(c.Tails) > maxLabelLen {
		v.Add("tails", "must be 50 characters or fewer")
	}
	return v.OrNil()
}

type CoinFlipResult struct {
	Value   string         `json:"value"`
	IsHeads bool           `json:"is_heads"`
	Config  CoinFlipConfig `json:"-"`
}
