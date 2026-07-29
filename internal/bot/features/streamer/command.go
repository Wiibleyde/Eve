package streamer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/stream"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const commandTimeout = 2500 * time.Millisecond

const (
	msgNoPermission  = "Vous devez avoir la permission **Gérer les salons** pour utiliser cette commande."
	msgGuildOnly     = "Cette commande doit être utilisée dans un serveur."
	msgDBError       = "Erreur lors de l'accès à la base de données."
	msgTwitchError   = "Impossible de contacter l'API Twitch. Réessayez dans un instant."
	msgInvalidLogin  = "Nom de chaîne Twitch invalide. Donnez le nom de la chaîne (par ex. `pseudo`) ou son lien."
	msgNotConfigured = "La fonctionnalité Twitch n'est pas configurée sur ce bot."
)

const listLimit = 2500

var loginPattern = regexp.MustCompile(`^[a-z0-9_]{3,25}$`)

var command = discord.SlashCommandCreate{
	Name:        CommandName,
	Description: "Gestion des notifications de streams Twitch",
	Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "add",
			Description: "[MODÉRATION] Suivre une chaîne Twitch",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "streamer",
					Description: "Nom de la chaîne Twitch (ou son lien)",
					Required:    true,
				},
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "Salon où poster les notifications",
					Required:    true,
					ChannelTypes: []discord.ChannelType{
						discord.ChannelTypeGuildText,
						discord.ChannelTypeGuildNews,
					},
				},
				discord.ApplicationCommandOptionRole{
					Name:        "role",
					Description: "Rôle à mentionner lors du passage en live",
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "[MODÉRATION] Ne plus suivre une chaîne Twitch",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "streamer",
					Description: "Nom de la chaîne Twitch suivie",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "list",
			Description: "[MODÉRATION] Lister les chaînes Twitch suivies sur ce serveur",
		},
	},
}

func Commands() []discord.ApplicationCommandCreate {
	if !Enabled() {
		warnDisabled()
		return nil
	}
	return []discord.ApplicationCommandCreate{command}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		helpers.RespondEphemeralCard(e, ui.Error(router.MsgUnknownInteraction))
		return
	}
	if !Enabled() {
		helpers.RespondEphemeralCard(e, ui.Error(msgNotConfigured))
		return
	}
	switch *data.SubCommandName {
	case "add":
		handleAdd(e)
	case "remove":
		handleRemove(e)
	case "list":
		handleList(e)
	default:
		logger.Debug("Unknown streamer subcommand", "subcommand", *data.SubCommandName)
		helpers.RespondEphemeralCard(e, ui.Error(router.MsgUnknownInteraction))
	}
}

