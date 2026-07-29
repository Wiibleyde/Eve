package twitch

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

	"Eve/internal/logger"
)

const (
	HelixBaseURL = "https://api.twitch.tv/helix"
	TokenURL     = "https://id.twitch.tv/oauth2/token"

	MaxIDsPerRequest = 100

	tokenRefreshMargin = 5 * time.Minute

	defaultTimeout = 15 * time.Second

	maxBodySize = 4 << 20
)

var ErrNoToken = errors.New("twitch: token response carried no access_token")

type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: defaultTimeout},
	}
}

type APIError struct {
	StatusCode int
	Endpoint   string
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("twitch: %s returned HTTP %d", e.Endpoint, e.StatusCode)
	}
	return fmt.Sprintf("twitch: %s returned HTTP %d: %s", e.Endpoint, e.StatusCode, e.Message)
}

type response[T any] struct {
	Data []T `json:"data"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}
	return c.refreshLocked(ctx)
}

func (c *Client) invalidate(used string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token == used {
		c.token = ""
		c.expiresAt = time.Time{}
	}
}

func (c *Client) refreshLocked(ctx context.Context) (string, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("twitch: building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("twitch: requesting app token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return "", fmt.Errorf("twitch: reading token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{StatusCode: resp.StatusCode, Endpoint: "/oauth2/token", Message: shortBody(body)}
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("twitch: decoding token response: %w", err)
	}
	if tr.AccessToken == "" {
		return "", ErrNoToken
	}

	ttl := time.Duration(tr.ExpiresIn) * time.Second
	switch {
	case ttl <= 0:
		ttl = time.Hour
	case ttl > tokenRefreshMargin:
		ttl -= tokenRefreshMargin
	default:
		ttl /= 2
	}

	c.token = tr.AccessToken
	c.expiresAt = time.Now().Add(ttl)
	logger.Debug("Twitch app token refreshed", "valid_for", ttl.String())

	return c.token, nil
}

func (c *Client) get(ctx context.Context, endpoint string, query url.Values, out any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	status, body, err := c.doGet(ctx, endpoint, query, token)
	if err != nil {
		return err
	}
	if status == http.StatusUnauthorized {
		logger.Debug("Twitch token rejected, retrying with a fresh one", "endpoint", endpoint)
		c.invalidate(token)
		if token, err = c.accessToken(ctx); err != nil {
			return err
		}
		if status, body, err = c.doGet(ctx, endpoint, query, token); err != nil {
			return err
		}
	}
	if status != http.StatusOK {
		return &APIError{StatusCode: status, Endpoint: endpoint, Message: shortBody(body)}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("twitch: decoding %s response: %w", endpoint, err)
	}
	return nil
}

func (c *Client) doGet(ctx context.Context, endpoint string, query url.Values, token string) (int, []byte, error) {
	u := HelixBaseURL + endpoint
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("twitch: building %s request: %w", endpoint, err)
	}
	req.Header.Set("Client-Id", c.clientID)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("twitch: calling %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return 0, nil, fmt.Errorf("twitch: reading %s response: %w", endpoint, err)
	}
	return resp.StatusCode, body, nil
}

func shortBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

func chunk(values []string, size int) [][]string {
	if size <= 0 || len(values) == 0 {
		return nil
	}
	out := make([][]string, 0, (len(values)+size-1)/size)
	for start := 0; start < len(values); start += size {
		end := min(start+size, len(values))
		out = append(out, values[start:end])
	}
	return out
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
