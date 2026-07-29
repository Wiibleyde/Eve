package motus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Eve/internal/logger"
)

const (
	// DefaultAPIURL is the external French word API used by the TS version.
	// It answers with a JSON array of objects carrying a "name" field.
	DefaultAPIURL = "https://trouve-mot.fr/api/random"
	// EnvAPIURL optionally overrides DefaultAPIURL.
	EnvAPIURL = "MOTUS_API_URL"
	// EnvWordsFile optionally overrides the bundled fallback list path.
	EnvWordsFile = "MOTUS_WORDS_FILE"
	// DefaultWordsPath is the bundled fallback list, relative to the working
	// directory (same convention as assets/img and assets/fonts).
	DefaultWordsPath = "assets/motus/words.txt"

	apiTimeout     = 3 * time.Second
	apiTries       = 3
	apiMaxBodySize = 256 << 10
)

// apiWord is one entry of the trouve-mot.fr response.
type apiWord struct {
	Name string `json:"name"`
}

// emergencyWords keeps the feature alive even when both the API and the
// bundled list are unreachable (bad working directory, missing asset...).
var emergencyWords = []string{
	"BATEAU", "MAISON", "JARDIN", "VOYAGE", "CAMION", "BUREAU", "CHEVAL",
	"ORANGE", "BANANE", "SOLEIL", "MIROIR", "FENETRE", "CHAPEAU", "MANTEAU",
	"FROMAGE", "VOITURE", "CUISINE", "CHAMBRE", "ARMOIRE", "TABLEAU",
	"MONTAGNE", "RIVIERE", "VILLAGE", "ESCALIER", "PISCINE", "JOURNAL",
	"LECTURE", "HISTOIRE", "SCIENCE", "MUSIQUE", "LUMIERE", "COURAGE",
	"BONHEUR", "FAMILLE", "SEMAINE", "MINUTE", "HORLOGE", "PAPILLON",
	"ESCARGOT", "ECUREUIL",
}

var (
	fallbackOnce  sync.Once
	fallbackWords []string
)

// FallbackWords returns the bundled word list, loading it on first use.
func FallbackWords() []string {
	fallbackOnce.Do(loadFallbackWords)
	return fallbackWords
}

func loadFallbackWords() {
	paths := make([]string, 0, 3)
	if custom := strings.TrimSpace(os.Getenv(EnvWordsFile)); custom != "" {
		paths = append(paths, custom)
	}
	paths = append(paths, DefaultWordsPath)
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), DefaultWordsPath))
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		words := ParseWordList(string(data))
		if len(words) == 0 {
			logger.Warn("Motus fallback word list is empty", "path", path)
			continue
		}
		fallbackWords = words
		logger.Debug("Motus fallback word list loaded", "path", path, "count", len(words))
		return
	}

	logger.Warn("Motus fallback word list not found, using the built-in emergency list",
		"path", DefaultWordsPath, "count", len(emergencyWords))
	fallbackWords = emergencyWords
}

// ParseWordList normalizes a word list file and keeps only playable entries.
// Blank lines and "#" comments are ignored.
func ParseWordList(content string) []string {
	lines := strings.Split(content, "\n")
	words := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		word := Normalize(line)
		if !IsPlayable(word) {
			continue
		}
		if _, dup := seen[word]; dup {
			continue
		}
		seen[word] = struct{}{}
		words = append(words, word)
	}
	return words
}

// PickWord returns a playable answer: the external API first, the bundled list
// as a fallback. It returns "" only if every source is empty, which cannot
// happen while emergencyWords is non-empty.
//
// The chosen word is never returned through logs — callers must only ever log
// it at DEBUG level.
func PickWord(ctx context.Context) string {
	for try := 1; try <= apiTries; try++ {
		word, err := fetchWordFromAPI(ctx)
		if err == nil {
			logger.Debug("Motus word fetched from the API", "try", try)
			return word
		}
		logger.Warn("Motus word API unavailable", "try", try, "error", err)
		if ctx.Err() != nil {
			break
		}
	}

	words := FallbackWords()
	if len(words) == 0 {
		return ""
	}
	logger.Debug("Motus word taken from the fallback list", "count", len(words))
	return words[rand.IntN(len(words))]
}

// fetchWordFromAPI queries the external API once and returns the first playable
// word of the response.
func fetchWordFromAPI(ctx context.Context) (string, error) {
	url := strings.TrimSpace(os.Getenv(EnvAPIURL))
	if url == "" {
		url = DefaultAPIURL
	}

	reqCtx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling word API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("word API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, apiMaxBodySize))
	if err != nil {
		return "", fmt.Errorf("reading word API response: %w", err)
	}

	word, err := parseAPIResponse(body)
	if err != nil {
		return "", err
	}
	return word, nil
}

// parseAPIResponse accepts both the documented array form and a bare object,
// and returns the first entry that is a word once normalized. Unlike the
// bundled list, the API is not length-bounded: the spec accepts "whatever the
// source returns (typically 6-9)".
func parseAPIResponse(body []byte) (string, error) {
	var list []apiWord
	if err := json.Unmarshal(body, &list); err != nil {
		var single apiWord
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			return "", fmt.Errorf("decoding word API response: %w", err)
		}
		list = []apiWord{single}
	}

	for _, entry := range list {
		word := Normalize(entry.Name)
		if IsWord(word) {
			return word, nil
		}
	}
	return "", errors.New("word API returned no playable word")
}
