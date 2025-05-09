package bot

import (
	"fmt"
	"main/pkg/bot_utils"
	"main/pkg/config"
	"main/pkg/intelligence"
	"main/pkg/logger"
	"main/pkg/twitch"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Helper functions to safely get user information
func safeGetUsername(user *discordgo.User) string {
	if user == nil {
		return "utilisateur_inconnu"
	}
	return user.Username
}

// logExecutionTime wraps a function and returns a non-blocking logging function that includes execution time
func logExecutionTime(fn func()) func(format string, args ...interface{}) {
	start := time.Now()
	fn()
	elapsed := time.Since(start)

	return func(format string, args ...interface{}) {
		// Run the actual logging in a separate goroutine to avoid blocking
		go func(elapsed time.Duration, format string, args ...interface{}) {
			timeStr := fmt.Sprintf("[%s] ", elapsed)
			logger.InfoLogger.Printf(timeStr+format, args...)
		}(elapsed, format, args...)
	}
}

// handleInteractionError responds to an interaction with an error message
func handleInteractionError(s *discordgo.Session, i *discordgo.InteractionCreate, errMsg string) {
	embed := bot_utils.ErrorEmbed(s, errMsg)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.ErrorLogger.Println("Erreur lors de la réponse à l'interaction:", err)
	}
}

// getUserInfo extracts user information from an interaction
func getUserInfo(i *discordgo.InteractionCreate) (string, string) {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID, i.Member.User.Username
	} else if i.User != nil {
		return i.User.ID, i.User.Username
	}
	return "id_inconnu", "utilisateur_inconnu"
}

// formatCommandArg formats a command argument for logging
func formatCommandArg(s *discordgo.Session, guildID string, arg *discordgo.ApplicationCommandInteractionDataOption) string {
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
			value = "salon_inconnu"
		}
	case discordgo.ApplicationCommandOptionRole:
		role := arg.RoleValue(s, guildID)
		if role != nil {
			value = role.Name
		} else {
			value = "rôle_inconnu"
		}
	default:
		value = "inconnu"
	}
	return arg.Name + "=" + fmt.Sprint(value)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Early returns without logging execution time
	if m.Author.Bot {
		return
	}

	if bot_utils.IsMaintenanceMode() && m.Author.ID != config.GetConfig().OwnerId {
		logger.WarningLogger.Println("Mode maintenance activé. Message ignoré de", safeGetUsername(m.Author))
		bot_utils.MaintenanceModeEmbed(s, m)
		return
	}

	// Only measure execution time when the bot actually processes the message
	if m.GuildID != "" { // Message in a guild
		if m.GuildID == config.GetConfig().EveHomeGuild ||
			(!m.MentionEveryone && bot_utils.UserContains(m.Mentions, s.State.User.ID)) {
			// Process in a wrapper that captures all results for unified logging
			processGuildMessage(s, m)
		} else {
			// Process messages that don't need timing
			bot_utils.DetectFeur(s, m)
		}
	} else { // Message in a DM
		log := logExecutionTime(func() {
			ReceiveMessage(s, m)
		})
		log("Message privé traité de %s", safeGetUsername(m.Author))
	}
}

// processGuildMessage handles guild messages with unified logging
func processGuildMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Variables to capture for logging
	var aiRequest string
	var aiResponse string
	var processingError error

	log := logExecutionTime(func() {
		if m.GuildID == config.GetConfig().EveHomeGuild {
			checkAndForwardThreadMessage(s, m)
		}

		if !m.MentionEveryone && bot_utils.UserContains(m.Mentions, s.State.User.ID) {
			// Send typing
			_ = s.ChannelTyping(m.ChannelID)

			// Capture the request
			aiRequest = m.Content

			// Process the AI request
			response, err := intelligence.SendMessageToAIChat(s, m.GuildID, m.Content, m.Author.ID)
			if err != nil {
				processingError = err
				logger.ErrorLogger.Println("Erreur lors de l'envoi du message à l'IA:", err)
				return
			}

			// Capture the response
			if response != "" {
				aiResponse = response
				_, err = s.ChannelMessageSend(m.ChannelID, response)
				if err != nil {
					processingError = err
					logger.ErrorLogger.Println("Erreur lors de l'envoi du message au salon:", err)
				}
			}
		}

		bot_utils.DetectFeur(s, m)
	})

	// Create a unified log message with request and response
	logMessage := fmt.Sprintf("Message de serveur traité de %s dans %s", safeGetUsername(m.Author), m.ChannelID)

	// Add AI request/response details if applicable
	if aiRequest != "" {
		truncatedRequest := truncateString(aiRequest, 100)
		logMessage += fmt.Sprintf(" | Requête IA: %s", truncatedRequest)

		if aiResponse != "" {
			truncatedResponse := truncateString(aiResponse, 100)
			logMessage += fmt.Sprintf(" | Réponse IA: %s", truncatedResponse)
		}

		if processingError != nil {
			logMessage += fmt.Sprintf(" | Erreur: %s", processingError)
		}
	}

	// Log everything in one message
	log(logMessage)
}

