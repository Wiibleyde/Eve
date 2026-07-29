package calendar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
)

const sampleICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Eve//Test//EN
BEGIN:VEVENT
UID:utc-1
SUMMARY:Réunion UTC
DTSTART:20260729T140000Z
DTEND:20260729T150000Z
LOCATION:Salle A
END:VEVENT
BEGIN:VEVENT
UID:tz-1
SUMMARY:Réunion Paris
DTSTART;TZID=Europe/Paris:20260730T090000
DURATION:PT1H30M
END:VEVENT
BEGIN:VEVENT
UID:allday-1
SUMMARY:Jour férié
DTSTART;VALUE=DATE:20260801
END:VEVENT
BEGIN:VEVENT
UID:rrule-1
SUMMARY:Point hebdo
DTSTART:20260731T080000Z
DTEND:20260731T083000Z
RRULE:FREQ=WEEKLY;BYDAY=FR
END:VEVENT
BEGIN:VEVENT
UID:escaped-1
SUMMARY:Atelier\, niveau 2
DTSTART:20260802T100000Z
DTEND:20260802T110000Z
END:VEVENT
END:VCALENDAR
`

func parseSample(t *testing.T) []Event {
	t.Helper()
	cal, err := ics.ParseCalendar(strings.NewReader(sampleICS))
	if err != nil {
		t.Fatalf("parsing sample calendar: %v", err)
	}
	return eventsFromCalendar(cal)
}

func findEvent(t *testing.T, events []Event, uid string) Event {
	t.Helper()
	for _, event := range events {
		if event.UID == uid {
			return event
		}
	}
	t.Fatalf("event %q not found", uid)
	return Event{}
}

func TestEventsFromCalendarSortedAndComplete(t *testing.T) {
	events := parseSample(t)
	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}
	for i := 1; i < len(events); i++ {
		if events[i].Start.Before(events[i-1].Start) {
			t.Fatalf("events are not sorted by start time: %v before %v", events[i].Start, events[i-1].Start)
		}
	}
}

func TestUTCEvent(t *testing.T) {
	event := findEvent(t, parseSample(t), "utc-1")
	want := time.Date(2026, 7, 29, 14, 0, 0, 0, time.UTC)
	if !event.Start.Equal(want) {
		t.Fatalf("start = %v, want %v", event.Start, want)
	}
	if !event.End.Equal(want.Add(time.Hour)) {
		t.Fatalf("end = %v, want %v", event.End, want.Add(time.Hour))
	}
	if event.Location != "Salle A" {
		t.Fatalf("location = %q", event.Location)
	}
	if event.AllDay {
		t.Fatal("timed event flagged as all-day")
	}
}

func TestTZIDAndDuration(t *testing.T) {
	event := findEvent(t, parseSample(t), "tz-1")
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Skipf("timezone database unavailable: %v", err)
	}
	want := time.Date(2026, 7, 30, 9, 0, 0, 0, paris)
	if !event.Start.Equal(want) {
		t.Fatalf("start = %v, want %v", event.Start, want)
	}
	if got := event.End.Sub(event.Start); got != 90*time.Minute {
		t.Fatalf("duration = %v, want 1h30m", got)
	}
}

func TestAllDayEvent(t *testing.T) {
	event := findEvent(t, parseSample(t), "allday-1")
	if !event.AllDay {
		t.Fatal("VALUE=DATE event not flagged as all-day")
	}
	if got := event.End.Sub(event.Start); got != 24*time.Hour {
		t.Fatalf("all-day span = %v, want 24h", got)
	}
}

func TestRecurringFlaggedButNotExpanded(t *testing.T) {
	events := parseSample(t)
	event := findEvent(t, events, "rrule-1")
	if !event.Recurring {
		t.Fatal("RRULE event not flagged as recurring")
	}
	occurrences := 0
	for _, candidate := range events {
		if candidate.UID == "rrule-1" {
			occurrences++
		}
	}
	if occurrences != 1 {
		t.Fatalf("recurring event yielded %d occurrences, want 1 (RRULE expansion is out of scope)", occurrences)
	}
}

func TestTextUnescaping(t *testing.T) {
	event := findEvent(t, parseSample(t), "escaped-1")
	if event.Summary != "Atelier, niveau 2" {
		t.Fatalf("summary = %q, want %q", event.Summary, "Atelier, niveau 2")
	}
}

// outlookICS mimics an Exchange feed: a Windows timezone name time.LoadLocation
// cannot resolve, described by the calendar's own VTIMEZONE block.
const outlookICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Microsoft Corporation//Outlook 16.0 MIMEDIR//EN
BEGIN:VTIMEZONE
TZID:Romance Standard Time
BEGIN:STANDARD
DTSTART:16011028T030000
TZOFFSETFROM:+0200
TZOFFSETTO:+0100
END:STANDARD
BEGIN:DAYLIGHT
DTSTART:16010325T020000
TZOFFSETFROM:+0100
TZOFFSETTO:+0200
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VEVENT
UID:outlook-1
SUMMARY:Réunion Outlook
DTSTART;TZID=Romance Standard Time:20260730T090000
DTEND;TZID=Romance Standard Time:20260730T103000
END:VEVENT
BEGIN:VEVENT
UID:unknown-tz-1
SUMMARY:Zone inconnue
DTSTART;TZID=Nowhere Standard Time:20260730T090000
END:VEVENT
END:VCALENDAR
`

