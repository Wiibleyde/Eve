package twitch

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const StreamTypeLive = "live"

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

func (s Stream) Name() string {
	if s.UserName != "" {
		return s.UserName
	}
	return s.UserLogin
}

func (s Stream) Thumbnail(width, height int) string {
	if s.ThumbnailURL == "" {
		return ""
	}
	u := strings.ReplaceAll(s.ThumbnailURL, "{width}", strconv.Itoa(width))
	return strings.ReplaceAll(u, "{height}", strconv.Itoa(height))
}

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
