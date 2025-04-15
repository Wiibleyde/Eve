package bot

import (
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
}

func OnReady(s *discordgo.Session, r *discordgo.Ready) {
	RegisterCommands(s)
	logger.InfoLogger.Println("Bot is ready!")
}

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		if handler, ok := CommandHandlers[i.ApplicationCommandData().Name]; ok {
			err := handler(s, i)
			if err != nil {
				logger.ErrorLogger.Println("Error handling command", i.ApplicationCommandData().Name, err)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error handling command: " + err.Error(),
					},
				})
				if err != nil {
					logger.ErrorLogger.Println("Error responding to interaction:", err)
				}
			}
		} else {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknown command",
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to unknown command interaction:", err)
			}
		}
	}
}