// truncateString truncates a string to maxLen and adds "..." if it was truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	log := logExecutionTime(func() {
		s.UpdateStatusComplex(discordgo.UpdateStatusData{
			IdleSince: nil,
			Activities: []*discordgo.Activity{
				{
					Name: "démarrage...",
					Type: discordgo.ActivityTypeCompeting,
				},
			},
			Status: "dnd",
			AFK:    false,
		})

		registerCommands(s)
		InitMpThreadManager()

		go startStatusChange(s)

		// Register stream handlers before starting the stream check
		RegisterStreamHandlers()
		go twitch.StartAutomaticStreamCheck()
	})
	log("Le bot est prêt !")
}

// formatCommandInfo creates a descriptive string for a command
func formatCommandInfo(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	cmdData := i.ApplicationCommandData()
	cmdName := cmdData.Name
	userID, username := getUserInfo(i)

	if cmdData.CommandType == discordgo.UserApplicationCommand {
		return fmt.Sprintf("commande utilisateur %s par %s sur %s", cmdName, username, cmdData.TargetID)
	} else if cmdData.CommandType == discordgo.MessageApplicationCommand {
		return fmt.Sprintf("commande message %s par %s sur %s", cmdName, username, cmdData.TargetID)
	} else {
		args := cmdData.Options
		if len(args) > 0 {
			var formattedArgs []string
			for _, arg := range args {
				formattedArgs = append(formattedArgs, formatCommandArg(s, i.GuildID, arg))
			}

			if i.GuildID != "" {
				return fmt.Sprintf("commande %s par %s (%s) dans %s avec arguments: %s",
					cmdName, username, userID, i.ChannelID, strings.Join(formattedArgs, ", "))
			} else {
				return fmt.Sprintf("commande %s par %s (%s) avec arguments: %s",
					cmdName, username, userID, strings.Join(formattedArgs, ", "))
			}
		} else {
			if i.GuildID != "" {
				return fmt.Sprintf("commande %s par %s (%s) dans %s", cmdName, username, userID, i.ChannelID)
			} else {
				return fmt.Sprintf("commande %s par %s (%s)", cmdName, username, userID)
			}
		}
	}
}

// formatComponentInfo creates a descriptive string for a component interaction
func formatComponentInfo(i *discordgo.InteractionCreate) string {
	customID := i.MessageComponentData().CustomID
	userID, username := getUserInfo(i)
	return fmt.Sprintf("composant %s par %s (%s)", customID, username, userID)
}

// formatModalInfo creates a descriptive string for a modal submission
func formatModalInfo(i *discordgo.InteractionCreate) string {
	modalID := i.ModalSubmitData().CustomID
	userID, username := getUserInfo(i)
	return fmt.Sprintf("modal %s par %s (%s)", modalID, username, userID)
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, username := getUserInfo(i)

	if bot_utils.IsMaintenanceMode() && userID != config.GetConfig().OwnerId {
		logger.WarningLogger.Println("Mode maintenance activé. Interaction ignorée de", username)
		bot_utils.MaintenanceModeEmbed(s, i)
		return
	}

	log := logExecutionTime(func() {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			handleCommandInteraction(s, i)
		case discordgo.InteractionMessageComponent:
			handleComponentInteraction(s, i)
		case discordgo.InteractionModalSubmit:
			handleModalInteraction(s, i)
		default:
			logger.WarningLogger.Printf("Type d'interaction non géré: %v\n", i.Type)
		}
	})

	// Single log with all interaction details
	var interactionInfo string
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		interactionInfo = formatCommandInfo(s, i)
	case discordgo.InteractionMessageComponent:
		interactionInfo = formatComponentInfo(i)
	case discordgo.InteractionModalSubmit:
		interactionInfo = formatModalInfo(i)
	default:
		interactionInfo = fmt.Sprintf("Interaction inconnue par %s", username)
	}

	log("Interaction traitée: %s", interactionInfo)
}

// handleCommandInteraction processes application command interactions
func handleCommandInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name
	if handler, ok := commandHandlers[cmdName]; ok {
		err := handler(s, i)
		if err != nil {
			handleInteractionError(s, i, err.Error())
		}
	} else {
		handleInteractionError(s, i, "commande inconnue: "+cmdName)
	}
}

// handleComponentInteraction processes message component interactions
func handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Split customID at --
	customId := strings.Split(i.MessageComponentData().CustomID, "--")[0]
	if handler, ok := componentHandlers[customId]; ok {
		err := handler(s, i)
		if err != nil {
			handleInteractionError(s, i, err.Error())
		}
	} else {
		handleInteractionError(s, i, "composant inconnu: "+i.MessageComponentData().CustomID)
	}
}

// handleModalInteraction processes modal submission interactions
func handleModalInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modalID := i.ModalSubmitData().CustomID
	if handler, ok := modalHandlers[modalID]; ok {
		err := handler(s, i)
		if err != nil {
			handleInteractionError(s, i, err.Error())
		}
	} else {
		handleInteractionError(s, i, "modal inconnu: "+modalID)
	}
}
