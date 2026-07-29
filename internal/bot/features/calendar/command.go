package calendar

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const commandTimeout = 30 * time.Second

const (
	msgGuildOnly       = "Cette commande doit être utilisée dans un serveur."
	msgNoPermission    = "Vous devez avoir la permission « Gérer les salons » pour utiliser cette commande."
	msgInvalidURL      = "URL invalide. Fournissez un lien `http`, `https` ou `webcal` vers un fichier `.ics`."
	msgFetchFailed     = "Impossible de récupérer ou de lire ce calendrier. Vérifiez que l'URL est publique et pointe bien vers un fichier `.ics`."
	msgSaveFailed      = "Erreur lors de l'enregistrement de la configuration."
	msgPostFailed      = "Impossible de publier le calendrier dans ce salon. Vérifiez mes permissions."
	msgNotConfigured   = "Aucun calendrier n'est configuré sur ce serveur. Utilisez `/calendar set`."
	msgRefreshFailed   = "Impossible d'actualiser le message du calendrier."
	msgRemoved         = "Calendrier supprimé."
	msgRefreshed       = "Calendrier actualisé."
	msgConfiguredFmt   = "Calendrier lié et publié dans <#%s>. Il sera actualisé automatiquement toutes les 5 minutes."
	msgEventsFoundFmt  = "\n%d évènement(s) lu(s) dans le fichier ICS."
	msgRecurringNotice = "\n⚠️ Les évènements récurrents (RRULE) ne sont pas développés : seule leur première occurrence est prise en compte."
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "calendar",
		Description: "Calendrier ICS du serveur",
		Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "set",
				Description: "[MODÉRATION] Lier un calendrier ICS et publier son embed dans ce salon",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "url",
						Description: "Lien vers le fichier .ics (http, https ou webcal)",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "[MODÉRATION] Délier le calendrier et supprimer son message",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "refresh",
				Description: "[MODÉRATION] Forcer l'actualisation du message du calendrier",
			},
		},
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return
	}
	switch *data.SubCommandName {
	case "set":
		handleSet(e)
	case "remove":
		handleRemove(e)
	case "refresh":
		handleRefresh(e)
	}
}

func guildContext(e *events.ApplicationCommandInteractionCreate) (snowflake.ID, bool) {
	if e.GuildID() == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgGuildOnly))
		return 0, false
	}
	if !helpers.RequirePermission(e, discord.PermissionManageChannels, msgNoPermission) {
		return 0, false
	}
	return *e.GuildID(), true
}

