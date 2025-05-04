package modalHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/game"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func MotusTryModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get the message ID from the interaction
	messageID := i.Message.ID

	// Get the game associated with the message ID
	currentGame := game.GetMotusGame(messageID)
	if currentGame == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{bot_utils.ErrorEmbed(s, "Jeu de motus introuvable.")},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
	}

	data := i.ModalSubmitData()
	wordGiven := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	currentGame.HandleTry(wordGiven, i.Member.User.GlobalName)

	embed := currentGame.GetEmbed(s)
	var components []discordgo.MessageComponent
	if currentGame.GameState == game.MotusGameStatePlaying {
		components = currentGame.GetComponents()

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
				Flags:      discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}
	} else {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: nil,
				Flags:      discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}
	}

	return nil
}
