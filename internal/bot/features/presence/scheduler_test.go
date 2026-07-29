package presence

import (
	"testing"
	"time"

	"Eve/internal/bot/maintenance"

	"github.com/disgoorg/disgo/discord"
)

func TestSeasonalBoundaries(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2026, time.October, 23, 12, 0, 0, 0, time.UTC), listDefault},
		{time.Date(2026, time.October, 24, 0, 0, 0, 0, time.UTC), listHalloween},
		{time.Date(2026, time.November, 7, 23, 59, 0, 0, time.UTC), listHalloween},
		{time.Date(2026, time.November, 8, 0, 0, 0, 0, time.UTC), listDefault},
		{time.Date(2026, time.November, 30, 12, 0, 0, 0, time.UTC), listDefault},
		{time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC), listChristmas},
		{time.Date(2026, time.December, 25, 23, 59, 0, 0, time.UTC), listChristmas},
		{time.Date(2026, time.December, 26, 0, 0, 0, 0, time.UTC), listDefault},
		{time.Date(2027, time.January, 2, 0, 0, 0, 0, time.UTC), listDefault},
	}
	for _, c := range cases {
		if got, _ := activeList(c.date); got != c.want {
			t.Errorf("%s: got %s want %s", c.date.Format("2006-01-02"), got, c.want)
		}
	}
}

func TestRotationAndMaintenance(t *testing.T) {
	r := &rotator{indices: make(map[string]int)}
	day := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)

	for i := 0; i < len(defaultActivities); i++ {
		got := r.compute(day)
		want := defaultActivities[i]
		if got.name != want.name || got.kind != want.kind || got.status != discord.OnlineStatusOnline {
			t.Fatalf("tick %d: got %+v want %+v", i, got, want)
		}
	}
	if got := r.compute(day); got.name != defaultActivities[0].name {
		t.Fatalf("expected wrap to first entry, got %q", got.name)
	}

	maintenance.Set(true)
	m := r.compute(day)
	if m.status != discord.OnlineStatusDND || m.name != "Regarde la maintenance" || m.kind != watching {
		t.Fatalf("maintenance presence wrong: %+v", m)
	}
	if m2 := r.compute(day); m2 != m {
		t.Fatalf("maintenance presence not stable: %+v vs %+v", m, m2)
	}
	maintenance.Set(false)
	if got := r.compute(day); got.name != defaultActivities[1].name {
		t.Fatalf("rotation did not resume where it left off, got %q", got.name)
	}
}
