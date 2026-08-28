package music

import (
	"context"
	"errors"
	"strings"
	"time"

	"Eve/internal/audio"
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const (
	maxThreadName        = 90
	lyricsTick           = 500 * time.Millisecond
	lyricsCandidateLimit = 4
	lyricsMatchThreshold = 0.75
)

func HandleSyncedLyrics(e *events.ApplicationCommandInteractionCreate) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	if _, ok := activePlayer(e, client, guildID); !ok {
		return
	}

	state := stateFor(guildID)
	current := state.current()
	if current == nil {
		helpers.RespondEphemeralCard(e, ui.Error(MsgNotPlaying))
		return
	}
	if state.lyricsThread() != 0 {
		helpers.RespondEphemeralCard(e, ui.Error(MsgLyricsRunning))
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Music: deferring lyrics", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	lyrics, err := fetchLyrics(ctx, client, *current)
	if err != nil {
		helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), lyricsErrorCard(guildID, err))
		return
	}
	if !lyrics.Timed || len(lyrics.Lines) == 0 {
		helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), ui.Error(MsgLyricsNotTimed))
		return
	}

	channelID, _ := state.channels()
	if channelID == 0 {
		channelID = e.Channel().ID()
		state.setChannels(channelID, 0)
	}

	thread, err := e.Client().Rest.CreateThread(channelID, discord.GuildPublicThreadCreate{
		Name:                threadName(current.Title),
		AutoArchiveDuration: discord.AutoArchiveDuration1h,
	})
	if err != nil {
		logger.Error("Music: creating lyrics thread", "guild", guildID.String(), "error", err)
		helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), ui.Error(MsgThreadFailed))
		return
	}

	streamCtx, streamCancel := context.WithCancel(context.Background())
	state.setLyrics(thread.ID(), streamCancel)
	go streamLyrics(streamCtx, e.Client(), guildID, thread.ID(), current.URI, lyrics.Lines)

	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(),
		ui.Success("Paroles synchronisées affichées dans <#"+thread.ID().String()+">"))
}

func streamLyrics(ctx context.Context, client *bot.Client, guildID snowflake.ID, threadID snowflake.ID, uri string, lines []audio.LyricsLine) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("Music: panic in lyrics streamer", "guild", guildID.String(), "panic", rec)
		}
	}()

	ticker := time.NewTicker(lyricsTick)
	defer ticker.Stop()

	state := stateFor(guildID)
	index := 0

	for index < len(lines) {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		current := state.current()
		if current == nil || current.URI != uri {
			return
		}

		audioClient := lavalinkClient()
		if audioClient == nil {
			return
		}
		player := audioClient.ExistingPlayer(guildID)
		if player == nil {
			return
		}
		position := player.Position()

		for index < len(lines) && lines[index].Start <= position {
			line := lines[index].Line
			index++

			if _, err := client.Rest.CreateMessage(threadID, discord.MessageCreate{
				Content:         line,
				AllowedMentions: ui.NoMentions(),
			}); err != nil {
				logger.Debug("Music: sending lyrics line", "guild", guildID.String(), "error", err)
				return
			}
		}
	}
}

func fetchLyrics(ctx context.Context, client *audio.Client, media audio.Media) (*audio.Lyrics, error) {
	if videoID := audio.VideoID(media.URI); videoID != "" {
		lyrics, err := client.LyricsForVideo(ctx, videoID)
		if err == nil {
			return lyrics, nil
		}
		if !errors.Is(err, audio.ErrNoLyrics) {
			return nil, err
		}
	}

	title, artist := trackIdentity(ctx, media)
	if title == "" {
		return nil, audio.ErrNoLyrics
	}

	results, err := client.SearchLyrics(ctx, strings.TrimSpace(title+" "+artist))
	if err != nil {
		return nil, err
	}

	tried := 0
	for _, result := range results {
		if tried >= lyricsCandidateLimit {
			break
		}
		if containment(title, result.Title) < lyricsMatchThreshold {
			continue
		}
		tried++

		lyrics, err := client.LyricsForVideo(ctx, result.VideoID)
		if err != nil {
			continue
		}
		if lyrics.Title != "" && containment(title, lyrics.Title) < lyricsMatchThreshold {
			logger.Debug("Music: rejecting mismatched lyrics",
				"expected", title, "found", lyrics.Title, "artist", lyrics.Artist)
			continue
		}
		return lyrics, nil
	}

	return nil, audio.ErrNoLyrics
}

func trackIdentity(ctx context.Context, media audio.Media) (string, string) {
	if tags, err := audio.MusicTags(ctx, media.URI); err == nil && tags.Track != "" {
		artist := tags.Artist
		if artist == "" {
			artist = cleanArtist(media.Author)
		}
		return tags.Track, artist
	}

	artist, title := splitArtistTitle(media.Title)
	if artist == "" {
		artist = cleanArtist(media.Author)
	}
	return title, artist
}

func lyricsErrorCard(guildID snowflake.ID, err error) *ui.Card {
	if errors.Is(err, audio.ErrNoLyrics) {
		return ui.Error(MsgLyricsMissing)
	}
	logger.Error("Music: fetching lyrics", "guild", guildID.String(), "error", err)
	return ui.Error(MsgLyricsFailed)
}

func closeLyrics(client *bot.Client, guildID snowflake.ID) {
	state, ok := existingState(guildID)
	if !ok {
		return
	}

	threadID, cancel := state.takeLyrics()
	if cancel != nil {
		cancel()
	}
	if threadID == 0 {
		return
	}

	if err := client.Rest.DeleteChannel(threadID); err != nil {
		logger.Debug("Music: deleting lyrics thread", "guild", guildID.String(), "error", err)
	}
}

func threadName(title string) string {
	name := "Paroles — " + title
	runes := []rune(name)
	if len(runes) <= maxThreadName {
		return name
	}
	return string(runes[:maxThreadName-1]) + "…"
}
