package music

import (
	"context"
	"errors"
	"fmt"

	"Eve/internal/audio"
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

func ready(responder helpers.EphemeralResponder) (*audio.Client, bool) {
	if !Enabled() {
		helpers.RespondEphemeralCard(responder, ui.Error(MsgDisabled))
		return nil, false
	}

	client := lavalinkClient()
	if client == nil || !client.Connected() {
		helpers.RespondEphemeralCard(responder, ui.Error(MsgNodeDown))
		return nil, false
	}
	return client, true
}

func guildOf(responder helpers.EphemeralResponder, guildID *snowflake.ID) (snowflake.ID, bool) {
	if guildID == nil {
		helpers.RespondEphemeralCard(responder, ui.Error(MsgGuildOnly))
		return 0, false
	}
	return *guildID, true
}

func activePlayer(responder helpers.EphemeralResponder, client *audio.Client, guildID snowflake.ID) (*disgolink.Player, bool) {
	player := client.ExistingPlayer(guildID)
	if player == nil || player.Track == nil {
		helpers.RespondEphemeralCard(responder, ui.Error(MsgNotPlaying))
		return nil, false
	}
	return player, true
}

func voiceChannelOf(client *bot.Client, guildID snowflake.ID, userID snowflake.ID) (snowflake.ID, bool) {
	state, ok := client.Caches.VoiceState(guildID, userID)
	if !ok || state.ChannelID == nil {
		return 0, false
	}
	return *state.ChannelID, true
}

func HandlePlay(e *events.ApplicationCommandInteractionCreate) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	if !audio.YtDlpAvailable() {
		helpers.RespondEphemeralCard(e, ui.Error(MsgNoExtractor))
		return
	}
	voiceChannelID, ok := voiceChannelOf(e.Client(), guildID, e.User().ID)
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(MsgNoVoice))
		return
	}

	query := e.SlashCommandInteractionData().String(OptionTrack)
	if err := e.DeferCreateMessage(false); err != nil {
		logger.Error("Music: deferring play", "error", err)
		return
	}

	state := stateFor(guildID)
	state.setChannels(e.Channel().ID(), voiceChannelID)
	state.cancelLeave()

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	resolution, err := audio.Resolve(ctx, query)
	if err != nil {
		logger.Error("Music: resolving query", "guild", guildID.String(), "query", query, "error", err)
		helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), resolveErrorCard(err))
		return
	}

	card := submit(e.Client(), client, guildID, voiceChannelID, resolution)
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), card)
}

func resolveErrorCard(err error) *ui.Card {
	switch {
	case errors.Is(err, audio.ErrNoMedia):
		return ui.Error(MsgNoResult)
	case errors.Is(err, audio.ErrNoExtractor):
		return ui.Error(MsgNoExtractor)
	default:
		return ui.Error(MsgLoadFailed)
	}
}

