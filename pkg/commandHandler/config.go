package commandHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"

	"github.com/bwmarrin/discordgo"
)

func ConfigHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(i.ApplicationCommandData().Options) == 0 {
		return errors.New("no subcommand provided")
	}

	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "set":
		if len(subCommand.Options) < 2 {
			return errors.New("missing required options for set command")
		}

		var strConfig string

		if subCommand.Options[0].Type == discordgo.ApplicationCommandOptionString {
			strConfig = subCommand.Options[0].StringValue()
		} else {
			return errors.New("config key must be a string")
		}

		channelValue := subCommand.Options[1].ChannelValue(s)

		client, ctx := data.GetDBClient()
		actualData, err := client.Config.FindMany(
			db.Config.GuildID.Equals(i.GuildID),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error updating config:", err)
			return err
		}
		if len(actualData) == 0 {
			_, err = client.Config.CreateOne(
				db.Config.Key.Set(strConfig),
				db.Config.Value.Set(channelValue.ID),
				db.Config.GuildID.Set(i.GuildID),
			).Exec(ctx)
			if err != nil {
				logger.ErrorLogger.Println("Error creating config:", err)
				return err
			}
		} else {
			_, err = client.Config.FindMany(
				db.Config.And(
					db.Config.GuildID.Equals(i.GuildID),
					db.Config.Key.Equals(strConfig),
				),
			).Update(
				db.Config.Value.Set(channelValue.ID),
			).Exec(ctx)
			if err != nil {
				logger.ErrorLogger.Println("Error updating config:", err)
				return err
			}
		}

		embed := bot_utils.SuccessEmbed(s, "Configuration mise à jour avec succès.")
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		return err
	case "get":
		client, ctx := data.GetDBClient()
		configs, err := client.Config.FindMany(
			db.Config.GuildID.Equals(i.GuildID),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error fetching config:", err)
			return err
		}
		if len(configs) == 0 {
			embed := bot_utils.WarningEmbed(s, "Aucune configuration trouvée pour ce serveur.")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to interaction:", err)
				return err
			}
			return nil
		}
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Configurations du serveur"
		embed.Description = "Voici les configurations actuelles du serveur :"
		for _, config := range configs {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   config.Key,
				Value:  config.Value,
				Inline: false,
			})
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}
		return nil
	case "delete":
		if len(subCommand.Options) < 1 {
			return errors.New("missing required options for delete command")
		}
		// Safely access options
		var strConfig string
		if subCommand.Options[0].Type == discordgo.ApplicationCommandOptionString {
			strConfig = subCommand.Options[0].StringValue()
		} else {
			return errors.New("config key must be a string")
		}
		client, ctx := data.GetDBClient()
		_, err := client.Config.FindMany(
			db.Config.GuildID.Equals(i.GuildID),
			db.Config.Key.Equals(strConfig),
		).Delete().Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error deleting config:", err)
			return err
		}
		embed := bot_utils.SuccessEmbed(s, "Configuration supprimée avec succès.")
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}
	}
	return nil
}
