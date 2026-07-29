package coinflip

import "testing"

func TestFlipOutcomes(t *testing.T) {
	const tosses = 200_000

	counts := map[string]int{}
	for range tosses {
		counts[flip()]++
	}

	for result := range counts {
		switch result {
		case msgHeads, msgTails, msgEdge:
		default:
			t.Fatalf("unexpected flip result: %q", result)
		}
	}

	heads, tails := counts[msgHeads], counts[msgTails]
	if heads == 0 || tails == 0 {
		t.Fatalf("expected both outcomes, got heads=%d tails=%d", heads, tails)
	}

	ratio := float64(heads) / float64(heads+tails)
	if ratio < 0.48 || ratio > 0.52 {
		t.Errorf("heads ratio = %.4f, want ~0.50 (heads=%d tails=%d edge=%d)",
			ratio, heads, tails, counts[msgEdge])
	}
}
