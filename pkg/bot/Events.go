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

// Helper functions to safely get user information
func safeGetUsername(user *discordgo.User) string {
	if user == nil {
		return "unknown_user"
	}
	return user.Username
}

func safeGetUserID(user *discordgo.User) string {
	if user == nil {
		return "unknown_id"
	}
	return user.ID
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if bot_utils.IsMaintenanceMode() && m.Author.ID != config.GetConfig().OwnerId {
		logger.WarningLogger.Println("Maintenance mode is enabled. Ignoring message from", safeGetUsername(m.Author))
		bot_utils.MaintenanceModeEmbed(s, m)
		return
	}

	if m.GuildID != "" { // Check if the message is in a guild
		if m.GuildID == config.GetConfig().EveHomeGuild {
			checkAndForwardThreadMessage(s, m)
		}

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

		bot_utils.DetectFeur(s, m)
	} else { // Check if the message is in a DM
		ReceiveMessage(s, m)
	}
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	registerCommands(s)
	InitMpThreadManager()

	go startStatusChange(s)

	// Register stream handlers before starting the stream check
	RegisterStreamHandlers()
	go twitch.StartAutomaticStreamCheck()

	logger.InfoLogger.Println("Bot is ready!")
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if Member exists before accessing User
	var userID, username string
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
		username = i.Member.User.Username
	} else if i.User != nil {
		userID = i.User.ID
		username = i.User.Username
	} else {
		userID = "unknown_id"
		username = "unknown_user"
	}

	if bot_utils.IsMaintenanceMode() && userID != config.GetConfig().OwnerId {
		logger.WarningLogger.Println("Maintenance mode is enabled. Ignoring interaction from", username)
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
				logger.InfoLogger.Println("User command executed:", i.ApplicationCommandData().Name, "by", username+" on "+i.ApplicationCommandData().TargetID)
			} else if cmdType == discordgo.MessageApplicationCommand {
				logger.InfoLogger.Println("Message command executed:", i.ApplicationCommandData().Name, "by", username+" on "+i.ApplicationCommandData().TargetID)
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
							user := arg.UserValue(s)
							value = safeGetUsername(user)
						case discordgo.ApplicationCommandOptionChannel:
							channel := arg.ChannelValue(s)
							if channel != nil {
								value = channel.Name
							} else {
								value = "unknown_channel"
							}
						case discordgo.ApplicationCommandOptionRole:
							role := arg.RoleValue(s, i.GuildID)
							if role != nil {
								value = role.Name
							} else {
								value = "unknown_role"
							}
						default:
							value = "unknown"
						}
						formattedArgs = append(formattedArgs, arg.Name+"="+fmt.Sprint(value))
					}
					logger.InfoLogger.Println("Commande", i.ApplicationCommandData().Name, "par", username, "<@"+userID+"> dans le serveur", i.GuildID, "(<#"+i.ChannelID+">) avec les arguments: %v\n", strings.Join(formattedArgs, ", "))
				} else {
					if i.GuildID != "" {
						logger.InfoLogger.Println("Commande", i.ApplicationCommandData().Name, "par", username, "<@"+userID+"> dans le serveur", i.GuildID, "(<#"+i.ChannelID+">)")
					} else {
						logger.InfoLogger.Println("Commande", i.ApplicationCommandData().Name, "par", username, "<@"+userID+">")
					}
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

			logger.InfoLogger.Println("Component interaction executed:", i.MessageComponentData().CustomID, "by", username, "<@"+userID+">")
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

			logger.InfoLogger.Println("Modal interaction executed:", i.ModalSubmitData().CustomID, "by", username, "<@"+userID+">")
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
