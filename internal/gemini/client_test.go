package gemini

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/genai"
)

func TestNilClientNewChatReturnsErrDisabled(t *testing.T) {
	var client *Client
	if _, err := client.NewChat(context.Background(), "system"); !errors.Is(err, ErrDisabled) {
		t.Errorf("error = %v, want ErrDisabled", err)
	}
}

func TestNilChatSendReturnsErrDisabled(t *testing.T) {
	var chat *Chat
	if _, err := chat.Send(context.Background(), "salut"); !errors.Is(err, ErrDisabled) {
		t.Errorf("error = %v, want ErrDisabled", err)
	}
}

func TestTranslateErrorMapsTooManyRequests(t *testing.T) {
	err := translateError(genai.APIError{Code: 429, Message: "quota exceeded"})
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("error = %v, want it to wrap ErrRateLimited", err)
	}
}

func TestTranslateErrorMapsServiceUnavailable(t *testing.T) {
	err := translateError(&genai.APIError{Code: 503, Message: "model overloaded"})
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("error = %v, want it to wrap ErrRateLimited", err)
	}
}

func TestTranslateErrorPassesThroughOtherErrors(t *testing.T) {
	original := errors.New("boom")
	err := translateError(original)
	if errors.Is(err, ErrRateLimited) {
		t.Error("error wraps ErrRateLimited, want a plain wrapped error")
	}
	if !errors.Is(err, original) {
		t.Errorf("error = %v, want it to wrap the original error", err)
	}
}

func TestTranslateErrorRateLimitedPointer(t *testing.T) {
	err := translateError(&genai.APIError{Code: 429, Message: "quota exceeded"})
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("error = %v, want it to wrap ErrRateLimited", err)
	}
}

func TestTranslateErrorRateLimitedWrapped(t *testing.T) {
	err := translateError(fmt.Errorf("sending message: %w", genai.APIError{Code: 429, Message: "quota exceeded"}))
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("error = %v, want it to wrap ErrRateLimited", err)
	}
}