func submit(client *bot.Client, audioClient *audio.Client, guildID snowflake.ID, voiceChannelID snowflake.ID, resolution *audio.Resolution) *ui.Card {
	if len(resolution.Media) == 0 {
		return ui.Error(MsgNoResult)
	}

	state := stateFor(guildID)
	player := audioClient.ExistingPlayer(guildID)

	if player != nil && player.Track != nil {
		state.enqueue(resolution.Media...)
		return queuedCard(resolution, false)
	}

	if len(resolution.Media) > 1 {
		state.enqueue(resolution.Media[1:]...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if err := client.UpdateVoiceState(ctx, guildID, &voiceChannelID, false, true); err != nil {
		logger.Error("Music: joining voice channel", "guild", guildID.String(), "error", err)
		return ui.Error(MsgPlayFailed)
	}

	play(guildID, resolution.Media[0])
	if state.current() == nil {
		return ui.Error(MsgPlayFailed)
	}

	return queuedCard(resolution, true)
}

func queuedCard(resolution *audio.Resolution, started bool) *ui.Card {
	var card *ui.Card

	switch {
	case resolution.PlaylistName != "":
		verb := "Ajout de"
		if started {
			verb = "Lecture de"
		}
		card = ui.Success(fmt.Sprintf("%s la playlist **%s** (%d musiques).", verb, resolution.PlaylistName, len(resolution.Media)))
	case started:
		card = ui.Success("Lecture de " + trackLine(resolution.Media[0]))
	default:
		card = ui.Success("Ajoutée à la file d'attente : " + trackLine(resolution.Media[0]))
	}

	if resolution.Truncated > 0 {
		card.Subtextf("%d musiques ignorées (limite de %d).", resolution.Truncated, audio.MaxPlaylistItems)
	}
	return card
}

func HandleQueue(e *events.ApplicationCommandInteractionCreate) {
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
	tracks := state.list()
	if len(tracks) == 0 {
		helpers.RespondEphemeralCard(e, ui.Error(MsgQueueEmpty))
		return
	}

	helpers.RespondEphemeralCard(e, queueCard(tracks, state.current(), state.repeatMode()))
}

func HandleNowPlaying(e *events.ApplicationCommandInteractionCreate) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	player, ok := activePlayer(e, client, guildID)
	if !ok {
		return
	}

	state := stateFor(guildID)
	current := state.current()
	if current == nil {
		helpers.RespondEphemeralCard(e, ui.Error(MsgNotPlaying))
		return
	}

	card := nowPlayingCard(*current, player.Position(), state.repeatMode(), player.Paused, state.size(), state.autoplayEnabled())
	helpers.RespondEphemeralCard(e, card)
}

func HandleSkip(e *events.ApplicationCommandInteractionCreate) {
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

	if err := e.DeferCreateMessage(false); err != nil {
		logger.Error("Music: deferring skip", "error", err)
		return
	}
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), skipCard(guildID))
}

type skipResult int

const (
	skipOK skipResult = iota
	skipQueueEnded
	skipFailed
)

func skipCard(guildID snowflake.ID) *ui.Card {
	switch skip(guildID) {
	case skipFailed:
		return ui.Error(MsgPlayFailed)
	case skipQueueEnded:
		return ui.Success(MsgQueueEnded)
	default:
		return ui.Success("Musique suivante.")
	}
}

func skip(guildID snowflake.ID) skipResult {
	state := stateFor(guildID)

	current := state.current()
	state.setCurrent(nil)
	if current != nil {
		state.pushHistory(*current)
		if state.repeatMode() == RepeatQueue {
			state.enqueue(*current)
		}
	}

	next, ok := state.pop()
	if !ok {
		if !stop(guildID) {
			return skipFailed
		}
		return skipQueueEnded
	}

	play(guildID, next)
	if state.current() == nil {
		return skipQueueEnded
	}
	return skipOK
}

func HandleBack(e *events.ApplicationCommandInteractionCreate) {
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

	if err := e.DeferCreateMessage(false); err != nil {
		logger.Error("Music: deferring back", "error", err)
		return
	}
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), backCard(guildID))
}

type backResult int

const (
	backOK backResult = iota
	backNoHistory
	backFailed
)

func backCard(guildID snowflake.ID) *ui.Card {
	switch back(guildID) {
	case backNoHistory:
		return ui.Error(MsgNoPrevious)
	case backFailed:
		return ui.Error(MsgPlayFailed)
	default:
		return ui.Success("Musique précédente.")
	}
}

func back(guildID snowflake.ID) backResult {
	state := stateFor(guildID)

	previous, ok := state.popHistory()
	if !ok {
		return backNoHistory
	}

	if current := state.current(); current != nil {
		state.enqueueFront(*current)
	}
	state.setCurrent(nil)

	play(guildID, previous)
	if state.current() == nil {
		return backFailed
	}
	return backOK
}

func HandlePause(e *events.ApplicationCommandInteractionCreate) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	player, ok := activePlayer(e, client, guildID)
	if !ok {
		return
	}
	if player.Paused {
		helpers.RespondEphemeralCard(e, ui.Error(MsgAlreadyPaused))
		return
	}

	if !setPaused(guildID, player, true) {
		helpers.RespondEphemeralCard(e, ui.Error(MsgPlayFailed))
		return
	}
	helpers.RespondCard(e, ui.Success("Musique mise en pause."))
}

