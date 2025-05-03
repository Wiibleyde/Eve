package buttonHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func JokeSetPublicButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	message := i.Message.Embeds[0]
	joke := message.Description
	answer := message.Fields[0].Value
	embed := bot_utils.BasicEmbedBuilder(s)
	embed.Title = "Blague"
	embed.Description = joke
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "Réponse",
			Value: answer,
		},
	}
	embed.Footer.Text = "Eve et ses développeurs ne sont pas responsable des blagues affichées."
	embed.Color = 0x00FF00

	channel := i.Message.ChannelID

	_, err := s.ChannelMessageSendEmbed(channel, embed)
	if err != nil {
		logger.ErrorLogger.Println("Error sending message:", err)
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{bot_utils.SuccessEmbed(s, "Blague affichée publiquement !")},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
