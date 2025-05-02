package modalHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/game"

	"github.com/bwmarrin/discordgo"
)

func MotusTryModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get the message ID from the interaction
	messageID := i.Message.ID

	// Get the game associated with the message ID
	game := game.GetMotusGame(messageID)
	if game == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{bot_utils.ErrorEmbed(s, errors.New("aucune partie de motus trouvé"))},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
	}

	//TODO: Update the game embed with the user's guess
	embed := game.GetEmbed(s)
	components := game.GetComponents()
	

	return nil
}
