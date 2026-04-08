// Package game contains pure game logic. All randomness flows through crypto/rand.
package game

import (
	"crypto/rand"
	"math/big"
)

// RandomInt returns a cryptographically random integer in [min, max] (inclusive).
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return min + int(n.Int64())
}

// RandomChoice returns a random element from the slice.
func RandomChoice[T any](items []T) T {
	return items[RandomInt(0, len(items)-1)]
}

// RandomWeighted picks an index from weights where each weight is the relative
// probability of that index being chosen.
func RandomWeighted(weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	r := RandomInt(0, total-1)
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if r < cumulative {
			return i
		}
	}
	return len(weights) - 1
}
