package bot

import (
	"fmt"
	"main/pkg/data"
	"main/pkg/events"
	"main/pkg/logger"
	"main/prisma/db"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// BotStreamHandler implements the events.StreamEventHandler interface
type BotStreamHandler struct{}

// OnNewStream handles new stream events
func (h *BotStreamHandler) OnNewStream(stream events.StreamData, userData events.UserData) {
	BotOnNewStream(stream, userData)
}

// OnStreamEnd handles stream end events
func (h *BotStreamHandler) OnStreamEnd(stream events.StreamData, userData events.UserData) {
	BotOnStreamEnd(stream, userData)
}

// RegisterStreamHandlers registers the bot's stream handlers
func RegisterStreamHandlers() {
	handler := &BotStreamHandler{}
	events.RegisterStreamEventHandler(handler)
}

func generateOnlineTwitchEmbed(s *discordgo.Session, stream events.StreamData, userData events.UserData) *discordgo.MessageEmbed {
	startTime, err := time.Parse(time.RFC3339, stream.StartedAt)
	if err != nil {
		logger.ErrorLogger.Println("Error parsing stream start time:", err)
		return nil
	}

	// Fix the URL formatting - use string replacement instead of fmt.Sprintf
	// Replace {width} and {height} placeholders with actual values
	thumbnailURL := stream.ThumbnailURL
	thumbnailURL = strings.Replace(thumbnailURL, "{width}", "1280", -1)
	thumbnailURL = strings.Replace(thumbnailURL, "{height}", "720", -1)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    userData.DisplayName,
			IconURL: userData.ProfileImageURL,
			URL:     "https://www.twitch.tv/" + stream.UserLogin,
		},
		Title:       stream.UserName,
		Description: stream.UserName + " est en live !",
		URL:         "https://www.twitch.tv/" + stream.UserLogin,
		Color:       0x9146FF,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://static-cdn.jtvnw.net/ttv-boxart/" + stream.GameID + ".jpg",
		},
		Image: &discordgo.MessageEmbedImage{
			URL: thumbnailURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Eve – Toujours prête à vous aider.",
			IconURL: s.State.User.AvatarURL("64"),
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Titre :",
				Value:  stream.Title,
				Inline: true,
			},
			{
				Name:   "Entrain de :",
				Value:  stream.GameName,
				Inline: true,
			},
			{
				Name:   "Démarré :",
				Value:  "<t:" + fmt.Sprintf("%d", startTime.Unix()) + ":R>",
				Inline: true,
			},
		},
	}
	return embed
}

func generateOfflineTwitchEmbed(s *discordgo.Session, userData events.UserData) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    userData.DisplayName,
			IconURL: userData.ProfileImageURL,
			URL:     "https://www.twitch.tv/" + userData.Login,
		},
		Title:       userData.DisplayName,
		Description: "Le stream est hors ligne.",
		URL:         "https://www.twitch.tv/" + userData.Login,
		Color:       0x4c2387,
		Image: &discordgo.MessageEmbedImage{
			URL: userData.OfflineImageURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Eve – Toujours prête à vous aider.",
			IconURL: s.State.User.AvatarURL("64"),
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
	return embed
}

func BotOnNewStream(stream events.StreamData, userData events.UserData) {
	client, ctx := data.GetDBClient()
	databaseStreamData, err := client.Stream.FindFirst(
		db.Stream.TwitchChannelName.Equals(stream.UserLogin),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error while checking if streamer exists:", err)
		return
	}
	if databaseStreamData == nil {
		logger.ErrorLogger.Println("Streamer not found in database:", stream.UserLogin)
		return
	}

	// Check if the streamer is already online
	messageID, valid := databaseStreamData.MessageID()
	if valid {
		currentMessage, err := session.ChannelMessage(databaseStreamData.ChannelID, messageID)
		if err != nil {
			logger.ErrorLogger.Println("Error while getting message:", err)
			return
		}
		embed := currentMessage.Embeds[0]
		if embed.Description == stream.UserName+" est en live !" {
			logger.InfoLogger.Println("Streamer is already online:", stream.UserLogin)
			return
		}

		err = session.ChannelMessageDelete(databaseStreamData.ChannelID, messageID)
		if err != nil {
			logger.ErrorLogger.Println("Error while deleting message:", err)
			return
		}
	}

	finalEmbed := generateOnlineTwitchEmbed(session, stream, userData)

	var newMessage *discordgo.Message

	roleId, valid := databaseStreamData.RoleID()
	if valid {
		newMessage, err = session.ChannelMessageSendComplex(databaseStreamData.ChannelID, &discordgo.MessageSend{
			Embed:   finalEmbed,
			Content: "<@&" + roleId + ">",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "Regarder le stream",
							Style:    discordgo.LinkButton,
							URL:      "https://www.twitch.tv/" + stream.UserLogin,
							Disabled: false,
						},
					},
				},
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error while editing message:", err)
			return
		}
	} else {
		newMessage, err = session.ChannelMessageSendComplex(databaseStreamData.ChannelID, &discordgo.MessageSend{
			Embed: finalEmbed,
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "Regarder le stream",
							Style:    discordgo.LinkButton,
							URL:      "https://www.twitch.tv/" + stream.UserLogin,
							Disabled: false,
						},
					},
				},
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error while editing message:", err)
			return
		}
	}

	// Update the message in the database
	_, err = client.Stream.FindUnique(
		db.Stream.UUID.Equals(databaseStreamData.UUID),
	).Update(
		db.Stream.MessageID.Set(newMessage.ID),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error while updating message ID in database:", err)
		return
	}
}

func BotOnStreamEnd(stream events.StreamData, userData events.UserData) {
	client, ctx := data.GetDBClient()
	databaseStreamData, err := client.Stream.FindFirst(
		db.Stream.TwitchChannelName.Equals(stream.UserLogin),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error while checking if streamer exists:", err)
		return
	}
	if databaseStreamData == nil {
		logger.ErrorLogger.Println("Streamer not found in database:", stream.UserLogin)
		return
	}

	messageID, valid := databaseStreamData.MessageID()
	if !valid {
		logger.InfoLogger.Println("Streamer is already offline:", stream.UserLogin)
		return
	}

	currentMessage, err := session.ChannelMessage(databaseStreamData.ChannelID, messageID)
	if err != nil {
		logger.ErrorLogger.Println("Error while getting message:", err)
		return
	}
	embed := currentMessage.Embeds[0]
	if embed.Description == "Le stream est hors ligne." {
		logger.InfoLogger.Println("Streamer is already offline:", stream.UserLogin)
		return
	}

	finalEmbed := generateOfflineTwitchEmbed(session, userData)

	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "Voir la chaîne",
					Style:    discordgo.LinkButton,
					URL:      "https://www.twitch.tv/" + stream.UserLogin,
					Disabled: true,
				},
			},
		},
	}
	emptyString := ""
	_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Content:    &emptyString,
		Channel:    databaseStreamData.ChannelID,
		Embed:      finalEmbed,
		Components: &components,
	})
	if err != nil {
		logger.ErrorLogger.Println("Error while editing message:", err)
		return
	}
}
