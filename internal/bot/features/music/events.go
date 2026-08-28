package music

import (
	"context"

	"Eve/internal/audio"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func onTrackStart(e *disgolink.PlayerTrackStartEvent) {
	guildID := e.Player.GuildID
	state := stateFor(guildID)
	state.cancelLeave()

	current := state.current()
	if current == nil {
		return
	}
	publishNowPlaying(guildID, *current, false)
}

func onTrackEnd(e *disgolink.PlayerTrackEndEvent) {
	guildID := e.Player.GuildID
	if client := discordClient(); client != nil {
		closeLyrics(client, guildID)
	}
	if !e.Reason.MayStartNext() {
		return
	}
	advance(guildID)
}

func onTrackException(e *disgolink.PlayerTrackExceptionEvent) {
	guildID := e.Player.GuildID
	title := ""
	if current := stateFor(guildID).current(); current != nil {
		title = current.Title
	}
	logger.Error("Music: track exception", "guild", guildID.String(), "track", title, "error", e.Exception.Error())
	notify(guildID, ui.Error(MsgPlayFailed))
}

func onTrackStuck(e *disgolink.PlayerTrackStuckEvent) {
	guildID := e.Player.GuildID
	logger.Warn("Music: track stuck", "guild", guildID.String(), "threshold", e.Threshold.String())
	advance(guildID)
}

func onWebSocketClosed(e *disgolink.PlayerWebSocketClosedEvent) {
	logger.Warn("Music: voice websocket closed",
		"guild", e.Player.GuildID.String(),
		"code", e.Code,
		"reason", e.Reason,
		"remote", e.ByRemote,
	)
	if !e.ByRemote {
		return
	}
	if client := discordClient(); client != nil {
		cleanup(client, e.Player.GuildID)
	}
}

func advance(guildID snowflake.ID) {
	state := stateFor(guildID)
	ended := state.current()
	state.setCurrent(nil)

	if ended != nil {
		if state.repeatMode() == RepeatTrack {
			play(guildID, *ended)
			return
		}
		if state.repeatMode() == RepeatQueue {
			state.enqueue(*ended)
		}
		state.pushHistory(*ended)
	}

	playNext(guildID)
}

func playNext(guildID snowflake.ID) {
	next, ok := stateFor(guildID).pop()
	if !ok {
		scheduleIdleLeave(guildID)
		return
	}
	play(guildID, next)
}

func play(guildID snowflake.ID, media audio.Media) {
	if err := startPlayback(guildID, media); err != nil {
		logger.Error("Music: starting playback", "guild", guildID.String(), "track", media.Title, "uri", media.URI, "error", err)
		stateFor(guildID).setCurrent(nil)
		notify(guildID, ui.Error("Impossible de lire **"+media.Title+"**, passage à la suivante."))
		playNext(guildID)
	}
}

func startPlayback(guildID snowflake.ID, media audio.Media) error {
	client := lavalinkClient()
	if client == nil {
		return audio.ErrDisabled
	}

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	streamURL, err := audio.Stream(ctx, media.URI)
	if err != nil {
		return err
	}

	track, err := client.LoadDirect(ctx, streamURL)
	if err != nil {
		return err
	}

	state := stateFor(guildID)
	state.setCurrent(&media)

	if err := client.Player(guildID).Update(ctx,
		disgolink.WithTrack(track),
		disgolink.WithVolume(defaultVolume),
	); err != nil {
		state.setCurrent(nil)
		return err
	}
	return nil
}

func scheduleIdleLeave(guildID snowflake.ID) {
	client := discordClient()
	if client == nil {
		return
	}
	stateFor(guildID).scheduleLeave(leaveOnEndDelay, func() {
		logger.Debug("Music: leaving after idle timeout", "guild", guildID.String())
		disconnect(client, guildID)
	})
}

func publishNowPlaying(guildID snowflake.ID, media audio.Media, paused bool) {
	client := discordClient()
	if client == nil {
		return
	}

	state := stateFor(guildID)
	textChannelID, _ := state.channels()
	if textChannelID == 0 {
		return
	}

	position := lavalink.Duration(0)
	if audioClient := lavalinkClient(); audioClient != nil {
		if player := audioClient.ExistingPlayer(guildID); player != nil {
			position = player.Position()
		}
	}

	card := nowPlayingCard(media, position, state.repeatMode(), paused, state.size())

	if messageID := state.nowPlaying(); messageID != 0 {
		if _, err := client.Rest.UpdateMessage(textChannelID, messageID, card.MessageUpdate()); err == nil {
			return
		}
	}

	message, err := client.Rest.CreateMessage(textChannelID, card.MessageCreate())
	if err != nil {
		logger.Error("Music: sending now playing message", "guild", guildID.String(), "error", err)
		return
	}
	state.setNowPlaying(message.ID)
}

func refreshNowPlaying(guildID snowflake.ID) {
	current := stateFor(guildID).current()
	if current == nil {
		return
	}

	paused := false
	if client := lavalinkClient(); client != nil {
		if player := client.ExistingPlayer(guildID); player != nil {
			paused = player.Paused
		}
	}
	publishNowPlaying(guildID, *current, paused)
}

func deleteNowPlaying(client *bot.Client, guildID snowflake.ID) {
	state, ok := existingState(guildID)
	if !ok {
		return
	}

	textChannelID, _ := state.channels()
	messageID := state.nowPlaying()
	if textChannelID == 0 || messageID == 0 {
		return
	}

	if err := client.Rest.DeleteMessage(textChannelID, messageID); err != nil {
		logger.Debug("Music: deleting now playing message", "guild", guildID.String(), "error", err)
	}
	state.setNowPlaying(0)
}

func notify(guildID snowflake.ID, card *ui.Card) {
	client := discordClient()
	if client == nil {
		return
	}

	state, ok := existingState(guildID)
	if !ok {
		return
	}
	textChannelID, _ := state.channels()
	if textChannelID == 0 {
		return
	}

	if _, err := client.Rest.CreateMessage(textChannelID, card.MessageCreate()); err != nil {
		logger.Error("Music: sending notification", "guild", guildID.String(), "error", err)
	}
}
