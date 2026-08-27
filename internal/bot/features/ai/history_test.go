package ai

import (
	"testing"
	"time"

	"Eve/internal/ollama"
)

func userMessage(content string) ollama.Message {
	return ollama.Message{Role: ollama.RoleUser, Content: content}
}

func TestHistoryKeepsOnlyTheLastMessages(t *testing.T) {
	h := newHistory(time.Hour, 2)
	now := time.Now()

	h.append(1, now, 10, userMessage("a"), userMessage("b"), userMessage("c"))

	got := h.messages(1, now)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Content != "b" || got[1].Content != "c" {
		t.Errorf("messages = %v, want [b c]", got)
	}
}

func TestHistoryIsolatesChannels(t *testing.T) {
	h := newHistory(time.Hour, 4)
	now := time.Now()

	h.append(1, now, 10, userMessage("un"))
	h.append(2, now, 20, userMessage("deux"))

	if got := h.messages(1, now); len(got) != 1 || got[0].Content != "un" {
		t.Errorf("channel 1 = %v, want [un]", got)
	}
	if got := h.messages(2, now); len(got) != 1 || got[0].Content != "deux" {
		t.Errorf("channel 2 = %v, want [deux]", got)
	}
}

func TestHistoryExpiresStaleConversations(t *testing.T) {
	h := newHistory(time.Minute, 4)
	now := time.Now()

	h.append(1, now, 10, userMessage("vieux"))

	if got := h.messages(1, now.Add(2*time.Minute)); got != nil {
		t.Errorf("messages = %v, want nil", got)
	}
}

func TestHistoryTracksParticipants(t *testing.T) {
	h := newHistory(time.Hour, 4)
	now := time.Now()

	h.append(1, now, 10, userMessage("un"))
	h.append(1, now, 11, userMessage("deux"))

	got := h.participants(1, now)
	if len(got) != 2 {
		t.Fatalf("participants = %v, want 2 entries", got)
	}
}

func TestHistoryReturnsACopy(t *testing.T) {
	h := newHistory(time.Hour, 4)
	now := time.Now()
	h.append(1, now, 10, userMessage("original"))

	got := h.messages(1, now)
	got[0].Content = "modifié"

	if again := h.messages(1, now); again[0].Content != "original" {
		t.Errorf("stored message = %q, want %q", again[0].Content, "original")
	}
}
