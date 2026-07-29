package blague

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"Eve/internal/logger"
)

// EnvToken is the environment variable holding the blagues-api.fr bearer token.
// When it is empty the whole feature disables itself: the command is not even
// registered on Discord (see Commands).
const EnvToken = "BLAGUE_API_TOKEN"

// RequestTimeout caps a single call to blagues-api.fr. It is deliberately
// shorter than Discord's 15 minute followup window but long enough for a slow
// upstream; the interaction is deferred before the call so the 3s ACK deadline
// is never at risk.
const RequestTimeout = 5 * time.Second

// baseURL is a var (not a const) purely so tests can point it at a httptest
// server. Nothing else rewrites it.
var baseURL = "https://www.blagues-api.fr/api"

var httpClient = &http.Client{Timeout: RequestTimeout}

// ErrDisabled is returned when no API token is configured.
var ErrDisabled = errors.New("blague: " + EnvToken + " is not configured")

// ErrUnknownCategory is returned for a category outside the allowed set.
var ErrUnknownCategory = errors.New("blague: unknown category")

// Joke is the blagues-api.fr payload:
//
//	{"id":1,"type":"global","joke":"...","answer":"..."}
type Joke struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Joke   string `json:"joke"`
	Answer string `json:"answer"`
}

// category maps an API category to its French slash command label.
type category struct {
	Value string
	Label string
}

// categories intentionally excludes "dark", "limit" and "blondes" — same
// editorial choice as the TypeScript version.
var categories = []category{
	{Value: "global", Label: "Générale"},
	{Value: "dev", Label: "Développeur"},
	{Value: "beauf", Label: "Beauf"},
}

// categoryLabel resolves an API category to its French label and doubles as the
// allowlist guarding the URL path we build.
func categoryLabel(value string) (string, bool) {
	for _, c := range categories {
		if c.Value == value {
			return c.Label, true
		}
	}
	return "", false
}

var (
	tokenOnce sync.Once
	apiToken  string
)

// token reads EnvToken exactly once, lazily, so godotenv has already populated
// the environment by the time it runs.
func token() string {
	tokenOnce.Do(func() {
		apiToken = strings.TrimSpace(os.Getenv(EnvToken))
		if apiToken == "" {
			logger.Warn("Blague feature disabled: "+EnvToken+" is not set", "command", CommandName)
		}
	})
	return apiToken
}

func Enabled() bool { return token() != "" }

func fetchRandom(ctx context.Context, categoryValue string) (*Joke, error) {
	tok := token()
	if tok == "" {
		return nil, ErrDisabled
	}
	if _, ok := categoryLabel(categoryValue); !ok {
		return nil, ErrUnknownCategory
	}

	url := fmt.Sprintf("%s/type/%s/random", baseURL, categoryValue)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("blague: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blague: calling blagues-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("blague: blagues-api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	var joke Joke
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&joke); err != nil {
		return nil, fmt.Errorf("blague: decoding blagues-api response: %w", err)
	}
	if strings.TrimSpace(joke.Joke) == "" {
		return nil, errors.New("blague: blagues-api returned an empty joke")
	}
	return &joke, nil
}
