package bot

import (
	"fmt"
	"main/pkg/bot_utils"
	"main/pkg/config"
	"main/pkg/intelligence"
	"main/pkg/logger"
	"main/pkg/twitch"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if bot_utils.IsMaintenanceMode() && m.Author.ID != config.GetConfig().OwnerId {
		logger.WarningLogger.Println("Maintenance mode is enabled. Ignoring message from", m.Author.Username)
		bot_utils.MaintenanceModeEmbed(s, m)
		return
	}

	if m.GuildID != "" {
		if !m.MentionEveryone && bot_utils.UserContains(m.Mentions, s.State.User.ID) {
			// Send typing
			_ = s.ChannelTyping(m.ChannelID)

			aiAnswer, err := intelligence.SendMessageToAIChat(s, m.GuildID, m.Content, m.Author.ID)
			if err != nil {
				logger.ErrorLogger.Println("Error sending message to AI:", err)
				return
			}
			if aiAnswer != "" {
				_, err = s.ChannelMessageSend(m.ChannelID, aiAnswer)
				if err != nil {
					logger.ErrorLogger.Println("Error sending message to channel:", err)
				}
			}
		}
	}

	bot_utils.DetectFeur(s, m)
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	registerCommands(s)
	go startStatusChange(s)

	// Register stream handlers before starting the stream check
	RegisterStreamHandlers()
	go twitch.StartAutomaticStreamCheck()

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
				embed := bot_utils.ErrorEmbed(s, err.Error())
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
			embed := bot_utils.ErrorEmbed(s, "unknown command: "+i.ApplicationCommandData().Name)
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
		// Split customID at --
		customId := strings.Split(i.MessageComponentData().CustomID, "--")[0]
		if handler, ok := componentHandlers[customId]; ok {
			err := handler(s, i)
			if err != nil {
				embed := bot_utils.ErrorEmbed(s, err.Error())
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
			embed := bot_utils.ErrorEmbed(s, "unknown component: "+i.MessageComponentData().CustomID)
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
				embed := bot_utils.ErrorEmbed(s, err.Error())
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
			embed := bot_utils.ErrorEmbed(s, "unknown modal: "+i.ModalSubmitData().CustomID)
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
