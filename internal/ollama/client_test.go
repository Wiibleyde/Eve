package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestChatSuccess(t *testing.T) {
	var got chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %q, want /api/chat", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"  Bonjour.  "},"done":true}`))
	}))
	defer srv.Close()

	client := New(srv.URL, "llama3.2:1b", 3)
	answer, err := client.Chat(context.Background(), []Message{{Role: RoleUser, Content: "salut"}})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if answer != "Bonjour." {
		t.Errorf("answer = %q, want %q", answer, "Bonjour.")
	}
	if got.Model != "llama3.2:1b" {
		t.Errorf("model = %q, want llama3.2:1b", got.Model)
	}
	if got.Stream {
		t.Error("stream = true, want false")
	}
	if got.Options.NumThread != 3 {
		t.Errorf("num_thread = %d, want 3", got.Options.NumThread)
	}
	if got.Options.NumPredict != numPredict {
		t.Errorf("num_predict = %d, want %d", got.Options.NumPredict, numPredict)
	}
}

func TestChatServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"model not found"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, "missing", 1).Chat(context.Background(), []Message{{Role: RoleUser, Content: "salut"}})
	if err == nil {
		t.Fatal("Chat() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "model not found") {
		t.Errorf("error = %v, want it to mention the server payload", err)
	}
}

func TestChatEmptyAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"   "},"done":true}`))
	}))
	defer srv.Close()

	if _, err := New(srv.URL, "llama3.2:1b", 1).Chat(context.Background(), nil); err == nil {
		t.Fatal("Chat() error = nil, want error")
	}
}

func TestChatRejectsWhenQueueIsFull(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"ok"},"done":true}`))
	}))
	defer srv.Close()

	client := New(srv.URL, "llama3.2:1b", 1)
	done := make(chan struct{})
	for range MaxQueued {
		go func() {
			defer func() { done <- struct{}{} }()
			_, _ = client.Chat(context.Background(), nil)
		}()
	}
	waitForQueue(t, client, MaxQueued)

	if _, err := client.Chat(context.Background(), nil); err != ErrBusy {
		t.Errorf("error = %v, want ErrBusy", err)
	}

	close(release)
	for range MaxQueued {
		<-done
	}
}

func TestNilClientIsDisabled(t *testing.T) {
	var client *Client
	if _, err := client.Chat(context.Background(), nil); err != ErrDisabled {
		t.Errorf("error = %v, want ErrDisabled", err)
	}
}

func waitForQueue(t *testing.T, client *Client, want int64) {
	t.Helper()
	for range 1000 {
		if client.queued.Load() == want {
			return
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("queued = %d, want %d", client.queued.Load(), want)
}
