package commandHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"

	"github.com/bwmarrin/discordgo"
)

func StreamerHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "add":
		twitchChannelName := subCommand.Options[0].StringValue()
		announceChannel := subCommand.Options[1].ChannelValue(s)
		var announceRole *discordgo.Role
		if len(subCommand.Options) > 2 {
			announceRole = subCommand.Options[2].RoleValue(s, i.GuildID)
		}

		client, ctx := data.GetDBClient()
		streamer, err := client.Stream.FindFirst(
			db.Stream.And(
				db.Stream.TwitchChannelName.Equals(twitchChannelName),
				db.Stream.GuildID.Equals(i.GuildID),
			),
		).Exec(ctx)
		if err != nil && err.Error() != "ErrNotFound" {
			logger.ErrorLogger.Println("Error while checking if streamer exists:", err)
			embed := bot_utils.ErrorEmbed(s, "Erreur lors de la vérification de l'existence du streamer")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error while responding to interaction:", err)
				return err
			}
			return err
		}
		if streamer != nil {
			embed := bot_utils.WarningEmbed(s, "Le streamer existe déjà dans la base de données.")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error while responding to interaction:", err)
				return err
			}
			return nil
		}

		// Create the streamer
		if announceRole == nil {
			_, err = client.Stream.CreateOne(
				db.Stream.GuildID.Set(i.GuildID),
				db.Stream.ChannelID.Set(announceChannel.ID),
				db.Stream.TwitchChannelName.Set(twitchChannelName),
			).Exec(ctx)
			if err != nil {
				logger.ErrorLogger.Println("Error while creating streamer:", err)
				embed := bot_utils.ErrorEmbed(s, "Erreur lors de la création du streamer")
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					logger.ErrorLogger.Println("Error while responding to interaction:", err)
					return err
				}
			}
		} else {
			_, err = client.Stream.CreateOne(
				db.Stream.GuildID.Set(i.GuildID),
				db.Stream.ChannelID.Set(announceChannel.ID),
				db.Stream.TwitchChannelName.Set(twitchChannelName),
				db.Stream.RoleID.Set(announceRole.ID),
			).Exec(ctx)
			if err != nil {
				logger.ErrorLogger.Println("Error while creating streamer:", err)
				embed := bot_utils.ErrorEmbed(s, "Erreur lors de la création du streamer")
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					logger.ErrorLogger.Println("Error while responding to interaction:", err)
					return err
				}
			}
		}

		// Send success message
		embed := bot_utils.SuccessEmbed(s, "Streamer ajouté avec succès !")
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error while responding to interaction:", err)
			return err
		}
	case "remove":
		twitchChannelName := subCommand.Options[0].StringValue()
		client, ctx := data.GetDBClient()
		rowToDelete, err := client.Stream.FindFirst(
			db.Stream.And(
				db.Stream.TwitchChannelName.Equals(twitchChannelName),
				db.Stream.GuildID.Equals(i.GuildID),
			),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error while checking if streamer exists:", err)
			embed := bot_utils.ErrorEmbed(s, "Erreur lors de la vérification de l'existence du streamer")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error while responding to interaction:", err)
				return err
			}
			return err
		}
		if rowToDelete == nil {
			embed := bot_utils.WarningEmbed(s, "Le streamer n'existe pas dans la base de données.")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error while responding to interaction:", err)
				return err
			}
			return nil
		}
		// Delete the streamer
		_, err = client.Stream.FindUnique(
			db.Stream.UUID.Equals(rowToDelete.UUID),
		).Delete().Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error while deleting streamer:", err)
			embed := bot_utils.ErrorEmbed(s, "Erreur lors de la suppression du streamer")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error while responding to interaction:", err)
				return err
			}
		}
		// Send success message
		embed := bot_utils.SuccessEmbed(s, "Streamer supprimé avec succès !")
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error while responding to interaction:", err)
			return err
		}

	}
	return nil
}
