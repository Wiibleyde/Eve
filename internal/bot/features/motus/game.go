package motus

import (
	"strings"
	"unicode"
)

const (
	MaxAttempts = 6
	MinWordLen  = 6
	MaxWordLen  = 9
)

type LetterState int

const (
	LetterAbsent LetterState = iota
	LetterMisplaced
	LetterFound
)

const (
	emojiFound     = "🟥"
	emojiMisplaced = "🟡"
	emojiAbsent    = "🟦"
)

const Legend = emojiFound + " bien placée · " + emojiMisplaced + " mal placée · " + emojiAbsent + " absente"

var ligatures = map[rune]string{
	'Œ': "OE",
	'Æ': "AE",
}

var diacritics = map[rune]rune{
	'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A',
	'Ç': 'C',
	'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E',
	'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I',
	'Ñ': 'N',
	'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O', 'Ø': 'O',
	'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
	'Ý': 'Y', 'Ÿ': 'Y',
}

func Normalize(s string) string {
	upper := strings.ToUpper(strings.TrimSpace(s))

	var b strings.Builder
	b.Grow(len(upper))
	for _, r := range upper {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if expanded, ok := ligatures[r]; ok {
			b.WriteString(expanded)
			continue
		}
		if base, ok := diacritics[r]; ok {
			b.WriteRune(base)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func IsWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func IsPlayable(s string) bool {
	if !IsWord(s) {
		return false
	}
	n := len([]rune(s))
	return n >= MinWordLen && n <= MaxWordLen
}

func Score(guess, answer string) []LetterState {
	g := []rune(guess)
	a := []rune(answer)

	states := make([]LetterState, len(g))
	if len(g) != len(a) {
		return states
	}

	remaining := make(map[rune]int, len(a))
	for i, r := range a {
		if g[i] == r {
			states[i] = LetterFound
			continue
		}
		remaining[r]++
	}

	for i, r := range g {
		if states[i] == LetterFound {
			continue
		}
		if remaining[r] > 0 {
			states[i] = LetterMisplaced
			remaining[r]--
		}
	}
	return states
}

func IsWinning(states []LetterState) bool {
	if len(states) == 0 {
		return false
	}
	for _, s := range states {
		if s != LetterFound {
			return false
		}
	}
	return true
}

func RenderStates(states []LetterState) string {
	parts := make([]string, 0, len(states))
	for _, s := range states {
		switch s {
		case LetterFound:
			parts = append(parts, emojiFound)
		case LetterMisplaced:
			parts = append(parts, emojiMisplaced)
		default:
			parts = append(parts, emojiAbsent)
		}
	}
	return strings.Join(parts, " ")
}

func RenderLetters(word string) string {
	runes := []rune(word)
	parts := make([]string, 0, len(runes))
	for _, r := range runes {
		if r >= 'A' && r <= 'Z' {
			parts = append(parts, string(rune(0x1F1E6+r-'A')))
		} else {
			parts = append(parts, "⬛")
		}
	}
	return strings.Join(parts, " ")
}

func MaskWord(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	masked := make([]rune, len(runes))
	masked[0] = runes[0]
	for i := 1; i < len(masked); i++ {
		masked[i] = '.'
	}
	return string(masked)
}

func FirstLetter(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[0])
}

func WordLen(word string) int {
	return len([]rune(word))
}
