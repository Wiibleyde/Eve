package commandHandler

import (
	"main/pkg/bot_utils"

	"github.com/bwmarrin/discordgo"
)

func TalkHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	strMessage := options[0].StringValue()

	var mpUser *discordgo.User
	sendingToMp := false

	// Safely check if there's a second option for mpUser
	if len(options) > 1 {
		mpUser = options[1].UserValue(s)
		if mpUser != nil {
			sendingToMp = true
		}
	}

	if sendingToMp {
		channel, err := s.UserChannelCreate(mpUser.ID)
		if err != nil {
			return err
		}

		// Envoyer le message dans le canal DM
		_, err = s.ChannelMessageSend(channel.ID, strMessage)
		if err != nil {
			return err
		}
	}

	// Envoyer le message dans le canal de la commande
	_, err := s.ChannelMessageSend(i.ChannelID, strMessage)
	if err != nil {
		return err
	}

	// Acknowledge the interaction
	embed := bot_utils.SuccessEmbed(s, "Message envoyé")
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
