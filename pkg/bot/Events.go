package bot

import (
	"fmt"
	"main/pkg/bot_utils"
	"main/pkg/config"
	"main/pkg/intelligence"
	"main/pkg/logger"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.GuildID != "" {
		logger.DebugLogger.Println(intelligence.SendMessageToAIChat(s, m.GuildID, m.Content, m.Author.ID))
	}
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	registerCommands(s)
	go startStatusChange(s)
	logger.InfoLogger.Println("Bot is ready!")
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if bot_utils.IsMaintenanceMode() && i.Member.User.ID != config.GetConfig().OwnerId {
		logger.WarningLogger.Println("Maintenance mode is enabled. Ignoring interaction from", i.Member.User.Username)
		bot_utils.MaintenanceModeEmbed(s, i)
		return
	}
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
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
					logger.ErrorLogger.Println("Error of errors responding to interaction:", err)
				}
			}
			// Log based on the command type
			cmdType := i.ApplicationCommandData().CommandType
			if cmdType == discordgo.UserApplicationCommand {
				logger.InfoLogger.Println("User command executed:", i.ApplicationCommandData().Name, "by", i.Member.User.Username+" on "+i.ApplicationCommandData().TargetID)
			} else if cmdType == discordgo.MessageApplicationCommand {
				logger.InfoLogger.Println("Message command executed:", i.ApplicationCommandData().Name, "by", i.Member.User.Username+" on "+i.ApplicationCommandData().TargetID)
			} else {
				// For regular slash commands
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
			}
		} else {
			embed := bot_utils.ErrorEmbed(s, fmt.Errorf("unknown command: %s", i.ApplicationCommandData().Name))
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to unknown command interaction:", err)
			}
		}
	case discordgo.InteractionMessageComponent:
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
			embed := bot_utils.ErrorEmbed(s, fmt.Errorf("unknown component: %s", i.MessageComponentData().CustomID))
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to unknown component interaction:", err)
			}
		}
	case discordgo.InteractionModalSubmit:
		if handler, ok := modalHandlers[i.ModalSubmitData().CustomID]; ok {
			err := handler(s, i)
			if err != nil {
				logger.ErrorLogger.Println("Error handling modal", i.ModalSubmitData().CustomID, err)
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
			embed := bot_utils.ErrorEmbed(s, fmt.Errorf("unknown modal: %s", i.ModalSubmitData().CustomID))
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to unknown modal interaction:", err)
			}
		}
	default:
		logger.WarningLogger.Printf("Unhandled interaction type: %v\n", i.Type)
	}
}
