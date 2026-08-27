package ai

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"Eve/internal/logger"
	"Eve/internal/ollama"
	"Eve/internal/search"
)

const (
	searchTimeout   = 6 * time.Second
	minQueryLength  = 12
	maxQueryLength  = 200
	groundingHeader = "Résultats d'une recherche web faite à l'instant :"
	groundingFooter = `Sers-toi de ces résultats s'ils répondent à la question, et cite la source par son nom de domaine.
S'ils ne répondent pas, ignore-les et dis simplement que tu ne sais pas.`
)

var (
	yearPattern     = regexp.MustCompile(`\b20\d{2}\b`)
	factualPatterns = []string{
		"actualité", "actualités", "news", "aujourd'hui", "hier", "en ce moment",
		"récent", "récente", "récemment", "dernier", "dernière", "derniers", "dernières",
		"qui est", "qui sont", "c'est quoi", "qu'est-ce que", "quest ce que",
		"combien", "quand est", "quand a", "quand sort", "date de",
		"météo", "prix de", "cours de", "score", "résultat de",
		"cherche", "recherche", "google", "version de", "sortie de",
	}
)

func needsSearch(prompt string) bool {
	lowered := strings.ToLower(prompt)
	if len([]rune(lowered)) < minQueryLength {
		return false
	}
	if yearPattern.MatchString(lowered) {
		return true
	}
	for _, pattern := range factualPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}
	return false
}

func groundingMessage(ctx context.Context, prompt string) (ollama.Message, bool) {
	if !search.Enabled() || !needsSearch(prompt) {
		return ollama.Message{}, false
	}

	searchCtx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()

	results, err := search.Default().Search(searchCtx, searchQuery(prompt))
	if err != nil {
		logger.Warn("AI: web search failed, answering without grounding", "error", err)
		return ollama.Message{}, false
	}
	if len(results) == 0 {
		return ollama.Message{}, false
	}

	logger.Debug("AI: grounding the answer", "results", len(results))
	return ollama.Message{Role: ollama.RoleSystem, Content: formatResults(results)}, true
}

func formatResults(results []search.Result) string {
	var b strings.Builder
	b.WriteString(groundingHeader)
	for i, result := range results {
		b.WriteString("\n[")
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString("] ")
		b.WriteString(result.Title)
		if result.Domain != "" {
			b.WriteString(" (")
			b.WriteString(result.Domain)
			b.WriteString(")")
		}
		b.WriteString(" : ")
		b.WriteString(result.Snippet)
	}
	b.WriteString("\n\n")
	b.WriteString(groundingFooter)
	return b.String()
}

func searchQuery(prompt string) string {
	return truncate(collapse(prompt), maxQueryLength)
}
