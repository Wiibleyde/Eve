package twitch

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// StreamTypeLive is the only Stream.Type value that means "actually live".
// Twitch also returns "" for an error state, which must not be treated as live.
const StreamTypeLive = "live"

// Stream is a subset of the GET /helix/streams payload.
type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	IsMature     bool      `json:"is_mature"`
}

func (s Stream) IsLive() bool { return s.Type == StreamTypeLive }

// Name returns the display name, falling back to the login.
func (s Stream) Name() string {
	if s.UserName != "" {
		return s.UserName
	}
	return s.UserLogin
}

// Thumbnail fills the {width}/{height} placeholders Twitch leaves in the
// preview URL. It returns "" when Twitch sent no thumbnail.
func (s Stream) Thumbnail(width, height int) string {
	if s.ThumbnailURL == "" {
		return ""
	}
	u := strings.ReplaceAll(s.ThumbnailURL, "{width}", strconv.Itoa(width))
	return strings.ReplaceAll(u, "{height}", strconv.Itoa(height))
}

// GetStreamsByUserIDs returns the live streams among the given user IDs.
// Offline users are simply absent from the result.
//
// Requests are batched at 100 IDs. "first" is pinned to the batch size because
// Helix otherwise caps the page at 20 entries and would silently hide live
// streams once more than 20 tracked channels are on air at the same time.
func (c *Client) GetStreamsByUserIDs(ctx context.Context, userIDs []string) ([]Stream, error) {
	userIDs = dedupe(userIDs)
	if len(userIDs) == 0 {
		return nil, nil
	}

	out := make([]Stream, 0, len(userIDs))
	for _, batch := range chunk(userIDs, MaxIDsPerRequest) {
		query := url.Values{}
		for _, id := range batch {
			query.Add("user_id", id)
		}
		query.Set("first", strconv.Itoa(len(batch)))

		var resp response[Stream]
		if err := c.get(ctx, "/streams", query, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Data...)
	}
	return out, nil
}