func HandleResume(e *events.ApplicationCommandInteractionCreate) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	player, ok := activePlayer(e, client, guildID)
	if !ok {
		return
	}
	if !player.Paused {
		helpers.RespondEphemeralCard(e, ui.Error(MsgAlreadyPlaying))
		return
	}

	if !setPaused(guildID, player, false) {
		helpers.RespondEphemeralCard(e, ui.Error(MsgPlayFailed))
		return
	}
	helpers.RespondCard(e, ui.Success("Musique reprise."))
}

func setPaused(guildID snowflake.ID, player *disgolink.Player, paused bool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if err := player.Update(ctx, disgolink.WithPaused(paused)); err != nil {
		logger.Error("Music: updating pause state", "guild", guildID.String(), "paused", paused, "error", err)
		return false
	}
	refreshNowPlaying(guildID)
	return true
}

func HandleStop(e *events.ApplicationCommandInteractionCreate) {
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

	if !stop(guildID) {
		helpers.RespondEphemeralCard(e, ui.Error(MsgPlayFailed))
		return
	}
	helpers.RespondCard(e, ui.Success("La musique a été arrêtée."))
}

func stop(guildID snowflake.ID) bool {
	client := discordClient()
	if client == nil {
		return false
	}

	stateFor(guildID).clear()
	disconnect(client, guildID)
	return true
}

func HandleLoop(e *events.ApplicationCommandInteractionCreate) {
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

	var mode RepeatMode
	switch e.SlashCommandInteractionData().String(OptionAction) {
	case loopValueQueue:
		mode = RepeatQueue
	case loopValueTrack:
		mode = RepeatTrack
	case loopValueDisabled:
		mode = RepeatOff
	default:
		helpers.RespondEphemeralCard(e, ui.Error(MsgUnknownLoop))
		return
	}

	stateFor(guildID).setRepeat(mode)
	refreshNowPlaying(guildID)
	helpers.RespondEphemeralCard(e, ui.Success("Boucle : **"+mode.Label()+"**"))
}

func HandleClear(e *events.ApplicationCommandInteractionCreate) {
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

	removed := stateFor(guildID).clear()
	if removed == 0 {
		helpers.RespondEphemeralCard(e, ui.Error(MsgQueueEmpty))
		return
	}
	refreshNowPlaying(guildID)
	helpers.RespondCard(e, ui.Success(fmt.Sprintf("File d'attente vidée (**%d** musiques).", removed)))
}

func HandleRemove(e *events.ApplicationCommandInteractionCreate) {
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

	position := e.SlashCommandInteractionData().Int(OptionPosition)
	media, removed := stateFor(guildID).removeAt(position - 1)
	if !removed {
		helpers.RespondEphemeralCard(e, ui.Error(MsgBadPosition))
		return
	}

	refreshNowPlaying(guildID)
	helpers.RespondCard(e, ui.Success("Retirée de la file d'attente : "+trackLine(media)))
}

func HandleFilter(e *events.ApplicationCommandInteractionCreate) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	player, ok := activePlayer(e, client, guildID)
	if !ok {
		return
	}

	name := e.SlashCommandInteractionData().String(OptionFilter)
	label, known := audio.FilterLabel(name)
	if !known {
		helpers.RespondEphemeralCard(e, ui.Error(MsgUnknownFilter))
		return
	}

	filters, enabled, _ := audio.ToggleFilter(player.Filters, name)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if err := player.Update(ctx, disgolink.WithFilters(filters)); err != nil {
		logger.Error("Music: updating filters", "guild", guildID.String(), "filter", name, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(MsgPlayFailed))
		return
	}

	status := "désactivé"
	if enabled {
		status = "activé"
	}
	helpers.RespondEphemeralCard(e, ui.Success(fmt.Sprintf("Filtre **%s** %s.", label, status)))
}
