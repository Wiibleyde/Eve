package blague

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"Eve/internal/bot/ui"
)

func withToken(t *testing.T, value string) {
	t.Helper()
	tokenOnce.Do(func() {})
	previous := apiToken
	apiToken = value
	t.Cleanup(func() { apiToken = previous })
}

func withBaseURL(t *testing.T, url string) {
	t.Helper()
	previous := baseURL
	baseURL = url
	t.Cleanup(func() { baseURL = previous })
}

func TestFetchRandomSuccess(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"type":"dev","joke":"Question ?","answer":"Réponse."}`))
	}))
	defer srv.Close()

	withToken(t, "secret")
	withBaseURL(t, srv.URL)

	joke, err := fetchRandom(context.Background(), "dev")
	if err != nil {
		t.Fatalf("fetchRandom: %v", err)
	}
	if gotPath != "/type/dev/random" {
		t.Errorf("path = %q, want /type/dev/random", gotPath)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("auth = %q, want Bearer secret", gotAuth)
	}
	if joke.ID != 42 || joke.Joke != "Question ?" || joke.Answer != "Réponse." {
		t.Errorf("unexpected joke: %+v", joke)
	}
}

func TestFetchRandomRejectsUnknownCategory(t *testing.T) {
	withToken(t, "secret")
	if _, err := fetchRandom(context.Background(), "dark"); err != ErrUnknownCategory {
		t.Fatalf("err = %v, want ErrUnknownCategory", err)
	}
}

func TestFetchRandomWithoutToken(t *testing.T) {
	withToken(t, "")
	if _, err := fetchRandom(context.Background(), "global"); err != ErrDisabled {
		t.Fatalf("err = %v, want ErrDisabled", err)
	}
}

func TestFetchRandomHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	withToken(t, "bad")
	withBaseURL(t, srv.URL)

	if _, err := fetchRandom(context.Background(), "global"); err == nil {
		t.Fatal("expected an error on 401")
	}
}

func TestCommandsHiddenWithoutToken(t *testing.T) {
	withToken(t, "")
	if got := Commands(); len(got) != 0 {
		t.Fatalf("Commands() = %d entries, want 0 without a token", len(got))
	}

	withToken(t, "secret")
	if got := Commands(); len(got) != 1 {
		t.Fatalf("Commands() = %d entries, want 1 with a token", len(got))
	}
}

func TestJokeCardSpoilersAnswer(t *testing.T) {
	texts := ui.Texts(jokeCard(&Joke{Joke: "Question ?", Answer: "Réponse."}, "Développeur").Components())
	joined := strings.Join(texts, "\n")

	for _, want := range []string{"## 😂 Blague — Développeur", "Question ?", "||Réponse.||", FooterDisclaimer} {
		if !strings.Contains(joined, want) {
			t.Errorf("card is missing %q, got:\n%s", want, joined)
		}
	}
}

func TestCustomIDLength(t *testing.T) {
	if len(CustomIDPublic) > 100 {
		t.Fatalf("custom ID too long: %d", len(CustomIDPublic))
	}
}
