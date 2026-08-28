package music

import (
	"context"
	"sync"
	"time"

	"Eve/internal/audio"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

const (
	leaveOnEndDelay   = time.Minute
	leaveOnEmptyDelay = time.Minute
	requestTimeout    = 15 * time.Second
	resolveTimeout    = 60 * time.Second
	defaultVolume     = 100
)

var (
	clientMu sync.RWMutex
	link     *audio.Client
)

func Enabled() bool {
	return audio.Enabled()
}

func lavalinkClient() *audio.Client {
	clientMu.RLock()
	defer clientMu.RUnlock()
	return link
}

func Attach(client *bot.Client) {
	if !Enabled() {
		return
	}
	client.AddEventListeners(
		bot.NewListenerFunc(onVoiceStateUpdate),
		bot.NewListenerFunc(onVoiceServerUpdate),
	)
}

func Start(client *bot.Client, userID snowflake.ID) {
	if !Enabled() {
		return
	}

	created, err := audio.New(userID,
		disgolink.WithListenerFunc(onTrackStart),
		disgolink.WithListenerFunc(onTrackEnd),
		disgolink.WithListenerFunc(onTrackException),
		disgolink.WithListenerFunc(onTrackStuck),
		disgolink.WithListenerFunc(onWebSocketClosed),
	)
	if err != nil {
		logger.Error("Music: creating lavalink client", "error", err)
		return
	}

	clientMu.Lock()
	link = created
	clientMu.Unlock()

	setDiscordClient(client)
	logger.Info("Music feature enabled")
}

var (
	discordMu sync.RWMutex
	botClient *bot.Client
)

func setDiscordClient(client *bot.Client) {
	discordMu.Lock()
	defer discordMu.Unlock()
	botClient = client
}

func discordClient() *bot.Client {
	discordMu.RLock()
	defer discordMu.RUnlock()
	return botClient
}

func onVoiceStateUpdate(e *events.GuildVoiceStateUpdate) {
	client := lavalinkClient()
	if client == nil {
		return
	}

	if e.VoiceState.UserID == e.Client().ApplicationID {
		client.OnVoiceStateUpdate(context.TODO(), e.VoiceState.GuildID, e.VoiceState.ChannelID, e.VoiceState.SessionID)
		if e.VoiceState.ChannelID == nil {
			cleanup(e.Client(), e.VoiceState.GuildID)
		}
		return
	}

	checkEmptyChannel(e.Client(), e.VoiceState.GuildID)
}

func onVoiceServerUpdate(e *events.VoiceServerUpdate) {
	client := lavalinkClient()
	if client == nil || e.Endpoint == nil {
		return
	}
	client.OnVoiceServerUpdate(context.TODO(), e.GuildID, e.Token, *e.Endpoint)
}

func checkEmptyChannel(client *bot.Client, guildID snowflake.ID) {
	state, ok := existingState(guildID)
	if !ok {
		return
	}

	_, voiceChannelID := state.channels()
	if voiceChannelID == 0 {
		return
	}

	if listeners(client, guildID, voiceChannelID) > 0 {
		state.cancelLeave()
		return
	}

	state.scheduleLeave(leaveOnEmptyDelay, func() {
		if listeners(client, guildID, voiceChannelID) > 0 {
			return
		}
		logger.Debug("Music: leaving empty voice channel", "guild", guildID.String())
		disconnect(client, guildID)
	})
}

func listeners(client *bot.Client, guildID snowflake.ID, voiceChannelID snowflake.ID) int {
	count := 0
	for state := range client.Caches.VoiceStates(guildID) {
		if state.ChannelID == nil || *state.ChannelID != voiceChannelID {
			continue
		}
		if state.UserID == client.ApplicationID {
			continue
		}
		count++
	}
	return count
}

func disconnect(client *bot.Client, guildID snowflake.ID) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if err := client.UpdateVoiceState(ctx, guildID, nil, false, false); err != nil {
		logger.Error("Music: leaving voice channel", "guild", guildID.String(), "error", err)
	}
	cleanup(client, guildID)
}

func cleanup(client *bot.Client, guildID snowflake.ID) {
	audioClient := lavalinkClient()
	if audioClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	closeLyrics(client, guildID)

	if player := audioClient.ExistingPlayer(guildID); player != nil {
		if err := player.Destroy(ctx); err != nil {
			logger.Debug("Music: destroying player", "guild", guildID.String(), "error", err)
		}
	}

	deleteNowPlaying(client, guildID)
	dropState(guildID)
}
