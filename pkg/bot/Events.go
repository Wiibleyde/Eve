package bot

import (
	"fmt"
	"main/pkg/bot_utils"
	"main/pkg/logger"
	"strings"

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
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			err := handler(s, i)
			if err != nil {
				logger.ErrorLogger.Println("Error handling command", i.ApplicationCommandData().Name, err)
				embed := bot_utils.ErrorEmbed(s, err)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					logger.ErrorLogger.Println("Error responding to interaction:", err)
				}
			}
			args := i.ApplicationCommandData().Options
			if len(args) > 0 {
				var formattedArgs []string
				for _, arg := range args {
					var value interface{}
					switch arg.Type {
					case discordgo.ApplicationCommandOptionString:
						value = arg.StringValue()
					case discordgo.ApplicationCommandOptionInteger:
						value = arg.IntValue()
					case discordgo.ApplicationCommandOptionBoolean:
						value = arg.BoolValue()
					case discordgo.ApplicationCommandOptionUser:
						value = arg.UserValue(s).Username
					case discordgo.ApplicationCommandOptionChannel:
						value = arg.ChannelValue(s).Name
					case discordgo.ApplicationCommandOptionRole:
						value = arg.RoleValue(s, i.GuildID).Name
					default:
						value = "unknown"
					}
					formattedArgs = append(formattedArgs, arg.Name+"="+fmt.Sprint(value))
				}
				logger.InfoLogger.Printf("Command executed: %s by %s (Args: %s)\n",
					i.ApplicationCommandData().Name,
					i.Member.User.Username,
					strings.Join(formattedArgs, ", "))
			} else {
				logger.InfoLogger.Println("Command executed:", i.ApplicationCommandData().Name, "by", i.Member.User.Username)
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
	} else if i.Type == discordgo.InteractionMessageComponent {
		if handler, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
			err := handler(s, i)
			if err != nil {
				logger.ErrorLogger.Println("Error handling component", i.MessageComponentData().CustomID, err)
				embed := bot_utils.ErrorEmbed(s, err)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
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
					Content: "Unknown component",
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to unknown component interaction:", err)
			}
		}
	}
}
