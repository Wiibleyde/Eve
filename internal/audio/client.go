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

func (client *Client) Connected() bool {
	return client.link.BestNode() != nil
}

func (client *Client) Node() (*disgolink.Node, error) {
	node := client.link.BestNode()
	if node == nil {
		return nil, ErrNoNode
	}
	return node, nil
}

func (client *Client) Player(guildID snowflake.ID) *disgolink.Player {
	return client.link.Player(guildID)
}

func (client *Client) ExistingPlayer(guildID snowflake.ID) *disgolink.Player {
	return client.link.ExistingPlayer(guildID)
}

func (client *Client) RemovePlayer(guildID snowflake.ID) {
	client.link.RemovePlayer(guildID)
}

func (client *Client) Load(ctx context.Context, query string, handler disgolink.TrackLoadingResultHandler) {
	node, err := client.Node()
	if err != nil {
		handler.OnError(err)
		return
	}
	node.Rest.LoadTracksHandler(ctx, query, handler)
}

func (client *Client) OnVoiceStateUpdate(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, sessionID string) {
	client.link.OnVoiceStateUpdate(ctx, guildID, channelID, sessionID)
}

func (client *Client) OnVoiceServerUpdate(ctx context.Context, guildID snowflake.ID, token string, endpoint string) {
	client.link.OnVoiceServerUpdate(ctx, guildID, token, endpoint)
}

func (client *Client) Close() {
	client.link.Close()
}

func (client *Client) LoadDirect(ctx context.Context, streamURL string) (lavalink.Track, error) {
	node, err := client.Node()
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
