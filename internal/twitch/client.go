// Package twitch is a minimal Twitch Helix REST client.
//
// It only covers what the bot needs: resolving channel logins to stable user
// IDs and polling the live state of a set of users.
//
// # Authentication
//
// The client uses the OAuth2 client-credentials ("app access token") flow: no
// user consent, no refresh token. The token is fetched lazily on the first
// call, cached in memory until shortly before its expiry, and renewed
// automatically. A 401 from Helix invalidates the cached token and the request
// is retried once with a fresh one — that covers tokens revoked server-side
// before their announced expiry.
//
// # Batching
//
// Helix accepts a repeated query parameter up to 100 times per request
// (id/login/user_id). Every method here takes a slice, de-duplicates it and
// splits it into 100-value chunks transparently, so callers never have to
// think about the cap.
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
	// HelixBaseURL is the Twitch Helix API root.
	HelixBaseURL = "https://api.twitch.tv/helix"
	// TokenURL is the OAuth2 endpoint used by the client-credentials flow.
	TokenURL = "https://id.twitch.tv/oauth2/token"

	// MaxIDsPerRequest is the Helix cap on repeated id/login/user_id params.
	MaxIDsPerRequest = 100

	// tokenRefreshMargin renews the app token before it actually expires, so a
	// request never races the expiry.
	tokenRefreshMargin = 5 * time.Minute

	// defaultTimeout bounds a single HTTP call. The poller also passes a
	// context deadline; whichever fires first wins.
	defaultTimeout = 15 * time.Second

	// maxBodySize caps how much of a response we read — Helix answers are tiny
	// and an unbounded read on a misbehaving endpoint is a memory hazard.
	maxBodySize = 4 << 20
)

// ErrNoToken is returned when Twitch answers the token endpoint with a body
// that carries no access token.
var ErrNoToken = errors.New("twitch: token response carried no access_token")

// Client is a Helix API client. It is safe for concurrent use.
type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// New builds a client for the given application credentials. No network call
// is made here: the app token is fetched on first use.
func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: defaultTimeout},
	}
}

// APIError is a non-2xx answer from Twitch.
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

// response is the Helix envelope shared by every list endpoint.
type response[T any] struct {
	Data []T `json:"data"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// accessToken returns a valid app access token, fetching a new one when the
// cached one is missing or about to expire.
func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}
	return c.refreshLocked(ctx)
}

// invalidate drops the cached token, but only if it is still the one the caller
// just failed with — otherwise a concurrent refresh would be thrown away.
func (c *Client) invalidate(used string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token == used {
		c.token = ""
		c.expiresAt = time.Time{}
	}
}

// refreshLocked performs the client-credentials exchange. The mutex is held
// across the HTTP call on purpose: it collapses concurrent refreshes into one.
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

// get calls a Helix endpoint and decodes its JSON body into out. A 401 is
// retried once with a freshly minted token.
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

// shortBody trims an error body so it stays readable in a log line.
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

// dedupe removes empty and duplicate values while preserving order.
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
