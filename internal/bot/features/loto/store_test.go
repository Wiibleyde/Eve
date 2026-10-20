package loto

import (
	"slices"
	"testing"

	"Eve/internal/database/ent"
)

func prizeIDs(prizes []*ent.LotoPrize) []string {
	ids := make([]string, 0, len(prizes))
	for _, p := range prizes {
		ids = append(ids, p.ID)
	}
	return ids
}

func samplePrizes() []*ent.LotoPrize {
	return []*ent.LotoPrize{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
}

func TestMoveToRank(t *testing.T) {
	cases := []struct {
		name string
		from int
		rank int
		want []string
	}{
		{"same place", 1, 2, []string{"a", "b", "c", "d"}},
		{"to first", 2, 1, []string{"c", "a", "b", "d"}},
		{"to last", 0, 4, []string{"b", "c", "d", "a"}},
		{"forward", 0, 3, []string{"b", "c", "a", "d"}},
		{"backward", 3, 2, []string{"a", "d", "b", "c"}},
		{"clamped above", 1, 99, []string{"a", "c", "d", "b"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := prizeIDs(moveToRank(samplePrizes(), tc.from, tc.rank))
			if !slices.Equal(got, tc.want) {
				t.Errorf("moveToRank(%d, %d) = %v, want %v", tc.from, tc.rank, got, tc.want)
			}
		})
	}
}

func TestNextFreePosition(t *testing.T) {
	prizes := []*ent.LotoPrize{{Position: 0}, {Position: 4}, {Position: 2}}
	if got := nextFreePosition(prizes); got != 5 {
		t.Errorf("nextFreePosition = %d, want 5", got)
	}
	if got := nextFreePosition(nil); got != 0 {
		t.Errorf("nextFreePosition(nil) = %d, want 0", got)
	}
}
