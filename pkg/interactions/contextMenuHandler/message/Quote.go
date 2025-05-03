package messageContextMenuHandler

import (
	"fmt"
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/image_generator"
	"main/pkg/logger"
	"main/prisma/db"
	"os"

	"github.com/bwmarrin/discordgo"
)

func CreateQuoteMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get the message ID from the interaction
	messageID := i.ApplicationCommandData().TargetID

	// Get the channel ID from the interaction
	channelID := i.ChannelID

	// Get the message from the channel
	message, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting message:", err)
		return err
	}

	avatarURL := message.Author.AvatarURL("1024")
	year, month, day := message.Timestamp.Date()
	formattedDate := fmt.Sprintf("%02d/%02d/%04d", day, int(month), year)
	path, err := image_generator.CreateQuote(message.Content, message.Author.GlobalName, "", formattedDate, avatarURL)
	if err != nil {
		return err
	}

	fileReader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	client, ctx := data.GetDBClient()
	quoteChannelConfig, err := client.Config.FindMany(
		db.Config.And(
			db.Config.GuildID.Equals(i.GuildID),
			db.Config.Key.Equals("quoteChannel"),
		),
	).Exec(ctx)
	if err != nil {
		return err
	}

	messageToSend := "\"" + message.Content + "\" - <@" + message.Author.ID + "> - " + formattedDate

	// Create the file for sending
	file := &discordgo.File{
		Name:   "quote.png",
		Reader: fileReader,
	}

	// Prepare message data with both text and image attachment
	messageData := &discordgo.MessageSend{
		Content: messageToSend,
		Files:   []*discordgo.File{file},
	}

	if len(quoteChannelConfig) > 0 {
		channelID = quoteChannelConfig[0].Value
	}

	// Send the message with the image
	_, err = s.ChannelMessageSendComplex(channelID, messageData)
	if err != nil {
		return err
	}

	embed := bot_utils.SuccessEmbed(s, "Citation envoyé avec succès !")
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
