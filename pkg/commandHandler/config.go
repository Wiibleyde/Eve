package commandHandler

import (
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"

	"github.com/bwmarrin/discordgo"
)

func ConfigHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check if we have subcommands
	if len(i.ApplicationCommandData().Options) == 0 {
		return respondWithError(s, i, "No subcommand provided")
	}

	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "set":
		// Check if we have the required options for the set command
		if len(subCommand.Options) < 2 {
			return respondWithError(s, i, "Missing config key or value")
		}

		// Safely access options
		var strConfig, strValue string

		if subCommand.Options[0].Type == discordgo.ApplicationCommandOptionString {
			strConfig = subCommand.Options[0].StringValue()
		} else {
			return respondWithError(s, i, "Config key must be a string")
		}

		if subCommand.Options[1].Type == discordgo.ApplicationCommandOptionString {
			strValue = subCommand.Options[1].StringValue()
		} else {
			return respondWithError(s, i, "Config value must be a string")
		}

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
				db.Config.Value.Set(strValue),
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
				db.Config.Value.Set(strValue),
			).Exec(ctx)
			if err != nil {
				logger.ErrorLogger.Println("Error updating config:", err)
				return err
			}
		}

		embed := &discordgo.MessageEmbed{
			Description: "Configuration mise à jour avec succès.",
			Color:       0x00FF00,
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
		return err
	}
	return nil
}

// Helper function to respond with error messages
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	embed := &discordgo.MessageEmbed{
		Description: "Error: " + message,
		Color:       0xFF0000, // Red color for errors
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
