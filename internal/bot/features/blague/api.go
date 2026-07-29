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

const EnvToken = "BLAGUE_API_TOKEN"

const RequestTimeout = 5 * time.Second

var baseURL = "https://www.blagues-api.fr/api"

var httpClient = &http.Client{Timeout: RequestTimeout}

var ErrDisabled = errors.New("blague: " + EnvToken + " is not configured")

var ErrUnknownCategory = errors.New("blague: unknown category")

type Joke struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Joke   string `json:"joke"`
	Answer string `json:"answer"`
}

type category struct {
	Value string
	Label string
}

var categories = []category{
	{Value: "global", Label: "Générale"},
	{Value: "dev", Label: "Développeur"},
	{Value: "beauf", Label: "Beauf"},
}

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
