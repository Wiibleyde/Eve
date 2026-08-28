package audio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"Eve/internal/config"
	"Eve/internal/logger"

	"github.com/disgoorg/disgolink/v4/lavalink"
)

const (
	MaxPlaylistItems = 100
	searchPrefix     = "ytsearch1:"
)

var ErrNoMedia = errors.New("audio: yt-dlp returned no media")

type Media struct {
	Title      string
	Author     string
	URI        string
	ArtworkURL string
	Length     lavalink.Duration
	IsStream   bool
}

type Resolution struct {
	Media        []Media
	PlaylistName string
	Truncated    int
}

var (
	ytDlpOnce sync.Once
	ytDlpPath string
)

func ytDlp() string {
	ytDlpOnce.Do(func() {
		configured := config.Get().YtDlpPath
		if configured == "" {
			configured = "yt-dlp"
		}
		resolved, err := exec.LookPath(configured)
		if err != nil {
			logger.Warn("yt-dlp not found, music extraction is unavailable", "path", configured, "error", err)
			return
		}
		ytDlpPath = resolved
	})
	return ytDlpPath
}

func YtDlpAvailable() bool {
	return ytDlp() != ""
}

func Resolve(ctx context.Context, query string) (*Resolution, error) {
	binary := ytDlp()
	if binary == "" {
		return nil, ErrNoExtractor
	}

	target := strings.TrimSpace(query)
	if !isURL(target) {
		target = searchPrefix + target
	}

	output, err := run(ctx, binary,
		"--dump-single-json",
		"--flat-playlist",
		"--no-warnings",
		"--no-progress",
		"--ignore-config",
		target,
	)
	if err != nil {
		return nil, err
	}

	var payload rawEntry
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("parsing yt-dlp output: %w", err)
	}

	if len(payload.Entries) == 0 {
		media, ok := payload.media()
		if !ok {
			return nil, ErrNoMedia
		}
		return &Resolution{Media: []Media{media}}, nil
	}

	entries := payload.Entries
	truncated := 0
	if len(entries) > MaxPlaylistItems {
		truncated = len(entries) - MaxPlaylistItems
		entries = entries[:MaxPlaylistItems]
	}

	items := make([]Media, 0, len(entries))
	for _, entry := range entries {
		if media, ok := entry.media(); ok {
			items = append(items, media)
		}
	}
	if len(items) == 0 {
		return nil, ErrNoMedia
	}

	name := ""
	if !payload.isSearch() {
		name = payload.Title
	}

	return &Resolution{Media: items, PlaylistName: name, Truncated: truncated}, nil
}

type Tags struct {
	Track  string
	Artist string
	Album  string
}

func MusicTags(ctx context.Context, uri string) (Tags, error) {
	binary := ytDlp()
	if binary == "" {
		return Tags{}, ErrNoExtractor
	}

	output, err := run(ctx, binary,
		"--dump-single-json",
		"--no-playlist",
		"--skip-download",
		"--no-warnings",
		"--no-progress",
		"--ignore-config",
		uri,
	)
	if err != nil {
		return Tags{}, err
	}

	var payload struct {
		Track  string `json:"track"`
		Artist string `json:"artist"`
		Album  string `json:"album"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return Tags{}, fmt.Errorf("parsing yt-dlp tags: %w", err)
	}

	return Tags{Track: payload.Track, Artist: payload.Artist, Album: payload.Album}, nil
}

func Stream(ctx context.Context, uri string) (string, error) {
	binary := ytDlp()
	if binary == "" {
		return "", ErrNoExtractor
	}

	output, err := run(ctx, binary,
		"--format", "bestaudio/best",
		"--no-playlist",
		"--no-warnings",
		"--no-progress",
		"--ignore-config",
		"--get-url",
		uri,
	)
	if err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		if candidate := strings.TrimSpace(line); candidate != "" {
			return candidate, nil
		}
	}
	return "", ErrNoMedia
}

func run(ctx context.Context, binary string, args ...string) ([]byte, error) {
	if runtime := config.Get().YtDlpJSRuntime; runtime != "" {
		args = append(args, "--js-runtimes", runtime)
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return output, nil
}

type rawThumbnail struct {
	URL string `json:"url"`
}

type rawEntry struct {
	Type       string         `json:"_type"`
	Extractor  string         `json:"extractor_key"`
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Uploader   string         `json:"uploader"`
	Channel    string         `json:"channel"`
	Duration   *float64       `json:"duration"`
	WebpageURL string         `json:"webpage_url"`
	URL        string         `json:"url"`
	Thumbnail  string         `json:"thumbnail"`
	Thumbnails []rawThumbnail `json:"thumbnails"`
	LiveStatus string         `json:"live_status"`
	IsLive     bool           `json:"is_live"`
	Entries    []rawEntry     `json:"entries"`
}

func (e rawEntry) isSearch() bool {
	return strings.Contains(strings.ToLower(e.Extractor), "search")
}

func (e rawEntry) media() (Media, bool) {
	uri := e.WebpageURL
	if uri == "" {
		uri = e.URL
	}
	if uri == "" && e.ID != "" {
		uri = "https://www.youtube.com/watch?v=" + e.ID
	}
	if uri == "" || e.Title == "" {
		return Media{}, false
	}

	author := e.Uploader
	if author == "" {
		author = e.Channel
	}

	var length lavalink.Duration
	if e.Duration != nil {
		length = lavalink.Duration(*e.Duration * float64(lavalink.Second))
	}

	return Media{
		Title:      e.Title,
		Author:     author,
		URI:        uri,
		ArtworkURL: e.artwork(),
		Length:     length,
		IsStream:   e.IsLive || e.LiveStatus == "is_live",
	}, true
}

func (e rawEntry) artwork() string {
	if e.Thumbnail != "" {
		return e.Thumbnail
	}
	for i := len(e.Thumbnails) - 1; i >= 0; i-- {
		if e.Thumbnails[i].URL != "" {
			return e.Thumbnails[i].URL
		}
	}
	return ""
}
