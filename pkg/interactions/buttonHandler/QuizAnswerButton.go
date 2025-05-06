package buttonHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/game"
	"main/pkg/logger"
	"main/prisma/db"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func QuizAnswerButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get the message ID from the interaction
	messageID := i.Message.ID

	message, err := s.ChannelMessage(i.ChannelID, messageID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting message:", err)
		return err
	}

	questionFoundInEmbed := strings.ReplaceAll(strings.Split(message.Embeds[0].Description, "```")[1], "\n", "")

	currentQuiz := game.GetQuiz(messageID)
	if currentQuiz == nil {
		client, ctx := data.GetDBClient()
		databaseQuiz, err := client.QuizQuestions.FindUnique(
			db.QuizQuestions.Question.Equals(questionFoundInEmbed),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error finding quiz question:", err)
			return err
		}

		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Et non !"
		embed.Description = "Ce quiz n'existe pas ou a déjà été terminé. (La bonne réponse était : `" + databaseQuiz.Answer + "`)"
		embed.Color = 0xDBC835
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

	// Get the user ID, safely handling nil cases
	var userID string
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	} else {
		// Handle case where we can't identify the user
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Erreur"
		embed.Description = "Impossible d'identifier l'utilisateur."
		embed.Color = 0xFF0000
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

	// Check if the user has already answered (right or wrong)
	for _, id := range currentQuiz.RightUsers {
		if id == userID {
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
	for _, id := range currentQuiz.WrongUsers {
		if id == userID {
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

	client, ctx := data.GetDBClient()

	// Check if the quiz is still active
	if !currentQuiz.IsActive() {
		client, ctx := data.GetDBClient()
		databaseQuiz, err := client.QuizQuestions.FindUnique(
			db.QuizQuestions.Question.Equals(questionFoundInEmbed),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error finding quiz question:", err)
			return err
		}

		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Et non !"
		embed.Description = "Ce quiz n'existe pas ou a déjà été terminé. (La bonne réponse était : `" + databaseQuiz.Answer + "`)"
		embed.Color = 0xDBC835
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

	answerSelectedId := strings.Split(i.MessageComponentData().CustomID, "--")[1]
	answerSelectedIdInt, err := strconv.Atoi(answerSelectedId)
	if err != nil {
		logger.ErrorLogger.Println("Error converting answer selected ID to int:", err)
		return err
	}
	answerSelected := currentQuiz.ShuffleAnswers[answerSelectedIdInt]
	if answerSelected == currentQuiz.Answer {
		currentQuiz.RightUsers = append(currentQuiz.RightUsers, userID)

		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Bravo !"
		embed.Description = "Vous avez trouvé la bonne réponse !"
		embed.Color = 0x4CAF50

		_, err = client.GlobalUserData.UpsertOne(
			db.GlobalUserData.UserID.Equals(userID),
		).Create(
			db.GlobalUserData.UserID.Set(userID),
			db.GlobalUserData.QuizGoodAnswers.Set(1),
		).Update(
			db.GlobalUserData.QuizGoodAnswers.Increment(1),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error updating user data:", err)
			return err
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
		currentQuiz.WrongUsers = append(currentQuiz.WrongUsers, userID)

		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Et non !"
		embed.Description = "La bonne réponse était : `" + currentQuiz.Answer + "`"
		embed.Color = 0xDBC835

		_, err = client.GlobalUserData.UpsertOne(
			db.GlobalUserData.UserID.Equals(userID),
		).Create(
			db.GlobalUserData.UserID.Set(userID),
			db.GlobalUserData.QuizBadAnswers.Set(1),
		).Update(
			db.GlobalUserData.QuizBadAnswers.Increment(1),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error updating user data:", err)
			return err
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
