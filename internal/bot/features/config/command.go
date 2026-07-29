package config

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	maxDescriptionLength = 100
	maxChoiceNameLength  = 100
	maxStringValueLength = 500
)

const (
	optionValue = "valeur"
	optionKey   = "cle"
)

const (
	msgNoPermission = "Vous n'avez pas la permission d'utiliser cette commande."
	msgGuildOnly    = "Cette commande doit être utilisée dans un serveur."
	msgSaveError    = "Erreur lors de l'enregistrement."
	msgReadError    = "Erreur lors de la récupération de la configuration."
	msgDeleteError  = "Erreur lors de la réinitialisation."
	msgUnknownKey   = "Clé de configuration inconnue."
	msgMissingValue = "Valeur manquante ou invalide."
	msgNotSet       = "Non défini"
)

var Commands = []discord.ApplicationCommandCreate{buildCommand()}

func buildCommand() discord.SlashCommandCreate {
	options := make([]discord.ApplicationCommandOption, 0, len(Keys)+3)

	for _, k := range Keys {
		options = append(options, discord.ApplicationCommandOptionSubCommand{
			Name:        k.Command,
			Description: truncate("[ADMIN] Définir : "+k.Description, maxDescriptionLength),
			Options:     []discord.ApplicationCommandOption{valueOption(k)},
		})
	}

	options = append(options,
		discord.ApplicationCommandOptionSubCommand{
			Name:        "get",
			Description: "[ADMIN] Afficher la valeur d'un paramètre",
			Options:     []discord.ApplicationCommandOption{keyOption("Paramètre à afficher")},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "reset",
			Description: "[ADMIN] Réinitialiser un paramètre",
			Options:     []discord.ApplicationCommandOption{keyOption("Paramètre à réinitialiser")},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "list",
			Description: "[ADMIN] Lister la configuration du serveur",
		},
	)

	return discord.SlashCommandCreate{
		Name:        "config",
		Description: "Configuration du serveur",
		Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
		Options:     options,
	}
}

func valueOption(k Key) discord.ApplicationCommandOption {
	description := truncate(k.Description, maxDescriptionLength)
	switch k.Kind {
	case KindChannel:
		return discord.ApplicationCommandOptionChannel{
			Name:        optionValue,
			Description: description,
			Required:    true,
			ChannelTypes: []discord.ChannelType{
				discord.ChannelTypeGuildText,
			},
		}
	case KindRole:
		return discord.ApplicationCommandOptionRole{
			Name:        optionValue,
			Description: description,
			Required:    true,
		}
	case KindBool:
		return discord.ApplicationCommandOptionBool{
			Name:        optionValue,
			Description: description,
			Required:    true,
		}
	default:
		maxLength := maxStringValueLength
		return discord.ApplicationCommandOptionString{
			Name:        optionValue,
			Description: description,
			Required:    true,
			MaxLength:   &maxLength,
		}
	}
}

func keyOption(description string) discord.ApplicationCommandOption {
	choices := make([]discord.ApplicationCommandOptionChoiceString, 0, len(Keys))
	for _, k := range Keys {
		choices = append(choices, discord.ApplicationCommandOptionChoiceString{
			Name:  choiceLabel(k),
			Value: k.Name,
		})
	}
	return discord.ApplicationCommandOptionString{
		Name:        optionKey,
		Description: truncate(description, maxDescriptionLength),
		Required:    true,
		Choices:     choices,
	}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return
	}
	guildID, ok := requireGuild(e)
	if !ok {
		return
	}
	if !requireAdmin(e) {
		return
	}

	switch name := *data.SubCommandName; name {
	case "get":
		handleGet(e, guildID)
	case "reset":
		handleReset(e, guildID)
	case "list":
		handleList(e, guildID)
	default:
		key, found := keyByCommand(name)
		if !found {
			helpers.RespondEphemeralCard(e, ui.Error(msgUnknownKey))
			return
		}
		handleSet(e, guildID, key)
	}
}

