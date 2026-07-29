package reactions

import (
	"math/rand/v2"
	"regexp"
	"strings"
)

type response struct {
	text   string
	weight int
}

type detector struct {
	name      string
	stems     []string
	responses []response
	fallback  string

	regexes []*regexp.Regexp
}

var suffixes = []string{
	`[ ?]*\?*$`,
	`[ +]*\?*$`,
}

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
				d.regexes = append(d.regexes, regexp.MustCompile(`(?i)`+stem+suffix))
			}
		}
	}
}

func Detect(content string) (reply string, ok bool) {
	d, ok := match(content)
	if !ok {
		return "", false
	}
	return d.pick(), true
}

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
