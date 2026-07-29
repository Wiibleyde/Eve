package motus

import (
	"strings"
	"unicode"
)

const (
	// MaxAttempts is the number of guesses a board allows before it is lost.
	MaxAttempts = 6
	// MinWordLen / MaxWordLen bound the bundled fallback list only (used by
	// IsPlayable); the external API is not length-bounded, per spec.
	MinWordLen = 6
	MaxWordLen = 9
)

// LetterState is the Motus result for one letter of a guess.
type LetterState int

const (
	// LetterAbsent means the letter is not in the word (or not anymore, once
	// the duplicates have been consumed by better-placed occurrences).
	LetterAbsent LetterState = iota
	// LetterMisplaced means the letter is in the word, at another position.
	LetterMisplaced
	// LetterFound means the letter is at the right position.
	LetterFound
)

// Motus convention: red square = well placed, yellow circle = misplaced,
// blue square = absent.
const (
	emojiFound     = "🟥"
	emojiMisplaced = "🟡"
	emojiAbsent    = "🟦"
)

// Legend is the emoji key printed under every board.
const Legend = emojiFound + " bien placée · " + emojiMisplaced + " mal placée · " + emojiAbsent + " absente"

// ligatures are expanded before the diacritics are dropped, because they carry
// two letters each.
var ligatures = map[rune]string{
	'Œ': "OE",
	'Æ': "AE",
}

// diacritics maps every accented uppercase letter used in French (plus a few
// neighbours) to its bare ASCII counterpart.
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

// Normalize turns any user or API supplied word into the canonical board form:
// trimmed, uppercase, ligatures expanded, diacritics and combining marks
// dropped. It deliberately keeps every other character (spaces, digits,
// hyphens) so that IsWord can reject them with a proper error message instead
// of silently "fixing" a wrong guess.
func Normalize(s string) string {
	upper := strings.ToUpper(strings.TrimSpace(s))

	var b strings.Builder
	b.Grow(len(upper))
	for _, r := range upper {
		// Unicode combining marks (decomposed input such as "e" + U+0301).
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

// IsWord reports whether s is made of nothing but A-Z characters.
// It expects an already normalized string.
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

// IsPlayable reports whether a normalized word can be used as a Motus answer.
func IsPlayable(s string) bool {
	if !IsWord(s) {
		return false
	}
	n := len([]rune(s))
	return n >= MinWordLen && n <= MaxWordLen
}

// Score grades a guess against the answer using the standard two-pass
// algorithm: exact positions are claimed first, then the leftover letters are
// matched against the remaining (unclaimed) occurrences so that duplicated
// letters are never over-reported.
//
// Both arguments must be normalized and of equal length; a length mismatch
// yields an all-absent row rather than a panic.
func Score(guess, answer string) []LetterState {
	g := []rune(guess)
	a := []rune(answer)

	states := make([]LetterState, len(g))
	if len(g) != len(a) {
		return states
	}

	// Pass 1: exact matches, and count what is left of the answer.
	remaining := make(map[rune]int, len(a))
	for i, r := range a {
		if g[i] == r {
			states[i] = LetterFound
			continue
		}
		remaining[r]++
	}

	// Pass 2: misplaced letters, respecting the remaining counts.
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

// IsWinning reports whether every letter of the row was found.
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

// RenderStates turns a scored row into its emoji line. Cells are separated by
// a space to stay column-aligned with RenderLetters rows.
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

// RenderLetters renders a word as regional-indicator emojis, which share the
// emoji cell width and therefore align with the RenderStates row above them.
// The separating spaces are load-bearing: two adjacent regional indicators
// merge into a country flag. Non-letters (the dots of MaskWord) render as ⬛.
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

// MaskWord builds the initial board row: the first letter is revealed (classic
// Motus), the rest is hidden behind dots.
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

// FirstLetter returns the revealed first letter of the answer.
func FirstLetter(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[0])
}

// WordLen returns the length of a word in letters (not bytes).
func WordLen(word string) int {
	return len([]rune(word))
}
