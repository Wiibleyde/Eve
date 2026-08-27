package ai

import (
	"strings"
	"testing"

	"Eve/internal/search"
)

func TestNeedsSearchOnFactualPrompts(t *testing.T) {
	for _, prompt := range []string{
		"c'est quoi le protocole QUIC",
		"quelles sont les actualités du jour",
		"qui est le président du Portugal",
		"quelle est la dernière version de Go",
		"que s'est-il passé en 2025",
	} {
		if !needsSearch(prompt) {
			t.Errorf("needsSearch(%q) = false, want true", prompt)
		}
	}
}

func TestNeedsSearchIgnoresSmallTalk(t *testing.T) {
	for _, prompt := range []string{
		"salut",
		"ça va ?",
		"raconte-moi une blague",
		"merci beaucoup Eve",
	} {
		if needsSearch(prompt) {
			t.Errorf("needsSearch(%q) = true, want false", prompt)
		}
	}
}

func TestFormatResultsNumbersAndAttributes(t *testing.T) {
	got := formatResults([]search.Result{
		{Title: "Go 1.26", Domain: "go.dev", Snippet: "Sortie de Go 1.26."},
		{Title: "Notes", Domain: "", Snippet: "Sans domaine."},
	})

	for _, want := range []string{groundingHeader, "[1] Go 1.26 (go.dev) : Sortie de Go 1.26.", "[2] Notes : Sans domaine.", groundingFooter} {
		if !strings.Contains(got, want) {
			t.Errorf("formatted results are missing %q:\n%s", want, got)
		}
	}
}

func TestSearchQueryIsCapped(t *testing.T) {
	got := searchQuery(strings.Repeat("a", maxQueryLength+50))
	if length := len([]rune(got)); length != maxQueryLength {
		t.Errorf("length = %d, want %d", length, maxQueryLength)
	}
}
