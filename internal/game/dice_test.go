package game

import (
	"testing"

	"github.com/dhenkes/luck-os-rng/internal/model"
)

func TestRollDice(t *testing.T) {
	cfg := model.DiceConfig{Count: 3, Sides: 6}
	for range 100 {
		r := RollDice(cfg)
		if len(r.Dice) != 3 {
			t.Fatalf("dice count = %d, want 3", len(r.Dice))
		}
		sum := 0
		for _, d := range r.Dice {
			if d < 1 || d > 6 {
				t.Fatalf("die value %d out of range [1, 6]", d)
			}
			sum += d
		}
		if r.Sum != sum {
			t.Fatalf("sum = %d, computed = %d", r.Sum, sum)
		}
	}
}

func TestRollDiceSides(t *testing.T) {
	for _, sides := range []int{4, 6, 8, 10, 12, 20} {
		cfg := model.DiceConfig{Count: 1, Sides: sides}
		for range 100 {
			r := RollDice(cfg)
			if r.Dice[0] < 1 || r.Dice[0] > sides {
				t.Fatalf("d%d rolled %d", sides, r.Dice[0])
			}
		}
	}
}
