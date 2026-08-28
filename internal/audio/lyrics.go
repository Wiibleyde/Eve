package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/disgoorg/disgolink/v4/lavalink"
)

type LyricsLine struct {
	Line  string
	Start lavalink.Duration
}

type Lyrics struct {
	Source string
	Title  string
	Artist string
	Timed  bool
	Text   string
	Lines  []LyricsLine
}

type rawLyricsRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type rawLyricsLine struct {
	Line  string         `json:"line"`
	Range rawLyricsRange `json:"range"`
}

type rawLyricsTrack struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Album  string `json:"album"`
}

type rawLyrics struct {
	Type   string          `json:"type"`
	Source string          `json:"source"`
	Text   string          `json:"text"`
	Track  rawLyricsTrack  `json:"track"`
	Lines  []rawLyricsLine `json:"lines"`
}

type LyricsSearchResult struct {
	VideoID string `json:"videoId"`
	Title   string `json:"title"`
}

func (client *Client) LyricsForVideo(ctx context.Context, videoID string) (*Lyrics, error) {
	var payload rawLyrics
	if err := client.request(ctx, http.MethodGet, "/v4/lyrics/"+url.PathEscape(videoID), &payload); err != nil {
		return nil, err
	}

	lyrics := &Lyrics{
		Source: payload.Source,
		Title:  payload.Track.Title,
		Artist: payload.Track.Author,
		Timed:  payload.Type == "timed",
		Text:   payload.Text,
	}
	for _, line := range payload.Lines {
		if line.Line == "" {
			continue
		}
		lyrics.Lines = append(lyrics.Lines, LyricsLine{
			Line:  line.Line,
			Start: lavalink.Duration(line.Range.Start),
		})
	}

	if len(lyrics.Lines) == 0 && lyrics.Text == "" {
		return nil, ErrNoLyrics
	}
	return lyrics, nil
}

func (client *Client) SearchLyrics(ctx context.Context, query string) ([]LyricsSearchResult, error) {
	var results []LyricsSearchResult
	if err := client.request(ctx, http.MethodGet, "/v4/lyrics/search/"+url.PathEscape(query), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNoLyrics
	}
	return results, nil
}

func VideoID(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	switch {
	case host == "youtu.be":
		return strings.Trim(parsed.Path, "/")
	case strings.HasSuffix(host, "youtube.com"):
		if id := parsed.Query().Get("v"); id != "" {
			return id
		}
		if after, found := strings.CutPrefix(parsed.Path, "/shorts/"); found {
			return strings.Trim(after, "/")
		}
	}
	return ""
}

func (client *Client) request(ctx context.Context, method string, path string, out any) error {
	node, err := client.Node()
	if err != nil {
		return err
	}

	rq, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return fmt.Errorf("building lavalink request: %w", err)
	}

	rs, err := node.Rest.Do(rq)
	if err != nil {
		return fmt.Errorf("calling lavalink: %w", err)
	}
	defer func() {
		_ = rs.Body.Close()
	}()

	switch {
	case rs.StatusCode == http.StatusNotFound:
		return ErrNoLyrics
	case rs.StatusCode == http.StatusNoContent:
		return ErrNoLyrics
	case rs.StatusCode >= http.StatusBadRequest:
		return fmt.Errorf("lavalink returned %s for %s", rs.Status, path)
	}

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		return fmt.Errorf("reading lavalink response: %w", err)
	}
	if len(body) == 0 {
		return ErrNoLyrics
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("unmarshalling lavalink response: %w", err)
	}
	return nil
}
