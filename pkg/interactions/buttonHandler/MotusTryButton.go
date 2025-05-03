package buttonHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/game"
	"main/pkg/logger"

	"fmt"

	"github.com/bwmarrin/discordgo"
)

func MotusTryButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	wordLength := len(game.Word)

	// Open modal to get the user's guess
	modal := &discordgo.InteractionResponseData{
		CustomID: "motusTryModal",
		Title:    "Essai de motus",
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						CustomID:    "motusTryInput",
						Label:       "Votre essai",
						Style:       discordgo.TextInputShort,
						Placeholder: "Entrez un mot de " + fmt.Sprint(wordLength) + " lettres",
						MaxLength:   wordLength,
						MinLength:   wordLength,
					},
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: modal,
	})
	if err != nil {
		logger.ErrorLogger.Println("Error responding to interaction:", err)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{bot_utils.ErrorEmbed(s, err)},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
		}
	}

	return nil
}
