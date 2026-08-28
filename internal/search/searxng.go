package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"Eve/internal/config"
	"Eve/internal/logger"
)

const (
	RequestTimeout   = 6 * time.Second
	MaxResults       = 3
	maxSnippetLength = 240
	language         = "fr"
)

var ErrDisabled = errors.New("search: " + config.EnvSearxngURL + " is not configured")

type Result struct {
	Title   string
	Domain  string
	Snippet string
}

type Client struct {
	baseURL string
	http    *http.Client
}

type searxResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

var (
	defaultOnce   sync.Once
	defaultClient *Client
)

func Default() *Client {
	defaultOnce.Do(func() {
		baseURL := config.Get().SearxngURL
		if baseURL == "" {
			logger.Warn("Web search disabled: " + config.EnvSearxngURL + " is not set")
			return
		}
		defaultClient = New(baseURL)
		logger.Info("Web search configured", "url", baseURL)
	})
	return defaultClient
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: RequestTimeout},
	}
}

func Enabled() bool { return Default() != nil }

func (client *Client) Search(ctx context.Context, query string) ([]Result, error) {
	if client == nil {
		return nil, ErrDisabled
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search: empty query")
	}

	endpoint := client.baseURL + "/search?" + url.Values{
		"q":          {query},
		"format":     {"json"},
		"language":   {language},
		"categories": {"general"},
		"safesearch": {"0"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("search: building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: calling %s: %w", client.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("search: searxng returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	var decoded searxResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("search: decoding response: %w", err)
	}

	results := make([]Result, 0, MaxResults)
	for _, raw := range decoded.Results {
		content := collapse(raw.Content)
		if content == "" {
			continue
		}
		results = append(results, Result{
			Title:   collapse(raw.Title),
			Domain:  domainOf(raw.URL),
			Snippet: truncate(content, maxSnippetLength),
		})
		if len(results) == MaxResults {
			break
		}
	}
	return results, nil
}

func domainOf(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(parsed.Hostname(), "www.")
}

func collapse(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}
