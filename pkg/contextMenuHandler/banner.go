package contextMenuHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func BannerContextMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userID := i.ApplicationCommandData().TargetID

	user, err := s.User(userID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting user:", err)
		return err
	}

	// Get the user's banner URL
	bannerURL := user.BannerURL("1024")
	if bannerURL == "" {
		logger.ErrorLogger.Println("User has no banner")
		return errors.New("this user has no banner")
	}

	embed := bot_utils.BasicEmbedBuilder(s)

	embed.Title = "Bannière demandée"
	embed.Description = "Voici la bannière de " + user.Username
	embed.Image = &discordgo.MessageEmbedImage{URL: bannerURL}
	embed.Color = 0x00FF00

	// Send the banner URL as a message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.ErrorLogger.Println("Error sending interaction response:", err)
		return err
	}
	return nil
}
