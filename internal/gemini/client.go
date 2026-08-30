package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"Eve/internal/config"
	"Eve/internal/logger"

	"google.golang.org/genai"
)

const Model = "gemini-2.5-flash"

var (
	ErrDisabled    = errors.New("gemini: " + config.EnvGoogleAPIKey + " is not configured")
	ErrRateLimited = errors.New("gemini: rate limited")
	ErrEmptyAnswer = errors.New("gemini: model returned an empty answer")
	safetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockThresholdBlockNone},
	}
	tools = []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}}
)

type Client struct {
	genai *genai.Client
}

type Chat struct {
	chat *genai.Chat
}

var (
	defaultOnce   sync.Once
	defaultClient *Client
)

func Default() *Client {
	defaultOnce.Do(func() {
		apiKey := config.Get().GoogleAPIKey
		if apiKey == "" {
			logger.Warn("Gemini disabled: " + config.EnvGoogleAPIKey + " is not set")
			return
		}
		client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			logger.Error("Gemini: creating client failed", "error", err)
			return
		}
		defaultClient = &Client{genai: client}
		logger.Info("Gemini configured", "model", Model)
	})
	return defaultClient
}

func Enabled() bool { return Default() != nil }

func (client *Client) NewChat(ctx context.Context, systemInstruction string) (*Chat, error) {
	if client == nil {
		return nil, ErrDisabled
	}

	chat, err := client.genai.Chats.Create(ctx, Model, &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, ""),
		SafetySettings:    safetySettings,
		Tools:             tools,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("gemini: creating chat: %w", err)
	}
	return &Chat{chat: chat}, nil
}

func (chat *Chat) Send(ctx context.Context, message string) (string, error) {
	if chat == nil {
		return "", ErrDisabled
	}

	resp, err := chat.chat.SendMessage(ctx, genai.Part{Text: message})
	if err != nil {
		return "", translateError(err)
	}

	answer := strings.TrimSpace(resp.Text())
	if answer == "" {
		return "", ErrEmptyAnswer
	}
	return answer, nil
}

func translateError(err error) error {
	if code, message, ok := apiError(err); ok && (code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable) {
		return fmt.Errorf("%w: %s", ErrRateLimited, message)
	}
	return fmt.Errorf("gemini: generating response: %w", err)
}

func apiError(err error) (int, string, bool) {
	if byValue, ok := errors.AsType[genai.APIError](err); ok {
		return byValue.Code, byValue.Message, true
	}
	if byPointer, ok := errors.AsType[*genai.APIError](err); ok {
		return byPointer.Code, byPointer.Message, true
	}
	return 0, "", false
}

func (client *Client) CompleteJSON(ctx context.Context, instruction string, prompt string) (string, error) {
	if client == nil {
		return "", ErrDisabled
	}

	resp, err := client.genai.Models.GenerateContent(ctx, Model, genai.Text(prompt), &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(instruction, ""),
		SafetySettings:    safetySettings,
		ResponseMIMEType:  "application/json",
		Temperature:       genai.Ptr[float32](1.0),
	})
	if err != nil {
		return "", translateError(err)
	}

	answer := strings.TrimSpace(resp.Text())
	if answer == "" {
		return "", ErrEmptyAnswer
	}
	return answer, nil
}