func TestWindowsTimezoneResolvedFromVTimezone(t *testing.T) {
	cal, err := ics.ParseCalendar(strings.NewReader(outlookICS))
	if err != nil {
		t.Fatalf("parsing outlook calendar: %v", err)
	}

	events := eventsFromCalendar(cal)
	if len(events) != 1 {
		t.Fatalf("expected the Windows-TZID event to be kept and the undeclared one dropped, got %d events", len(events))
	}

	event := findEvent(t, events, "outlook-1")
	want := time.Date(2026, 7, 30, 8, 0, 0, 0, time.UTC)
	if !event.Start.Equal(want) {
		t.Fatalf("start = %v, want %v", event.Start.UTC(), want)
	}
	if got := event.End.Sub(event.Start); got != 90*time.Minute {
		t.Fatalf("duration = %v, want 1h30m", got)
	}
}

func TestParseUTCOffset(t *testing.T) {
	tests := []struct {
		value string
		want  int
		ok    bool
	}{
		{value: "+0100", want: 3600, ok: true},
		{value: "-0530", want: -(5*3600 + 30*60), ok: true},
		{value: "+010030", want: 3600 + 30, ok: true},
		{value: "0100", ok: false},
		{value: "+01:00", ok: false},
		{value: "+01AB", ok: false},
	}
	for _, test := range tests {
		prop := &ics.IANAProperty{BaseProperty: ics.BaseProperty{Value: test.value}}
		got, ok := parseUTCOffset(prop)
		if ok != test.ok {
			t.Fatalf("parseUTCOffset(%q) ok = %v, want %v", test.value, ok, test.ok)
		}
		if ok && got != test.want {
			t.Fatalf("parseUTCOffset(%q) = %d, want %d", test.value, got, test.want)
		}
	}
	if _, ok := parseUTCOffset(nil); ok {
		t.Fatal("a missing property must not yield an offset")
	}
}

func TestFetchEventsRejectsNonCalendarBody(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "empty body", body: "", wantErr: true},
		{name: "valid calendar", body: sampleICS},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/calendar")
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			events, err := fetchEvents(context.Background(), server.URL)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %d events", len(events))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(events) != 5 {
				t.Fatalf("got %d events, want 5", len(events))
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "https", input: " https://example.com/cal.ics ", want: "https://example.com/cal.ics"},
		{name: "webcal rewritten", input: "webcal://example.com/cal.ics", want: "https://example.com/cal.ics"},
		{name: "empty", input: "   ", wantErr: true},
		{name: "bad scheme", input: "file:///etc/passwd", wantErr: true},
		{name: "no host", input: "https://", wantErr: true},
		{name: "too long", input: "https://example.com/" + strings.Repeat("a", maxURLLength), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeURL(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestParseICSDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		ok    bool
	}{
		{input: "PT1H", want: time.Hour, ok: true},
		{input: "P1DT2H30M", want: 26*time.Hour + 30*time.Minute, ok: true},
		{input: "P2W", want: 14 * 24 * time.Hour, ok: true},
		{input: "-PT15M", want: -15 * time.Minute, ok: true},
		{input: "", ok: false},
		{input: "garbage", ok: false},
	}
	for _, test := range tests {
		got, ok := parseICSDuration(test.input)
		if ok != test.ok {
			t.Fatalf("parseICSDuration(%q) ok = %v, want %v", test.input, ok, test.ok)
		}
		if ok && got != test.want {
			t.Fatalf("parseICSDuration(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}