func handleAdd(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := requireGuildAndPermission(e)
	if !ok {
		return
	}

	data := e.SlashCommandInteractionData()
	login, ok := normalizeLogin(data.String("streamer"))
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(msgInvalidLogin))
		return
	}
	channel := data.Channel("channel")
	role, hasRole := data.OptRole("role")

	db := entClient()
	if db == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	user, found, err := helixClient().GetUserByLogin(ctx, login)
	if err != nil {
		logger.Error("Resolving Twitch login", "login", login, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgTwitchError))
		return
	}
	if !found {
		helpers.RespondEphemeralCard(e, ui.Error(
			fmt.Sprintf("La chaîne Twitch `%s` est introuvable.", login),
		))
		return
	}

	existing, err := findTracked(ctx, db, guildID, user.ID)
	if err != nil {
		logger.Error("Querying tracked stream", "guild", guildID, "twitch_user", user.ID, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	channelID := channel.ID.String()

	if existing == nil {
		create := db.Stream.Create().
			SetGuildID(guildID).
			SetChannelID(channelID).
			SetTwitchUserID(user.ID).
			SetTwitchLogin(user.Login)
		if hasRole {
			create.SetRoleID(role.ID.String())
		}
		err := create.Exec(ctx)
		if err == nil {
			helpers.RespondEphemeralCard(e, ui.Success(addedMessage(user.Name(), channelID, role.ID.String(), hasRole)))
			logger.Info("Twitch channel tracked", "guild", guildID, "login", user.Login, "channel", channelID)
			return
		}
		if !ent.IsConstraintError(err) {
			logger.Error("Creating tracked stream", "guild", guildID, "twitch_user", user.ID, "error", err)
			helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
			return
		}
		existing, err = findTracked(ctx, db, guildID, user.ID)
		if err != nil || existing == nil {
			logger.Error("Reloading tracked stream after conflict", "guild", guildID, "twitch_user", user.ID, "error", err)
			helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
			return
		}
	}

	update := db.Stream.UpdateOneID(existing.ID).
		SetChannelID(channelID).
		SetTwitchLogin(user.Login)
	if hasRole {
		update.SetRoleID(role.ID.String())
	} else {
		update.ClearRoleID()
	}
	moved := existing.ChannelID != channelID
	if moved {
		update.ClearMessageID()
	}
	if err := update.Exec(ctx); err != nil {
		logger.Error("Updating tracked stream", "guild", guildID, "twitch_user", user.ID, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	helpers.RespondEphemeralCard(e, ui.Success(updatedMessage(user.Name(), channelID, role.ID.String(), hasRole)))
	if moved {
		closeNotifications(e.Client(), existing)
	}
	logger.Info("Twitch channel tracking updated", "guild", guildID, "login", user.Login, "channel", channelID)
}

func findTracked(ctx context.Context, db *ent.Client, guildID, twitchUserID string) (*ent.Stream, error) {
	row, err := db.Stream.Query().
		Where(stream.GuildID(guildID), stream.TwitchUserID(twitchUserID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return row, err
}

func handleRemove(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := requireGuildAndPermission(e)
	if !ok {
		return
	}

	data := e.SlashCommandInteractionData()
	login, ok := normalizeLogin(data.String("streamer"))
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(msgInvalidLogin))
		return
	}

	db := entClient()
	if db == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	rows, err := db.Stream.Query().
		Where(stream.GuildID(guildID), stream.TwitchLogin(login)).
		All(ctx)
	if err != nil {
		logger.Error("Querying tracked stream by login", "guild", guildID, "login", login, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	if len(rows) == 0 {
		user, found, err := helixClient().GetUserByLogin(ctx, login)
		if err != nil {
			logger.Error("Resolving Twitch login", "login", login, "error", err)
			helpers.RespondEphemeralCard(e, ui.Error(msgTwitchError))
			return
		}
		if found {
			rows, err = db.Stream.Query().
				Where(stream.GuildID(guildID), stream.TwitchUserID(user.ID)).
				All(ctx)
			if err != nil {
				logger.Error("Querying tracked stream by id", "guild", guildID, "twitch_user", user.ID, "error", err)
				helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
				return
			}
		}
	}

	if len(rows) == 0 {
		helpers.RespondEphemeralCard(e, ui.Error(
			fmt.Sprintf("La chaîne `%s` n'est pas suivie sur ce serveur.", login),
		))
		return
	}

	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	if _, err := db.Stream.Delete().Where(stream.IDIn(ids...)).Exec(ctx); err != nil {
		logger.Error("Deleting tracked stream", "guild", guildID, "login", login, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	helpers.RespondEphemeralCard(e, ui.Success(
		fmt.Sprintf("La chaîne **%s** n'est plus suivie.", login),
	))
	closeNotifications(e.Client(), rows...)
	logger.Info("Twitch channel untracked", "guild", guildID, "login", login)
}

func handleList(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := requireGuildAndPermission(e)
	if !ok {
		return
	}

	db := entClient()
	if db == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	rows, err := db.Stream.Query().
		Where(stream.GuildID(guildID)).
		Order(stream.ByTwitchLogin()).
		All(ctx)
	if err != nil {
		logger.Error("Listing tracked streams", "guild", guildID, "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(msgDBError))
		return
	}
	if len(rows) == 0 {
		helpers.RespondEphemeral(e, "Aucune chaîne Twitch suivie sur ce serveur.")
		return
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		line := fmt.Sprintf("**%s** — <#%s>", row.TwitchLogin, row.ChannelID)
		if row.RoleID != "" {
			line += fmt.Sprintf(" — <@&%s>", row.RoleID)
		}
		if row.MessageID != "" {
			line += " — 🔴 en live"
		}
		lines = append(lines, line)
	}

	card := ui.New().
		Accent(twitchColor).
		Title("📺 Chaînes Twitch suivies").
		Text(joinCapped(lines))
	helpers.RespondEphemeralCard(e, card)
}

func joinCapped(lines []string) string {
	var b strings.Builder
	for i, line := range lines {
		omitted := fmt.Sprintf("\n… et %d de plus", len(lines)-i)
		if b.Len()+len("\n")+len(line) > listLimit-len(omitted) {
			b.WriteString(omitted)
			break
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	return b.String()
}

func requireGuildAndPermission(e *events.ApplicationCommandInteractionCreate) (string, bool) {
	if e.GuildID() == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgGuildOnly))
		return "", false
	}
	if !helpers.RequirePermission(e, discord.PermissionManageChannels, msgNoPermission) {
		return "", false
	}
	return e.GuildID().String(), true
}

func normalizeLogin(raw string) (string, bool) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.TrimPrefix(s, "twitch.tv/")
	s = strings.TrimPrefix(s, "m.twitch.tv/")
	s = strings.TrimPrefix(s, "@")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	if !loginPattern.MatchString(s) {
		return "", false
	}
	return s, true
}

func addedMessage(name, channelID, roleID string, hasRole bool) string {
	msg := fmt.Sprintf("La chaîne **%s** est désormais suivie dans <#%s>.", name, channelID)
	if hasRole {
		msg += fmt.Sprintf("\nRôle mentionné : <@&%s>.", roleID)
	}
	return msg
}

func updatedMessage(name, channelID, roleID string, hasRole bool) string {
	msg := fmt.Sprintf("Le suivi de la chaîne **%s** a été mis à jour : notifications dans <#%s>.", name, channelID)
	if hasRole {
		msg += fmt.Sprintf("\nRôle mentionné : <@&%s>.", roleID)
	} else {
		msg += "\nAucun rôle ne sera mentionné."
	}
	return msg
}
