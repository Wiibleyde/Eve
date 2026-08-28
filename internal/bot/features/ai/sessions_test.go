package ai

import (
	"context"
	"testing"
	"time"

	"Eve/internal/gemini"
)

func newTestStore(ttl time.Duration) (*sessionStore, *int) {
	calls := new(int)
	factory := func(ctx context.Context, systemInstruction string) (*gemini.Chat, error) {
		*calls++
		return &gemini.Chat{}, nil
	}
	return newSessionStore(ttl, factory), calls
}

func TestChatForReusesSessionWithinTTL(t *testing.T) {
	store, calls := newTestStore(time.Hour)
	now := time.Now()

	first, err := store.chatFor(context.Background(), 1, now, "system")
	if err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	second, err := store.chatFor(context.Background(), 1, now, "system")
	if err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	if first != second {
		t.Error("chatFor() returned a new chat instead of reusing the session")
	}
	if *calls != 1 {
		t.Errorf("factory called %d times, want 1", *calls)
	}
}

func TestChatForIsolatesChannels(t *testing.T) {
	store, calls := newTestStore(time.Hour)
	now := time.Now()

	if _, err := store.chatFor(context.Background(), 1, now, "system"); err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	if _, err := store.chatFor(context.Background(), 2, now, "system"); err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	if *calls != 2 {
		t.Errorf("factory called %d times, want 2", *calls)
	}
}

func TestChatForCreatesNewSessionAfterExpiry(t *testing.T) {
	store, _ := newTestStore(time.Minute)
	now := time.Now()

	first, err := store.chatFor(context.Background(), 1, now, "system")
	if err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	second, err := store.chatFor(context.Background(), 1, now.Add(2*time.Minute), "system")
	if err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	if first == second {
		t.Error("chatFor() reused an expired session")
	}
}

func TestTouchTracksParticipants(t *testing.T) {
	store, _ := newTestStore(time.Hour)
	now := time.Now()
	if _, err := store.chatFor(context.Background(), 1, now, "system"); err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}

	store.touch(1, now, 10)
	store.touch(1, now, 11)

	got := store.participants(1, now)
	if len(got) != 2 {
		t.Fatalf("participants = %v, want 2 entries", got)
	}
}

func TestParticipantsEmptyForUnknownChannel(t *testing.T) {
	store, _ := newTestStore(time.Hour)
	if got := store.participants(999, time.Now()); got != nil {
		t.Errorf("participants = %v, want nil", got)
	}
}

func TestTouchRefreshesTTL(t *testing.T) {
	store, _ := newTestStore(time.Minute)
	now := time.Now()

	first, err := store.chatFor(context.Background(), 1, now, "system")
	if err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	store.touch(1, now.Add(30*time.Second), 10)

	second, err := store.chatFor(context.Background(), 1, now.Add(80*time.Second), "system")
	if err != nil {
		t.Fatalf("chatFor() error = %v", err)
	}
	if first != second {
		t.Error("chatFor() dropped a session that touch should have refreshed")
	}
}
