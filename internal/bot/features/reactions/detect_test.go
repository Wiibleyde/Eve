package reactions

import (
	"strings"
	"testing"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func TestMatchDetector(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		detector string
	}{
		{"quoi", "quoi", "quoi"},
		{"koa", "koa", "quoi"},
		{"qoa", "qoa", "quoi"},
		{"koi", "koi", "quoi"},
		{"kwa", "kwa", "quoi"},
		{"kewa", "kewa", "quoi"},

		{"repeated letters", "quoiiii", "quoi"},
		{"repeated middle letters", "quuuoooiii", "quoi"},
		{"spec example", "quoiiii ??", "quoi"},
		{"uppercase", "QUOI", "quoi"},
		{"mixed case", "QuOi ?", "quoi"},
		{"trailing question mark", "quoi?", "quoi"},
		{"trailing spaces", "quoi   ", "quoi"},
		{"trailing plus", "quoi+++", "quoi"},
		{"space then question marks", "quoi ???", "quoi"},
		{"end of sentence", "tu as dit quoi ?", "quoi"},
		{"koa repeated", "kooooaaa ?", "quoi"},
		{"kewa repeated", "kkkeeewwwaaa ??", "quoi"},
		{"kwa repeated", "kwaaa", "quoi"},

		{"comment", "comment", "comment"},
		{"comment repeated", "commenttt ?", "comment"},
		{"komen", "komen", "comment"},
		{"komen repeated", "komennn ???", "comment"},
		{"comment in sentence", "mais tu fais comment ?", "comment"},
		{"comment uppercase", "COMMENT ?", "comment"},

		{"empty", "", ""},
		{"blank", "   ", ""},
		{"not at the end", "quoi de neuf docteur", ""},
		{"comment not at the end", "comment vas-tu aujourd'hui", ""},
		{"unrelated", "bonjour tout le monde", ""},
		{"prefix only", "qu", ""},
		{"commentaire", "un commentaire", ""},
		{"url", "https://example.com/koala", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, ok := match(tt.content)
			if tt.detector == "" {
				if ok {
					t.Fatalf("match(%q) matched detector %q, want no match", tt.content, d.name)
				}
				return
			}
			if !ok {
				t.Fatalf("match(%q) got no match, want detector %q", tt.content, tt.detector)
			}
			if d.name != tt.detector {
				t.Errorf("match(%q) = %q, want %q", tt.content, d.name, tt.detector)
			}
		})
	}
}

func TestDetectReturnsKnownReply(t *testing.T) {
	for _, tt := range []struct {
		content  string
		detector string
	}{
		{"quoiiii ??", "quoi"},
		{"comment ?", "comment"},
	} {
		d := detectorByName(t, tt.detector)
		allowed := map[string]bool{d.fallback: true}
		for _, r := range d.responses {
			allowed[r.text] = true
		}

		for range 200 {
			reply, ok := Detect(tt.content)
			if !ok {
				t.Fatalf("Detect(%q) = false, want a match", tt.content)
			}
			if !allowed[reply] {
				t.Fatalf("Detect(%q) = %q, not in detector %q responses", tt.content, reply, d.name)
			}
		}
	}
}

func TestDetectSilentOnNoMatch(t *testing.T) {
	reply, ok := Detect("bonjour")
	if ok || reply != "" {
		t.Fatalf("Detect(\"bonjour\") = (%q, %v), want (\"\", false)", reply, ok)
	}
}

func TestWeightsSumTo100(t *testing.T) {
	for _, d := range detectors {
		total := 0
		for _, r := range d.responses {
			if r.weight <= 0 {
				t.Errorf("detector %q: response %q has non-positive weight %d", d.name, r.text, r.weight)
			}
			total += r.weight
		}
		if total != 100 {
			t.Errorf("detector %q: weights sum to %d, want 100", d.name, total)
		}
		if d.fallback == "" {
			t.Errorf("detector %q: empty fallback", d.name)
		}
	}
}

func TestPickDistribution(t *testing.T) {
	d := detectorByName(t, "quoi")
	const draws = 100_000

	counts := map[string]int{}
	for range draws {
		counts[d.pick()]++
	}

	for _, r := range d.responses {
		share := float64(counts[r.text]) / float64(draws) * 100
		want := float64(r.weight)
		if share < want-3 || share > want+3 {
			t.Errorf("response %q share = %.2f%%, want ~%.0f%%", r.text, share, want)
		}
	}
}

func TestPickFallback(t *testing.T) {
	d := &detector{name: "test", fallback: "Feur."}
	if got := d.pick(); got != "Feur." {
		t.Errorf("pick() on empty table = %q, want fallback %q", got, d.fallback)
	}

	d = &detector{name: "test", responses: []response{{text: "x", weight: 0}}, fallback: "Feur."}
	if got := d.pick(); got != "Feur." {
		t.Errorf("pick() with zero weights = %q, want fallback %q", got, d.fallback)
	}
}

func TestDetectDoesNotPanicOnOddInput(t *testing.T) {
	inputs := []string{
		strings.Repeat("quoi", 500),
		strings.Repeat("k", 2000),
		"\x00\x01�",
		"quoi\n\n",
		"🐨🐨🐨",
		"quoi​",
	}
	for _, in := range inputs {
		if _, _ = Detect(in); false {
			t.Fatal("unreachable")
		}
	}
}

func detectorByName(t *testing.T, name string) *detector {
	t.Helper()
	for _, d := range detectors {
		if d.name == name {
			return d
		}
	}
	t.Fatalf("detector %q not found", name)
	return nil
}

func TestCooldown(t *testing.T) {
	const channelA = 1
	const channelB = 2

	c := newCooldownMap(30 * time.Second)
	now := time.Now()

	if !c.allow(channelA, now) {
		t.Fatal("first message in a channel must be allowed")
	}
	if c.allow(channelA, now.Add(time.Second)) {
		t.Error("second message inside the window must be suppressed")
	}
	if c.allow(channelA, now.Add(29*time.Second)) {
		t.Error("message just before the window ends must be suppressed")
	}
	if !c.allow(channelB, now) {
		t.Error("other channels must not share a cooldown")
	}
	if !c.allow(channelA, now.Add(30*time.Second)) {
		t.Error("message after the window must be allowed again")
	}
}

func TestCooldownPrune(t *testing.T) {
	c := newCooldownMap(30 * time.Second)
	now := time.Now()

	for id := range 100 {
		c.allow(snowflake.ID(id), now)
	}
	if len(c.last) != 100 {
		t.Fatalf("len(last) = %d, want 100", len(c.last))
	}

	c.allow(snowflake.ID(1000), now.Add(time.Hour))
	if len(c.last) != 1 {
		t.Errorf("len(last) = %d after prune, want 1", len(c.last))
	}
}

func TestParseBool(t *testing.T) {
	tests := map[string]bool{
		"true":  true,
		"TRUE":  true,
		" true": true,
		"1":     true,
		"oui":   true,
		"yes":   true,
		"on":    true,
		"false": false,
		"0":     false,
		"non":   false,
		"":      false,
		"maybe": false,
	}
	for raw, want := range tests {
		if got := parseBool(raw); got != want {
			t.Errorf("parseBool(%q) = %v, want %v", raw, got, want)
		}
	}
}
