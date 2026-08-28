package music

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	bracketNoise    = regexp.MustCompile(`[\(\[][^\)\]]*[\)\]]`)
	nonAlphanumeric = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	artistSuffix    = regexp.MustCompile(`(?i)\s+-\s+topic$`)
)

var ignoredTokens = map[string]struct{}{
	"official": {}, "officiel": {}, "video": {}, "audio": {}, "lyric": {},
	"lyrics": {}, "paroles": {}, "hd": {}, "4k": {}, "mv": {}, "remaster": {},
	"remastered": {}, "feat": {}, "ft": {}, "the": {}, "le": {}, "la": {}, "les": {},
}

func stripBrackets(raw string) string {
	return strings.TrimSpace(bracketNoise.ReplaceAllString(raw, " "))
}

func fold(raw string) string {
	chain := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	folded, _, err := transform.String(chain, raw)
	if err != nil {
		return raw
	}
	return folded
}

func tokenize(raw string) map[string]struct{} {
	lowered := strings.ToLower(fold(stripBrackets(raw)))
	out := make(map[string]struct{})
	for _, part := range nonAlphanumeric.Split(lowered, -1) {
		if part == "" {
			continue
		}
		if _, ignored := ignoredTokens[part]; ignored {
			continue
		}
		out[part] = struct{}{}
	}
	return out
}

func containment(left string, right string) float64 {
	a, b := tokenize(left), tokenize(right)
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	shared := 0
	for token := range a {
		if _, ok := b[token]; ok {
			shared++
		}
	}

	smaller := len(a)
	if len(b) < smaller {
		smaller = len(b)
	}
	return float64(shared) / float64(smaller)
}

func splitArtistTitle(raw string) (string, string) {
	cleaned := stripBrackets(raw)
	for _, separator := range []string{" - ", " – ", " — ", " | "} {
		if left, right, found := strings.Cut(cleaned, separator); found {
			return strings.TrimSpace(left), strings.TrimSpace(right)
		}
	}
	return "", cleaned
}

func cleanArtist(raw string) string {
	return strings.TrimSpace(artistSuffix.ReplaceAllString(raw, ""))
}
