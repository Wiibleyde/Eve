package buttonHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/game"

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

	// Open modal to get the user's guess
	modal := &discordgo.InteractionResponseData{
		CustomID: "motusTryModal",
		Title:    "Essai de motus",
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.TextInput{
						CustomID:    "motusTryInput",
						Label:       "Entrez votre essai",
						Style:       discordgo.TextInputShort,
						Placeholder: "Entrez un mot de 5 lettres",
						MaxLength:   5,
						MinLength:   5,
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
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{bot_utils.ErrorEmbed(s, err)},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
	}

	return nil
}
