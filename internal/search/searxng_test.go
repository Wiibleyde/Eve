package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchParsesAndCapsResults(t *testing.T) {
	var gotQuery, gotFormat string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		gotFormat = r.URL.Query().Get("format")
		_, _ = w.Write([]byte(`{"results":[
			{"title":"Un  titre","url":"https://www.example.com/a","content":"Premier  extrait"},
			{"title":"Deux","url":"https://fr.wikipedia.org/b","content":"Deuxième extrait"},
			{"title":"Trois","url":"https://example.org/c","content":"Troisième extrait"},
			{"title":"Quatre","url":"https://example.net/d","content":"Quatrième extrait"}
		]}`))
	}))
	defer srv.Close()

	results, err := New(srv.URL).Search(context.Background(), "météo demain")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != MaxResults {
		t.Fatalf("len = %d, want %d", len(results), MaxResults)
	}
	if gotQuery != "météo demain" || gotFormat != "json" {
		t.Errorf("query = %q, format = %q", gotQuery, gotFormat)
	}
	if results[0].Title != "Un titre" || results[0].Snippet != "Premier extrait" {
		t.Errorf("result[0] = %+v, want collapsed whitespace", results[0])
	}
	if results[0].Domain != "example.com" || results[1].Domain != "fr.wikipedia.org" {
		t.Errorf("domains = %q, %q", results[0].Domain, results[1].Domain)
	}
}

func TestSearchSkipsResultsWithoutContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"title":"Vide","url":"https://a.fr","content":"   "}]}`))
	}))
	defer srv.Close()

	results, err := New(srv.URL).Search(context.Background(), "test query")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results = %+v, want none", results)
	}
}

func TestSearchTruncatesLongSnippets(t *testing.T) {
	long := strings.Repeat("é", maxSnippetLength+100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"title":"T","url":"https://a.fr","content":"` + long + `"}]}`))
	}))
	defer srv.Close()

	results, err := New(srv.URL).Search(context.Background(), "test query")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if length := len([]rune(results[0].Snippet)); length != maxSnippetLength {
		t.Errorf("snippet length = %d, want %d", length, maxSnippetLength)
	}
}

func TestSearchServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("blocked"))
	}))
	defer srv.Close()

	if _, err := New(srv.URL).Search(context.Background(), "test query"); err == nil {
		t.Fatal("Search() error = nil, want error")
	}
}

func TestNilClientIsDisabled(t *testing.T) {
	var client *Client
	if _, err := client.Search(context.Background(), "test"); err != ErrDisabled {
		t.Errorf("error = %v, want ErrDisabled", err)
	}
}
