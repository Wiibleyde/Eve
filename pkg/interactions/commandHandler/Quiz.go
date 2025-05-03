package commandHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/game"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

func QuizHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "launch":
		client, ctx := data.GetDBClient()

		quizes, err := client.QuizQuestions.FindMany().Exec(ctx)
		if err != nil {
			return err
		}
		if len(quizes) == 0 {
			return errors.New("aucun quiz trouvé")
		}

		// Select a random quiz
		quiz := quizes[rand.Intn(len(quizes))]

		// Shuffle the answers
		shuffleAnswers := [4]string{quiz.Answer, quiz.BadAnswer1, quiz.BadAnswer2, quiz.BadAnswer3}
		rand.Shuffle(len(shuffleAnswers), func(i, j int) {
			shuffleAnswers[i], shuffleAnswers[j] = shuffleAnswers[j], shuffleAnswers[i]
		})

		quizObj := game.Quiz{
			Question:   quiz.Question,
			Answer:     quiz.Answer,
			BadAnswers: [3]string{quiz.BadAnswer1, quiz.BadAnswer2, quiz.BadAnswer3},
			Author: func() string {
				if authorData, ok := quiz.Author(); ok && authorData != nil {
					return authorData.UserID
				}
				return ""
			}(),
			ShuffleAnswers: shuffleAnswers,
			Category:       quiz.Category,
			Difficulty:     quiz.Difficulty,
			CreatedAt:      time.Now(),
		}

		embed := quizObj.GenerateEmbed(s)

		components := quizObj.GenerateComponents()

		channel := i.ChannelID
		newMessage, err := s.ChannelMessageSendComplex(channel, &discordgo.MessageSend{
			Embed:      embed,
			Components: components,
		})
		if err != nil {
			return err
		}

		game.AddQuiz(newMessage.ID, &quizObj)

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{bot_utils.SuccessEmbed(s, "Quiz lancé !")},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}

		return nil
	case "create":
		// Handle the creation of a quiz
		return nil
	case "leaderboard":
		// Handle the leaderboard of the quiz
	case "me":
		// Handle the quiz for the user
		return nil
	}
	return nil
}
