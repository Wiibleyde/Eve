package contextMenuHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func ProfilePictureContextMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userID := i.ApplicationCommandData().TargetID

	user, err := s.User(userID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting user:", err)
		return err
	}

	// Get the user's avatar URL
	avatarURL := user.AvatarURL("1024")
	if avatarURL == "" {
		logger.ErrorLogger.Println("User has no avatar")
		return errors.New("this user has no avatar")
	}

	embed := bot_utils.BasicEmbedBuilder(s)

	embed.Title = "Avatar demandé"
	embed.Description = "Voici l'avatar de " + user.Username
	embed.Image = &discordgo.MessageEmbedImage{URL: avatarURL}
	embed.Color = 0x00FF00

	// Send the avatar URL as a message
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
