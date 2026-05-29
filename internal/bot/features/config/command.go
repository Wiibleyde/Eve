package config

import (
	"context"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/database"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "config",
		Description: "Configuration du serveur",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "birthday-channel",
				Description: "[ADMIN] Définir le salon des anniversaires",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "Salon cible",
						Required:    true,
						ChannelTypes: []discord.ChannelType{
							discord.ChannelTypeGuildText,
						},
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "quote-channel",
				Description: "[ADMIN] Définir le salon des citations",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "Salon cible",
						Required:    true,
						ChannelTypes: []discord.ChannelType{
							discord.ChannelTypeGuildText,
						},
					},
				},
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
	case "birthday-channel":
		handleSetBirthdayChannel(e)
	case "quote-channel":
		handleSetQuoteChannel(e)
	}
}

func handleSetBirthdayChannel(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.RequirePermission(e, discord.PermissionAdministrator, "Vous n'avez pas la permission d'utiliser cette commande.") {
		return
	}
	if e.GuildID() == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Cette commande doit être utilisée dans un serveur."))
		return
	}

	data := e.SlashCommandInteractionData()
	channel := data.Channel("channel")
	guildID := e.GuildID().String()
	channelID := channel.ID.String()
	key := tables.BirthdayChannel.String()

	ctx := context.Background()

	n, err := database.Default.Ent().GuildConfig.Update().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(key)).
		SetValue(channelID).
		Save(ctx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
		return
	}

	if n == 0 {
		err = database.Default.Ent().GuildConfig.Create().
			SetGuildID(guildID).
			SetKey(key).
			SetValue(channelID).
			Exec(ctx)
		if err != nil {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
			return
		}
	}

	embed := embeds.SuccessEmbed("Salon des anniversaires défini sur <#" + channelID + ">.")
	helpers.RespondEphemeralEmbed(e, embed)
}

func handleSetQuoteChannel(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.RequirePermission(e, discord.PermissionAdministrator, "Vous n'avez pas la permission d'utiliser cette commande.") {
		return
	}
	if e.GuildID() == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Cette commande doit être utilisée dans un serveur."))
		return
	}

	data := e.SlashCommandInteractionData()
	channel := data.Channel("channel")
	guildID := e.GuildID().String()
	channelID := channel.ID.String()
	key := tables.QuoteChannel.String()

	ctx := context.Background()

	n, err := database.Default.Ent().GuildConfig.Update().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(key)).
		SetValue(channelID).
		Save(ctx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
		return
	}

	if n == 0 {
		err = database.Default.Ent().GuildConfig.Create().
			SetGuildID(guildID).
			SetKey(key).
			SetValue(channelID).
			Exec(ctx)
		if err != nil {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
			return
		}
	}

	embed := embeds.SuccessEmbed("Salon des anniversaires défini sur <#" + channelID + ">.")
	helpers.RespondEphemeralEmbed(e, embed)
}
