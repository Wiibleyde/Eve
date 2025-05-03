package commandHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/prisma/db"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	QuizMaxTime = 30 // seconds
)

type Quiz struct {
	Question       string
	Answer         string
	BadAnswers     [3]string
	ShuffleAnswers [4]string
	Category       string
	Difficulty     string
	CreatedAt      time.Time
	RightUsers     []string
	WrongUsers     []string
}

func QuizHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "launch":
		client, ctx := data.GetDBClient()
		var quizCount []struct {
			Count db.RawInt `json:"count"`
		}
		err := client.Prisma.QueryRaw("SELECT COUNT(*) FROM QuizQuestions").Exec(ctx, &quizCount)
		if err != nil {
			return err
		}
		if quizCount[0].Count == 0 {
			return errors.New("no quiz in the database")
		}
		// Skip random number of quiz to get a random quiz
		randomNumber := rand.Intn(int(quizCount[0].Count))
		randomQuiz, err := client.QuizQuestions.FindMany().Take(1).Skip(randomNumber).Exec(ctx)
		if err != nil {
			return err
		}

		if len(randomQuiz) == 0 {
			return errors.New("no quiz found")
		}
		quiz := randomQuiz[0]

		// Shuffle the answers
		shuffleAnswers := [4]string{quiz.Answer, quiz.BadAnswer1, quiz.BadAnswer2, quiz.BadAnswer3}
		rand.Shuffle(len(shuffleAnswers), func(i, j int) {
			shuffleAnswers[i], shuffleAnswers[j] = shuffleAnswers[j], shuffleAnswers[i]
		})

		quizObj := Quiz{
			Question:       quiz.Question,
			Answer:         quiz.Answer,
			BadAnswers:     [3]string{quiz.BadAnswer1, quiz.BadAnswer2, quiz.BadAnswer3},
			ShuffleAnswers: shuffleAnswers,
			Category:       quiz.Category,
			Difficulty:     quiz.Difficulty,
			CreatedAt:      quiz.CreatedAt,
		}

		// Create a message embed
		var embed *discordgo.MessageEmbed
		if authorData, ok := randomQuiz[0].Author(); ok && authorData != nil {
			if authorData, ok := randomQuiz[0].Author(); ok && authorData != nil {
				embed = quizObj.GenerateEmbed(s, authorData.UserID)
			} else {
				embed = quizObj.GenerateEmbed(s, "")
			}
		} else {
			embed = quizObj.GenerateEmbed(s, "")
		}

		channel := i.ChannelID
		_, err = s.ChannelMessageSendEmbed(channel, embed)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
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

func (q *Quiz) GenerateEmbed(s *discordgo.Session, author string) *discordgo.MessageEmbed {
	embed := bot_utils.BasicEmbedBuilder(s)
	embed.Title = "Question de quiz"
	embed.Description = "```\n" + q.Question + "\n```\n1) `" + q.ShuffleAnswers[0] + "`\n2) `" + q.ShuffleAnswers[1] + "`\n3) `" + q.ShuffleAnswers[2] + "`\n4) `" + q.ShuffleAnswers[3] + "`"
	embed.Color = 0x4EBAF6
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "Catégorie / Difficulté",
			Value:  q.Category + " / " + q.Difficulty,
			Inline: true,
		},
		{
			Name:   "Invalide",
			Value:  "<t:" + strconv.FormatInt(q.CreatedAt.Unix(), 10) + ":R>",
			Inline: true,
		},
	}

	if author != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Créateur",
			Value:  "<@" + author + ">",
			Inline: true,
		})
	}

	return embed
}
