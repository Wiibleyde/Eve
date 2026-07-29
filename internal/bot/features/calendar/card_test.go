package calendar

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"Eve/internal/bot/ui"

	"github.com/disgoorg/disgo/discord"
)

func TestSplitEvents(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{Summary: "passé", Start: now.Add(-3 * time.Hour), End: now.Add(-2 * time.Hour)},
		{Summary: "en cours", Start: now.Add(-time.Hour), End: now.Add(time.Hour)},
		{Summary: "à venir 1", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
		{Summary: "à venir 2", Start: now.Add(4 * time.Hour), End: now.Add(5 * time.Hour)},
	}

	ongoing, upcoming := splitEvents(events, now)
	if len(ongoing) != 1 || ongoing[0].Summary != "en cours" {
		t.Fatalf("ongoing = %+v", ongoing)
	}
	if len(upcoming) != 2 {
		t.Fatalf("upcoming = %+v", upcoming)
	}
}

func TestSplitEventsBoundsAreInclusive(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{Summary: "se termine maintenant", Start: now.Add(-time.Hour), End: now},
		{Summary: "commence maintenant", Start: now, End: now.Add(time.Hour)},
	}

	ongoing, upcoming := splitEvents(events, now)
	if len(ongoing) != 2 {
		t.Fatalf("ongoing = %+v, want both events (start ≤ now ≤ end)", ongoing)
	}
	if len(upcoming) != 0 {
		t.Fatalf("upcoming = %+v, want none", upcoming)
	}
}

func TestSplitEventsCapsUpcoming(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	events := make([]Event, 0, 12)
	for i := 1; i <= 12; i++ {
		start := now.Add(time.Duration(i) * time.Hour)
		events = append(events, Event{Summary: fmt.Sprintf("event %d", i), Start: start, End: start.Add(time.Hour)})
	}

	_, upcoming := splitEvents(events, now)
	if len(upcoming) != maxUpcoming {
		t.Fatalf("upcoming length = %d, want %d", len(upcoming), maxUpcoming)
	}
}

func TestBuildCardEmpty(t *testing.T) {
	joined := strings.Join(ui.Texts(buildCard(nil, time.Now()).Components()), "\n")
	if !strings.Contains(joined, msgNoEvents) {
		t.Fatalf("card = %q, want it to contain %q", joined, msgNoEvents)
	}
	if strings.Contains(joined, sectionOngoing) || strings.Contains(joined, sectionUpcoming) {
		t.Fatalf("expected no section headings, got %q", joined)
	}
}

func TestBuildCardSections(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{Summary: "En réunion", Start: now.Add(-time.Hour), End: now.Add(time.Hour), Location: "Salle A"},
		{Summary: "Plus tard", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)},
	}

	texts := ui.Texts(buildCard(events, now).Components())
	joined := strings.Join(texts, "\n")

	for _, want := range []string{
		"### " + sectionOngoing,
		"### " + sectionUpcoming,
		"📍 Salle A",
		fmt.Sprintf("<t:%d:R>", events[1].Start.Unix()),
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("card is missing %q, got:\n%s", want, joined)
		}
	}
	for _, text := range texts {
		if len(text) > maxSectionLength {
			t.Fatalf("text block exceeds the section limit: %d", len(text))
		}
	}
}

func TestRenderLinesRespectsFieldLimit(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	events := make([]Event, 0, 20)
	for i := range 20 {
		events = append(events, Event{
			Summary:  strings.Repeat("x", maxSummaryLength),
			Location: strings.Repeat("y", maxLocationLength),
			Start:    now.Add(time.Duration(i) * time.Hour),
		})
	}
	if got := len(renderLines(events)); got > maxSectionLength {
		t.Fatalf("rendered %d characters, limit is %d", got, maxSectionLength)
	}
}

func TestEventsStartingWithin(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{Summary: "déjà commencé", Start: now.Add(-time.Minute)},
		{Summary: "dans 10 min", Start: now.Add(10 * time.Minute)},
		{Summary: "dans 30 min", Start: now.Add(30 * time.Minute)},
		{Summary: "dans 31 min", Start: now.Add(31 * time.Minute)},
	}

	due := eventsStartingWithin(events, now, scheduledEventLead)
	if len(due) != 2 {
		t.Fatalf("due = %+v", due)
	}
	if due[0].Summary != "dans 10 min" || due[1].Summary != "dans 30 min" {
		t.Fatalf("unexpected due events: %+v", due)
	}
}

func TestScheduledEventExists(t *testing.T) {
	start := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	event := Event{Summary: "Réunion", Start: start}

	existing := []discord.GuildScheduledEvent{
		{Name: "Autre", ScheduledStartTime: start},
		{Name: "Réunion", ScheduledStartTime: start.Add(20 * time.Second).In(time.FixedZone("CEST", 2*3600))},
	}
	if !scheduledEventExists(existing, event) {
		t.Fatal("expected the existing scheduled event to be detected")
	}

	other := Event{Summary: "Réunion", Start: start.Add(time.Hour)}
	if scheduledEventExists(existing, other) {
		t.Fatal("a different start time must not be treated as a duplicate")
	}
}

func TestScheduledEventCreateDefaults(t *testing.T) {
	start := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	create := scheduledEventCreate(Event{Summary: "Réunion", Start: start})

	if create.EntityType != discord.ScheduledEventEntityTypeExternal {
		t.Fatalf("entity type = %d", create.EntityType)
	}
	if create.EntityMetaData == nil || create.EntityMetaData.Location != defaultEventLocation {
		t.Fatalf("entity metadata = %+v", create.EntityMetaData)
	}
	if create.ScheduledEndTime == nil || !create.ScheduledEndTime.After(create.ScheduledStartTime) {
		t.Fatalf("end time = %v", create.ScheduledEndTime)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("éèêë", 10); got != "éèêë" {
		t.Fatalf("got %q", got)
	}
	got := truncate("éèêë", 3)
	if len([]rune(got)) != 3 {
		t.Fatalf("got %q (%d runes), want 3 runes", got, len([]rune(got)))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("got %q, want an ellipsis suffix", got)
	}
}
