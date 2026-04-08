package model

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

type CoinFlipResult struct {
	Value   string `json:"value"`
	IsHeads bool   `json:"is_heads"`
	Config  CoinFlipConfig `json:"-"`
}
