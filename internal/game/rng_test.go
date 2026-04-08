package game

import "testing"

func TestRandomInt(t *testing.T) {
	for range 1000 {
		n := RandomInt(5, 10)
		if n < 5 || n > 10 {
			t.Fatalf("RandomInt(5, 10) = %d, want [5, 10]", n)
		}
	}
}

func TestRandomIntSameMinMax(t *testing.T) {
	for range 100 {
		n := RandomInt(7, 7)
		if n != 7 {
			t.Fatalf("RandomInt(7, 7) = %d, want 7", n)
		}
	}
}

func TestRandomChoice(t *testing.T) {
	items := []string{"a", "b", "c"}
	seen := map[string]bool{}
	for range 1000 {
		seen[RandomChoice(items)] = true
	}
	for _, item := range items {
		if !seen[item] {
			t.Fatalf("RandomChoice never returned %q", item)
		}
	}
}

func TestRandomWeighted(t *testing.T) {
	weights := []int{100, 0, 0}
	for range 100 {
		idx := RandomWeighted(weights)
		if idx != 0 {
			t.Fatalf("RandomWeighted([100,0,0]) = %d, want 0", idx)
		}
	}
}

func TestRandomWeightedDistribution(t *testing.T) {
	weights := []int{1, 1, 1}
	counts := [3]int{}
	for range 3000 {
		counts[RandomWeighted(weights)]++
	}
	for i, c := range counts {
		if c < 500 {
			t.Fatalf("RandomWeighted index %d only hit %d/3000 times, expected ~1000", i, c)
		}
	}
}
