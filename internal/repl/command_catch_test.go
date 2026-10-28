package repl

import "testing"

func TestCatchChance(t *testing.T) {
	// Chance must always be a valid, non-zero probability...
	for _, exp := range []int{-5, 0, 36, 112, 306, 608, 100000} {
		p := catchChance(exp)
		if p <= 0 || p > 1 {
			t.Errorf("catchChance(%d) = %v, want 0 < p <= 1", exp, p)
		}
	}

	// ...that gets lower as base experience goes up.
	if !(catchChance(36) > catchChance(112) && catchChance(112) > catchChance(608)) {
		t.Error("expected catch chance to decrease as base experience increases")
	}

	// Even the hardest Pokemon should be catchable in a reasonable number of
	// tries: expected tries is 1/p.
	if tries := 1 / catchChance(700); tries > 10 {
		t.Errorf("expected at most ~10 tries on average for the hardest Pokemon, got %.1f", tries)
	}
}
