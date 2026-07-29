package motus

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"épée", "EPEE"},
		{"  Bateau ", "BATEAU"},
		{"MAISON", "MAISON"},
		{"cœur", "COEUR"},
		{"Noël", "NOEL"},
		{"ça", "CA"},
		// Decomposed input: "e" + U+0301 combining acute accent.
		{"épée", "EPEE"},
		// Non-letters are kept so validation can reject them explicitly.
		{"mot us", "MOT US"},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsWord(t *testing.T) {
	for _, s := range []string{"BATEAU", "A"} {
		if !IsWord(s) {
			t.Errorf("IsWord(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "MOT US", "MOT-US", "M0TUS", "épée"} {
		if IsWord(s) {
			t.Errorf("IsWord(%q) = true, want false", s)
		}
	}
}

func TestIsPlayable(t *testing.T) {
	for _, s := range []string{"BATEAU", "ESCARGOT", "LIBELLULE"} {
		if !IsPlayable(s) {
			t.Errorf("IsPlayable(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"CHAT", "ANTICONSTITUTIONNEL", "MOT US"} {
		if IsPlayable(s) {
			t.Errorf("IsPlayable(%q) = true, want false", s)
		}
	}
}

func statesString(states []LetterState) string {
	out := make([]byte, len(states))
	for i, s := range states {
		switch s {
		case LetterFound:
			out[i] = 'F'
		case LetterMisplaced:
			out[i] = 'M'
		default:
			out[i] = '.'
		}
	}
	return string(out)
}

func TestScore(t *testing.T) {
	cases := []struct{ guess, answer, want string }{
		{"BATEAU", "BATEAU", "FFFFFF"},
		{"MAISON", "BATEAU", ".F...."},
		{"CAMION", "MAISON", ".FMMFF"},
		{"SALADE", "MALADE", ".FFFFF"},
		// Duplicated letters must never be over-reported: every E of the answer
		// is already claimed by an exact match, so the extra E stay absent.
		{"EPEEEE", "EPEISE", "FFF..F"},
		// Misplaced duplicates are consumed exactly once each.
		{"AAAABB", "BBAAAA", "MMFFMM"},
		{"ELEVES", "SEVREE", "M.MMFM"},
	}
	for _, c := range cases {
		if got := statesString(Score(c.guess, c.answer)); got != c.want {
			t.Errorf("Score(%q, %q) = %q, want %q", c.guess, c.answer, got, c.want)
		}
	}
}

func TestScoreLengthMismatch(t *testing.T) {
	got := Score("ABC", "ABCDEF")
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if statesString(got) != "..." {
		t.Errorf("mismatched lengths should score all absent, got %q", statesString(got))
	}
}

func TestIsWinning(t *testing.T) {
	if !IsWinning(Score("BATEAU", "BATEAU")) {
		t.Error("identical guess should win")
	}
	if IsWinning(Score("BATEAU", "MAISON")) {
		t.Error("wrong guess should not win")
	}
	if IsWinning(nil) {
		t.Error("empty row should not win")
	}
}

// A guess typed with accents must be able to win against the normalized answer.
func TestAccentedGuessWins(t *testing.T) {
	answer := Normalize("épée")
	guess := Normalize("Épée")
	if !IsWinning(Score(guess, answer)) {
		t.Errorf("accented guess %q should match answer %q", guess, answer)
	}
}

func TestMaskWord(t *testing.T) {
	if got := MaskWord("BATEAU"); got != "B....." {
		t.Errorf("MaskWord(BATEAU) = %q", got)
	}
	if got := MaskWord(""); got != "" {
		t.Errorf("MaskWord(\"\") = %q", got)
	}
}

func TestRenderLetters(t *testing.T) {
	if got := RenderLetters("MOT"); got != "🇲 🇴 🇹" {
		t.Errorf("RenderLetters(MOT) = %q", got)
	}
	if got := RenderLetters("B.."); got != "🇧 ⬛ ⬛" {
		t.Errorf("RenderLetters(B..) = %q", got)
	}
}

func TestRenderStates(t *testing.T) {
	got := RenderStates([]LetterState{LetterFound, LetterMisplaced, LetterAbsent})
	if got != emojiFound+" "+emojiMisplaced+" "+emojiAbsent {
		t.Errorf("RenderStates = %q", got)
	}
}
