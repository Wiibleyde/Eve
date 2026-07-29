// Package reactions implements the joke auto-replies to message patterns: the
// classic «quoi ?» → «Feur.» and its siblings.
//
// # Structure
//
// Detection is a pure function, Detect, driven by a table of detectors. Adding
// a joke is a data change in detectors — no new code path. The event glue
// (Attach) is deliberately thin so the interesting part stays unit-testable.
//
// # Matching
//
// Patterns are anchored at the end of the message, case-insensitive, and
// tolerate letter repetition («quoiiii») plus trailing spaces, "+" and "?"
// («quoi ??»). The regexes are ported from the TypeScript version.
//
// # Guards (see listener.go)
//
// Guild messages only, never bots, never when the bot itself is mentioned (the
// AI/talk path owns those), silent during maintenance, one reply per channel
// per 30 seconds, and a per-guild opt-out stored in guild_configs under
// "jokes.disabled".
package reactions

import (
	"math/rand/v2"
	"regexp"
	"strings"
)

// response weights are relative; the tables below sum to 100 so they read as
// percentages.
type response struct {
	text   string
	weight int
}

// detector's stems hold the "meat" of each pattern; the trailing-punctuation
// suffixes are appended at compile time (see suffixes), which keeps the table
// readable and guarantees every pattern is end-of-message anchored.
type detector struct {
	name      string
	stems     []string
	responses []response
	fallback  string

	regexes []*regexp.Regexp // compiled from stems × suffixes at init
}

// suffixes are the two trailing-punctuation variants of the TypeScript
// regexes: spaces/"?" then "?"s, and spaces/"+" then "?"s. Both anchor at the
// end of the message.
var suffixes = []string{
	`[ ?]*\?*$`,
	`[ +]*\?*$`,
}

// detectors is the joke table, checked in order — the first match wins.
var detectors = []*detector{
	{
		name: "quoi",
		stems: []string{
			`qu+o+i+`,
			`ko+a+`,
			`q+o+a+`,
			`ko+i+`,
			`kw+a+`,
			`k+e+w+a+`,
		},
		responses: []response{
			{text: "Feur.", weight: 70},
			{text: "coubeh.", weight: 10},
			{text: "la 🐨", weight: 10},
			{text: "drilatère.", weight: 10},
		},
		fallback: "Feur.",
	},
	{
		name: "comment",
		stems: []string{
			`comment+`,
			`komen+`,
		},
		responses: []response{
			{text: "dant.", weight: 100},
		},
		fallback: "dant.",
	},
}

func init() {
	for _, d := range detectors {
		d.regexes = make([]*regexp.Regexp, 0, len(d.stems)*len(suffixes))
		for _, stem := range d.stems {
			for _, suffix := range suffixes {
				// (?i) rather than relying on the caller: Detect lowercases,
				// but the table must be correct on its own.
				d.regexes = append(d.regexes, regexp.MustCompile(`(?i)`+stem+suffix))
			}
		}
	}
}

// Detect reports whether a message triggers a joke and returns the reply to
// send. It is pure apart from the weighted random pick, and never panics on
// arbitrary user input.
func Detect(content string) (reply string, ok bool) {
	d, ok := match(content)
	if !ok {
		return "", false
	}
	return d.pick(), true
}

// match normalizes the message (lowercase, trimmed) and returns the first
// detector whose patterns match.
func match(content string) (*detector, bool) {
	normalized := strings.ToLower(strings.TrimSpace(content))
	if normalized == "" {
		return nil, false
	}
	for _, d := range detectors {
		for _, re := range d.regexes {
			if re.MatchString(normalized) {
				return d, true
			}
		}
	}
	return nil, false
}

// pick draws one response according to the weights, falling back to the
// detector default when the table is empty or the weights are non-positive.
func (d *detector) pick() string {
	total := 0
	for _, r := range d.responses {
		if r.weight > 0 {
			total += r.weight
		}
	}
	if total <= 0 {
		return d.fallback
	}

	roll := rand.IntN(total)
	cumulative := 0
	for _, r := range d.responses {
		if r.weight <= 0 {
			continue
		}
		cumulative += r.weight
		if roll < cumulative {
			return r.text
		}
	}
	return d.fallback
}
