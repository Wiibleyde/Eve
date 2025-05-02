package commandHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/game"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func MotusCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	newGame := game.NewMotusGame()
	embed := newGame.GetEmbed(s)
	components := newGame.GetComponents()

	message, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})
	if err != nil {
		logger.ErrorLogger.Println("Error sending message:", err)
		return err
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{bot_utils.SuccessEmbed(s, "Jeu de motus lancé !")},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.ErrorLogger.Println("Error responding to interaction:", err)
		return err
	}

	game.AddMotusGame(message.ID, newGame)

	return nil
}
