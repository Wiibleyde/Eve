package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"Eve/internal/logger"
)

const (
	EnvURL     = "OLLAMA_URL"
	EnvModel   = "OLLAMA_MODEL"
	EnvThreads = "OLLAMA_NUM_THREADS"
)

const (
	DefaultModel   = "llama3.2:1b"
	DefaultThreads = 3
	RequestTimeout = 90 * time.Second
	KeepAlive      = "24h"
	MaxQueued      = 3
	numPredict     = 180
	numCtx         = 4096
	temperature    = 0.6
	topP           = 0.9
	repeatPenalty  = 1.15
)

var (
	ErrDisabled = errors.New("ollama: " + EnvURL + " is not configured")
	ErrBusy     = errors.New("ollama: too many pending generations")
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	baseURL string
	model   string
	threads int
	http    *http.Client
	slot    chan struct{}
	queued  atomic.Int64
}

type options struct {
	NumPredict    int      `json:"num_predict"`
	NumCtx        int      `json:"num_ctx"`
	NumThread     int      `json:"num_thread,omitempty"`
	Temperature   float64  `json:"temperature"`
	TopP          float64  `json:"top_p"`
	RepeatPenalty float64  `json:"repeat_penalty"`
	Stop          []string `json:"stop,omitempty"`
}

type chatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	KeepAlive string    `json:"keep_alive"`
	Options   options   `json:"options"`
}

type chatResponse struct {
	Message Message `json:"message"`
	Error   string  `json:"error"`
}

var (
	defaultOnce   sync.Once
	defaultClient *Client
)

func Default() *Client {
	defaultOnce.Do(func() {
		baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv(EnvURL)), "/")
		if baseURL == "" {
			logger.Warn("Ollama disabled: " + EnvURL + " is not set")
			return
		}
		defaultClient = New(baseURL, modelFromEnv(), threadsFromEnv())
		logger.Info("Ollama configured",
			"url", defaultClient.baseURL,
			"model", defaultClient.model,
			"threads", defaultClient.threads,
		)
	})
	return defaultClient
}

func New(baseURL string, model string, threads int) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		threads: threads,
		http:    &http.Client{Timeout: RequestTimeout},
		slot:    make(chan struct{}, 1),
	}
}

func Enabled() bool { return Default() != nil }

func (client *Client) Model() string { return client.model }

func modelFromEnv() string {
	if model := strings.TrimSpace(os.Getenv(EnvModel)); model != "" {
		return model
	}
	return DefaultModel
}

func threadsFromEnv() int {
	raw := strings.TrimSpace(os.Getenv(EnvThreads))
	if raw == "" {
		return DefaultThreads
	}
	threads, err := strconv.Atoi(raw)
	if err != nil || threads < 1 {
		logger.Warn("Ignoring invalid "+EnvThreads, "value", raw)
		return DefaultThreads
	}
	return threads
}

func (client *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	if client == nil {
		return "", ErrDisabled
	}
	release, err := client.acquire(ctx)
	if err != nil {
		return "", err
	}
	defer release()

	body, err := json.Marshal(chatRequest{
		Model:     client.model,
		Messages:  messages,
		Stream:    false,
		KeepAlive: KeepAlive,
		Options: options{
			NumPredict:    numPredict,
			NumCtx:        numCtx,
			NumThread:     client.threads,
			Temperature:   temperature,
			TopP:          topP,
			RepeatPenalty: repeatPenalty,
		},
	})
	if err != nil {
		return "", fmt.Errorf("ollama: encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	started := time.Now()
	resp, err := client.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: calling %s: %w", client.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ollama: server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	var decoded chatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 256<<10)).Decode(&decoded); err != nil {
		return "", fmt.Errorf("ollama: decoding response: %w", err)
	}
	if decoded.Error != "" {
		return "", fmt.Errorf("ollama: server error: %s", decoded.Error)
	}

	answer := strings.TrimSpace(decoded.Message.Content)
	if answer == "" {
		return "", errors.New("ollama: model returned an empty answer")
	}

	logger.Debug("Ollama generation done",
		"model", client.model,
		"duration", time.Since(started).String(),
		"length", len(answer),
	)
	return answer, nil
}

func (client *Client) Warmup(ctx context.Context) error {
	if client == nil {
		return ErrDisabled
	}
	_, err := client.Chat(ctx, []Message{{Role: RoleUser, Content: "ping"}})
	return err
}

func (client *Client) acquire(ctx context.Context) (func(), error) {
	if client.queued.Add(1) > MaxQueued {
		client.queued.Add(-1)
		return nil, ErrBusy
	}
	select {
	case client.slot <- struct{}{}:
		return func() {
			<-client.slot
			client.queued.Add(-1)
		}, nil
	case <-ctx.Done():
		client.queued.Add(-1)
		return nil, ctx.Err()
	}
}
