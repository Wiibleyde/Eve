package commandHandler

import (
	"errors"
	"main/pkg/data"
	"main/pkg/image_generator"
	"main/pkg/logger"
	"main/prisma/db"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func QuoteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	strQuote := i.ApplicationCommandData().Options[0].StringValue()
	author := i.ApplicationCommandData().Options[1].UserValue(s)
	context := ""
	if len(i.ApplicationCommandData().Options) > 2 {
		context = i.ApplicationCommandData().Options[2].StringValue()
	}

	// Check if the quote is empty
	if strQuote == "" {
		logger.ErrorLogger.Println("Quote is empty")
		return errors.New("quote cannot be empty")
	}

	today := time.Now()
	// FOrmat to DD/MM/YYYY
	formattedDate := today.Format("02/01/2006")

	// Get user profile picture
	avatarURL := author.AvatarURL("1024")
	path, err := image_generator.CreateQuote(strQuote, author.GlobalName, context, formattedDate, avatarURL)
	if err != nil {
		logger.ErrorLogger.Println("Error creating quote:", err)
		return err
	}

	// Open the file for reading
	fileReader, err := os.Open(path)
	if err != nil {
		logger.ErrorLogger.Println("Error opening quote image file:", err)
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
		logger.ErrorLogger.Println("Error retrieving quote channel config:", err)
		return err
	}

	messageToSend := "\"" + strQuote + "\" - <@" + author.ID + "> - " + formattedDate

	if context != "" {
		messageToSend += "\n*" + context + "*"
	}

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

	// Determine which channel to send the message to
	channelID := i.ChannelID
	if len(quoteChannelConfig) > 0 {
		channelID = quoteChannelConfig[0].Value
	}

	// Send the message with the image
	_, err = s.ChannelMessageSendComplex(channelID, messageData)
	if err != nil {
		logger.ErrorLogger.Println("Error sending message with quote image:", err)
		return err
	}

	return nil
}