func handleSet(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := guildContext(e)
	if !ok {
		return
	}

	normalized, err := normalizeURL(e.SlashCommandInteractionData().String("url"))
	if err != nil {
		logger.Debug("Calendar: rejected url", "guild", guildID.String(), "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgInvalidURL))
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Calendar: deferring set response failed", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	parsed, err := fetchEvents(ctx, normalized)
	if err != nil {
		logger.Warn("Calendar: ICS validation failed", "guild", guildID.String(), "url", normalized, "error", err)
		editDeferred(e, ui.Error(msgFetchFailed))
		return
	}

	guildIDStr := guildID.String()

	if previous, err := loadGuildCalendar(ctx, guildIDStr); err != nil {
		logger.Warn("Calendar: could not read previous config", "guild", guildIDStr, "error", err)
	} else {
		deleteCalendarMessage(e.Client(), previous)
	}

	if err := saveConfig(ctx, guildIDStr, tables.CalendarURL, normalized); err != nil {
		logger.Error("Calendar: saving url failed", "guild", guildIDStr, "error", err)
		editDeferred(e, ui.Error(msgSaveFailed))
		return
	}

	channelID := e.Channel().ID()
	if err := postCalendarMessage(ctx, e.Client(), guildIDStr, channelID, buildCard(parsed, time.Now())); err != nil {
		logger.Error("Calendar: publishing calendar message failed", "guild", guildIDStr, "error", err)
		if rollbackErr := deleteConfig(ctx, guildIDStr, tables.CalendarURL, tables.CalendarChannel, tables.CalendarMessage); rollbackErr != nil {
			logger.Error("Calendar: rolling back config after a failed publish", "guild", guildIDStr, "error", rollbackErr)
		}
		editDeferred(e, ui.Error(msgPostFailed))
		return
	}

	message := fmt.Sprintf(msgConfiguredFmt, channelID.String()) +
		fmt.Sprintf(msgEventsFoundFmt, len(parsed))
	if hasRecurring(parsed) {
		message += msgRecurringNotice
	}
	editDeferred(e, ui.Success(message))
}

func handleRemove(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := guildContext(e)
	if !ok {
		return
	}
	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Calendar: deferring remove response failed", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	guildIDStr := guildID.String()

	cfg, err := loadGuildCalendar(ctx, guildIDStr)
	if err != nil {
		logger.Error("Calendar: loading config failed", "guild", guildIDStr, "error", err)
		editDeferred(e, ui.Error(msgSaveFailed))
		return
	}
	if !cfg.configured() && !cfg.hasMessage() {
		editDeferred(e, ui.Error(msgNotConfigured))
		return
	}

	deleteCalendarMessage(e.Client(), cfg)

	if err := deleteConfig(ctx, guildIDStr, tables.CalendarURL, tables.CalendarChannel, tables.CalendarMessage); err != nil {
		logger.Error("Calendar: deleting config failed", "guild", guildIDStr, "error", err)
		editDeferred(e, ui.Error(msgSaveFailed))
		return
	}
	editDeferred(e, ui.Success(msgRemoved))
}

func handleRefresh(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := guildContext(e)
	if !ok {
		return
	}
	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Calendar: deferring refresh response failed", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	guildIDStr := guildID.String()

	cfg, err := loadGuildCalendar(ctx, guildIDStr)
	if err != nil {
		logger.Error("Calendar: loading config failed", "guild", guildIDStr, "error", err)
		editDeferred(e, ui.Error(msgSaveFailed))
		return
	}
	if !cfg.configured() {
		editDeferred(e, ui.Error(msgNotConfigured))
		return
	}

	parsed, err := fetchEvents(ctx, cfg.URL)
	if err != nil {
		logger.Warn("Calendar: refresh fetch failed", "guild", guildIDStr, "error", err)
		editDeferred(e, ui.Error(msgFetchFailed))
		return
	}

	card := buildCard(parsed, time.Now())
	err = updateCalendarMessage(ctx, e.Client(), cfg, card)
	switch {
	case err == nil:
	case errors.Is(err, errCalendarMessageGone), errors.Is(err, errNoCalendarMessage):
		channelID := e.Channel().ID()
		if cfg.ChannelID != "" {
			if parsedID, parseErr := snowflake.Parse(cfg.ChannelID); parseErr == nil {
				channelID = parsedID
			}
		}
		if postErr := postCalendarMessage(ctx, e.Client(), guildIDStr, channelID, card); postErr != nil {
			logger.Error("Calendar: republishing calendar message failed", "guild", guildIDStr, "error", postErr)
			editDeferred(e, ui.Error(msgPostFailed))
			return
		}
	default:
		logger.Error("Calendar: refresh failed", "guild", guildIDStr, "error", err)
		editDeferred(e, ui.Error(msgRefreshFailed))
		return
	}

	message := msgRefreshed + fmt.Sprintf(msgEventsFoundFmt, len(parsed))
	if hasRecurring(parsed) {
		message += msgRecurringNotice
	}
	editDeferred(e, ui.Success(message))
}

func editDeferred(e *events.ApplicationCommandInteractionCreate, card *ui.Card) {
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), card)
}

func hasRecurring(parsed []Event) bool {
	for _, event := range parsed {
		if event.Recurring {
			return true
		}
	}
	return false
}
