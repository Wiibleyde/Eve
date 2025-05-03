package buttonHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/game"
	"main/pkg/logger"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func QuizAnswerButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get the message ID from the interaction
	messageID := i.Message.ID

	currentQuiz := game.GetQuiz(messageID)
	if currentQuiz == nil {
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Et non !"
		embed.Description = "Ce quiz n'existe pas ou a déjà été terminé."
		embed.Color = 0xDBC835
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
	}

	// Check if the user has already answered (right or wrong)
	for _, userID := range currentQuiz.RightUsers {
		if userID == i.Member.User.ID {
			embed := bot_utils.BasicEmbedBuilder(s)
			embed.Title = "Et non !"
			embed.Description = "Vous avez déjà répondu à ce quiz."
			embed.Color = 0xDBC835
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to interaction:", err)
				return err
			}
			return nil
		}
	}
	for _, userID := range currentQuiz.WrongUsers {
		if userID == i.Member.User.ID {
			embed := bot_utils.BasicEmbedBuilder(s)
			embed.Title = "Et non !"
			embed.Description = "Vous avez déjà répondu à ce quiz."
			embed.Color = 0xDBC835
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to interaction:", err)
				return err
			}
			return nil
		}
	}

	// Check if the quiz is still active
	if !currentQuiz.IsActive() {
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Et non !"
		embed.Description = "Ce quiz n'existe pas ou a déjà été terminé."
		embed.Color = 0xDBC835
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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

	answerSelectedId := strings.Split(i.MessageComponentData().CustomID, "--")[1]
	answerSelectedIdInt, err := strconv.Atoi(answerSelectedId)
	if err != nil {
		logger.ErrorLogger.Println("Error converting answer selected ID to int:", err)
		return err
	}
	answerSelected := currentQuiz.ShuffleAnswers[answerSelectedIdInt]
	if answerSelected == currentQuiz.Answer {
		currentQuiz.RightUsers = append(currentQuiz.RightUsers, i.Member.User.ID)
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Bravo !"
		embed.Description = "Vous avez trouvé la bonne réponse !"
		embed.Color = 0x4CAF50
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}
	} else {
		currentQuiz.WrongUsers = append(currentQuiz.WrongUsers, i.Member.User.ID)
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Et non !"
		embed.Description = "La bonne réponse était : `" + currentQuiz.Answer + "`"
		embed.Color = 0xDBC835
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}
	}

	embed := currentQuiz.GenerateEmbed(s)
	components := currentQuiz.GenerateComponents()
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    i.ChannelID,
		Embed:      embed,
		Components: &components,
	})

	if err != nil {
		logger.ErrorLogger.Println("Error editing message:", err)
		return err
	}

	return nil
}
