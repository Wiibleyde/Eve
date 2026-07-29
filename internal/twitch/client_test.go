package twitch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"a", " b ", "a", "", "  ", "b", "c"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dedupe = %v, want %v", got, want)
	}
}

func TestChunk(t *testing.T) {
	values := make([]string, 250)
	for i := range values {
		values[i] = string(rune('a' + i%26))
	}

	batches := chunk(values, MaxIDsPerRequest)
	if len(batches) != 3 {
		t.Fatalf("chunk produced %d batches, want 3", len(batches))
	}
	if len(batches[0]) != 100 || len(batches[1]) != 100 || len(batches[2]) != 50 {
		t.Fatalf("unexpected batch sizes: %d/%d/%d", len(batches[0]), len(batches[1]), len(batches[2]))
	}
	if chunk(nil, 10) != nil {
		t.Fatal("chunk(nil) should be nil")
	}
}

func TestStreamThumbnail(t *testing.T) {
	s := Stream{ThumbnailURL: "https://cdn/{width}x{height}.jpg"}
	if got := s.Thumbnail(1280, 720); got != "https://cdn/1280x720.jpg" {
		t.Fatalf("Thumbnail = %q", got)
	}
	if got := (Stream{}).Thumbnail(1280, 720); got != "" {
		t.Fatalf("empty thumbnail should stay empty, got %q", got)
	}
}

func TestStreamIsLive(t *testing.T) {
	if !(Stream{Type: "live"}).IsLive() {
		t.Fatal("type live should be live")
	}
	if (Stream{Type: ""}).IsLive() {
		t.Fatal("empty type must not count as live")
	}
}

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := New("cid", "secret")
	c.httpClient = srv.Client()
	return c, srv
}

type roundTripTo struct {
	base string
	rt   http.RoundTripper
}

func (r roundTripTo) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = "http"
	clone.URL.Host = r.base
	return r.rt.RoundTrip(clone)
}

func TestGetStreamsRefreshesTokenAndBatches(t *testing.T) {
	var tokenCalls, streamCalls atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalls.Add(1)
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok", ExpiresIn: 3600, TokenType: "bearer"})
	})
	mux.HandleFunc("/helix/streams", func(w http.ResponseWriter, r *http.Request) {
		streamCalls.Add(1)
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Client-Id"); got != "cid" {
			t.Errorf("Client-Id = %q", got)
		}
		ids := r.URL.Query()["user_id"]
		if first := r.URL.Query().Get("first"); first == "" {
			t.Error("first must be pinned to the batch size, Helix caps pages at 20 otherwise")
		}
		out := response[Stream]{}
		for _, id := range ids {
			out.Data = append(out.Data, Stream{ID: "s" + id, UserID: id, Type: StreamTypeLive})
		}
		json.NewEncoder(w).Encode(out)
	})

	c, srv := newTestClient(t, mux)
	c.httpClient.Transport = roundTripTo{base: srv.Listener.Addr().String(), rt: http.DefaultTransport}

	ids := make([]string, 150)
	for i := range ids {
		ids[i] = string(rune('A'+i/26)) + string(rune('a'+i%26))
	}

	streams, err := c.GetStreamsByUserIDs(context.Background(), ids)
	if err != nil {
		t.Fatalf("GetStreamsByUserIDs: %v", err)
	}
	if len(streams) != 150 {
		t.Fatalf("got %d streams, want 150", len(streams))
	}
	if got := streamCalls.Load(); got != 2 {
		t.Fatalf("streams endpoint called %d times, want 2 (150 IDs / 100 per request)", got)
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (token must be cached)", got)
	}
}

func TestGetRetriesOnceOn401(t *testing.T) {
	var tokenCalls, userCalls atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		n := tokenCalls.Add(1)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "tok" + string(rune('0'+n)),
			ExpiresIn:   3600,
		})
	})
	mux.HandleFunc("/helix/users", func(w http.ResponseWriter, r *http.Request) {
		if userCalls.Add(1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(response[User]{Data: []User{{ID: "42", Login: "someone"}}})
	})

	c, srv := newTestClient(t, mux)
	c.httpClient.Transport = roundTripTo{base: srv.Listener.Addr().String(), rt: http.DefaultTransport}

	user, found, err := c.GetUserByLogin(context.Background(), "SomeOne")
	if err != nil {
		t.Fatalf("GetUserByLogin: %v", err)
	}
	if !found || user.ID != "42" {
		t.Fatalf("got %+v found=%v", user, found)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token fetched %d times, want 2 (initial + refresh after 401)", got)
	}
	if got := userCalls.Load(); got != 2 {
		t.Fatalf("users called %d times, want 2", got)
	}
}

func TestGetUserByLoginMissingIsNotAnError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok", ExpiresIn: 3600})
	})
	mux.HandleFunc("/helix/users", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response[User]{})
	})

	c, srv := newTestClient(t, mux)
	c.httpClient.Transport = roundTripTo{base: srv.Listener.Addr().String(), rt: http.DefaultTransport}

	_, found, err := c.GetUserByLogin(context.Background(), "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("unknown login should report found=false")
	}
}

func TestAccessTokenCachedUntilExpiry(t *testing.T) {
	c := New("cid", "secret")
	c.token = "cached"
	c.expiresAt = time.Now().Add(time.Hour)

	got, err := c.accessToken(context.Background())
	if err != nil || got != "cached" {
		t.Fatalf("accessToken = %q, %v", got, err)
	}

	c.invalidate("other")
	if c.token != "cached" {
		t.Fatal("invalidate must ignore a token it does not hold")
	}
	c.invalidate("cached")
	if c.token != "" {
		t.Fatal("invalidate must drop the matching token")
	}
}

func TestAPIErrorMessage(t *testing.T) {
	e := &APIError{StatusCode: 429, Endpoint: "/streams", Message: "slow down"}
	if e.Error() == "" {
		t.Fatal("APIError must render")
	}
}
