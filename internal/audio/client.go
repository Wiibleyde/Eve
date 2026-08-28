package audio

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"Eve/internal/config"
	"Eve/internal/logger"

	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

const NodeName = "main"

var (
	ErrDisabled  = errors.New("audio: lavalink is not configured")
	ErrNoNode    = errors.New("audio: no lavalink node available")
	ErrNoLyrics  = errors.New("audio: no lyrics available for this track")
	ErrNoSession = errors.New("audio: the lavalink node has no session yet")

	ErrNoExtractor = errors.New("audio: yt-dlp is not available")
)

func Enabled() bool {
	return config.Get().LavalinkAddress != ""
}

type Client struct {
	link *disgolink.Client
}

func New(userID snowflake.ID, opts ...disgolink.ConfigOpt) (*Client, error) {
	cfg := config.Get()
	if cfg.LavalinkAddress == "" {
		return nil, ErrDisabled
	}

	client := &Client{link: disgolink.New(userID, opts...)}

	go func() {
		node, err := client.link.AddNode(context.Background(), disgolink.NodeConfig{
			Name:     NodeName,
			Address:  cfg.LavalinkAddress,
			Password: cfg.LavalinkPassword,
			Secure:   cfg.LavalinkSecure,
		})
		if err != nil {
			logger.Error("Lavalink node could not be opened", "address", cfg.LavalinkAddress, "error", err)
			return
		}
		logger.Info("Lavalink node connected", "name", node.Config.Name, "address", node.Config.Address)
	}()

	return client, nil
}

func (c *Client) Connected() bool {
	return c.link.BestNode() != nil
}

func (c *Client) Node() (*disgolink.Node, error) {
	node := c.link.BestNode()
	if node == nil {
		return nil, ErrNoNode
	}
	return node, nil
}

func (c *Client) Player(guildID snowflake.ID) *disgolink.Player {
	return c.link.Player(guildID)
}

func (c *Client) ExistingPlayer(guildID snowflake.ID) *disgolink.Player {
	return c.link.ExistingPlayer(guildID)
}

func (c *Client) RemovePlayer(guildID snowflake.ID) {
	c.link.RemovePlayer(guildID)
}

func (c *Client) Load(ctx context.Context, query string, handler disgolink.TrackLoadingResultHandler) {
	node, err := c.Node()
	if err != nil {
		handler.OnError(err)
		return
	}
	node.Rest.LoadTracksHandler(ctx, query, handler)
}

func (c *Client) OnVoiceStateUpdate(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, sessionID string) {
	c.link.OnVoiceStateUpdate(ctx, guildID, channelID, sessionID)
}

func (c *Client) OnVoiceServerUpdate(ctx context.Context, guildID snowflake.ID, token string, endpoint string) {
	c.link.OnVoiceServerUpdate(ctx, guildID, token, endpoint)
}

func (c *Client) Close() {
	c.link.Close()
}

func (c *Client) LoadDirect(ctx context.Context, streamURL string) (lavalink.Track, error) {
	node, err := c.Node()
	if err != nil {
		return lavalink.Track{}, err
	}

	result, err := node.Rest.LoadTracks(ctx, streamURL)
	if err != nil {
		return lavalink.Track{}, fmt.Errorf("loading stream: %w", err)
	}

	switch data := result.Data.(type) {
	case lavalink.Track:
		return data, nil
	case lavalink.Search:
		if len(data) > 0 {
			return data[0], nil
		}
		return lavalink.Track{}, ErrNoMedia
	case lavalink.Exception:
		return lavalink.Track{}, data
	default:
		return lavalink.Track{}, ErrNoMedia
	}
}

func isURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}