func handleSet(e *events.ApplicationCommandInteractionCreate, guildID string, key Key) {
	raw, ok := readValue(e, key)
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(msgMissingValue))
		return
	}

	if err := setValue(context.Background(), guildID, key.Name, raw); err != nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgSaveError))
		return
	}

	helpers.RespondEphemeralCard(e, ui.Success(
		fmt.Sprintf("%s : %s", key.Description, formatValue(key, raw)),
	))
}

func readValue(e *events.ApplicationCommandInteractionCreate, key Key) (string, bool) {
	data := e.SlashCommandInteractionData()
	switch key.Kind {
	case KindChannel:
		channel, ok := data.OptChannel(optionValue)
		if !ok {
			return "", false
		}
		return channel.ID.String(), true
	case KindRole:
		role, ok := data.OptRole(optionValue)
		if !ok {
			return "", false
		}
		return role.ID.String(), true
	case KindBool:
		value, ok := data.OptBool(optionValue)
		if !ok {
			return "", false
		}
		return FormatBool(value), true
	default:
		value, ok := data.OptString(optionValue)
		if !ok {
			return "", false
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", false
		}
		return truncate(value, maxStringValueLength), true
	}
}

func handleGet(e *events.ApplicationCommandInteractionCreate, guildID string) {
	key, ok := keyByName(e.SlashCommandInteractionData().String(optionKey))
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(msgUnknownKey))
		return
	}

	raw, found, err := Value(context.Background(), guildID, key.Name)
	if err != nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgReadError))
		return
	}

	value := msgNotSet
	if found {
		value = formatValue(key, raw)
	}
	card := ui.New().
		Title("⚙️ Configuration").
		Fields(ui.Field{Name: key.Description, Value: value, Inline: true}).
		Subtext("Clé : `" + key.Name + "`")
	helpers.RespondEphemeralCard(e, card)
}

func handleReset(e *events.ApplicationCommandInteractionCreate, guildID string) {
	key, ok := keyByName(e.SlashCommandInteractionData().String(optionKey))
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(msgUnknownKey))
		return
	}

	n, err := resetValue(context.Background(), guildID, key.Name)
	if err != nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgDeleteError))
		return
	}
	if n == 0 {
		helpers.RespondEphemeral(e, fmt.Sprintf("Aucune valeur définie pour « %s ».", key.Description))
		return
	}

	helpers.RespondEphemeralCard(e, ui.Success(
		fmt.Sprintf("%s : valeur réinitialisée.", key.Description),
	))
}

func handleList(e *events.ApplicationCommandInteractionCreate, guildID string) {
	values, err := visibleValues(context.Background(), guildID)
	if err != nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgReadError))
		return
	}

	sorted := make([]Key, len(Keys))
	copy(sorted, Keys)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	lines := make([]string, 0, len(sorted))
	for _, k := range sorted {
		display := msgNotSet
		if raw, ok := values[k.Name]; ok {
			display = formatValue(k, raw)
		}
		lines = append(lines, fmt.Sprintf("**%s** (`%s`) — %s", k.Description, k.Name, display))
	}

	card := ui.New().Title("⚙️ Configuration du serveur")
	if len(lines) == 0 {
		card.Text("Aucun paramètre configurable.")
	} else {
		card.Text(strings.Join(lines, "\n"))
	}
	helpers.RespondEphemeralCard(e, card)
}

func requireAdmin(e *events.ApplicationCommandInteractionCreate) bool {
	if member := e.Member(); member != nil && member.Permissions.Has(discord.PermissionAdministrator) {
		return true
	}
	return helpers.RequirePermission(e, discord.PermissionManageGuild, msgNoPermission)
}

func requireGuild(e *events.ApplicationCommandInteractionCreate) (string, bool) {
	guildID := e.GuildID()
	if guildID == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgGuildOnly))
		return "", false
	}
	return guildID.String(), true
}
