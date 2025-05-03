package commandHandler

import (
	"errors"
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/game"
	"main/pkg/logger"
	"main/prisma/db"
	"math/rand"
	"sort"
	"strconv"
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
		strQuestion := subCommand.Options[0].StringValue()
		strAnswer := subCommand.Options[1].StringValue()
		strBadAnswer1 := subCommand.Options[2].StringValue()
		strBadAnswer2 := subCommand.Options[3].StringValue()
		strBadAnswer3 := subCommand.Options[4].StringValue()
		strCategory := subCommand.Options[5].StringValue()
		strDifficulty := subCommand.Options[6].StringValue()
		strAuthorID := i.Member.User.ID
		strGuildID := i.GuildID

		client, ctx := data.GetDBClient()

		_, err := client.QuizQuestions.CreateOne(
			db.QuizQuestions.Question.Set(strQuestion),
			db.QuizQuestions.Answer.Set(strAnswer),
			db.QuizQuestions.BadAnswer1.Set(strBadAnswer1),
			db.QuizQuestions.BadAnswer2.Set(strBadAnswer2),
			db.QuizQuestions.BadAnswer3.Set(strBadAnswer3),
			db.QuizQuestions.GuildID.Set(strGuildID),
			db.QuizQuestions.Category.Set(strCategory),
			db.QuizQuestions.Difficulty.Set(strDifficulty),
			db.QuizQuestions.Author.Link(
				db.GlobalUserData.UserID.Equals(strAuthorID),
			),
		).Exec(ctx)

		if err != nil {
			logger.ErrorLogger.Println("Error creating quiz question:", err)
			return err
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{bot_utils.SuccessEmbed(s, "Quiz créé !")},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.ErrorLogger.Println("Error responding to interaction:", err)
			return err
		}

		return nil
	case "leaderboard":
		// Handle the leaderboard of the quiz
		client, ctx := data.GetDBClient()
		leaderboard, err := client.GlobalUserData.FindMany().Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error fetching leaderboard data:", err)
			return err
		}

		// Define round function
		round := func(x float64) int {
			return int(x + 0.5)
		}
		leaderboardType := subCommand.Options[0].StringValue()
		switch leaderboardType {
		case "ratio":
			// Handle the leaderboard by ratio
			var ratioLeaderboard []*db.GlobalUserDataModel
			for _, userData := range leaderboard {
				if userData.QuizGoodAnswers+userData.QuizBadAnswers > 0 {
					ratioLeaderboard = append(ratioLeaderboard, &userData)
				}
			}
			// Sort the leaderboard by ratio
			sort.Slice(ratioLeaderboard, func(i, j int) bool {
				ratioI := float64(ratioLeaderboard[i].QuizGoodAnswers) / float64(ratioLeaderboard[i].QuizGoodAnswers+ratioLeaderboard[i].QuizBadAnswers)
				ratioJ := float64(ratioLeaderboard[j].QuizGoodAnswers) / float64(ratioLeaderboard[j].QuizGoodAnswers+ratioLeaderboard[j].QuizBadAnswers)
				return ratioI > ratioJ
			})
			// Create the embed
			embed := bot_utils.BasicEmbedBuilder(s)
			embed.Title = "Leaderboard des ratios"
			embed.Description = "Voici le leaderboard des ratios :"
			embed.Fields = []*discordgo.MessageEmbedField{}
			for i, userData := range ratioLeaderboard {
				if i >= 10 {
					break
				}
				ratio := float64(userData.QuizGoodAnswers) / float64(userData.QuizGoodAnswers+userData.QuizBadAnswers) * 100
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:  strconv.Itoa(i+1) + ". Ratio : " + strconv.Itoa(round(ratio)) + "%",
					Value: "<@" + userData.UserID + "> " + strconv.Itoa(userData.QuizGoodAnswers) + "/" + strconv.Itoa(userData.QuizBadAnswers),
				})
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
			return nil
		case "goodAnswers":
			// Handle the leaderboard by good answers
			var goodAnswerLeaderboard []*db.GlobalUserDataModel
			for _, userData := range leaderboard {
				if userData.QuizGoodAnswers > 0 {
					goodAnswerLeaderboard = append(goodAnswerLeaderboard, &userData)
				}
			}
			// Sort the leaderboard by good answers
			sort.Slice(goodAnswerLeaderboard, func(i, j int) bool {
				return goodAnswerLeaderboard[i].QuizGoodAnswers > goodAnswerLeaderboard[j].QuizGoodAnswers
			})
			// Create the embed
			embed := bot_utils.BasicEmbedBuilder(s)
			embed.Title = "Leaderboard des bonnes réponses"
			embed.Description = "Voici le leaderboard des bonnes réponses :"
			embed.Fields = []*discordgo.MessageEmbedField{}
			for i, userData := range goodAnswerLeaderboard {
				if i >= 10 {
					break
				}
				ratio := float64(userData.QuizGoodAnswers) / float64(userData.QuizGoodAnswers+userData.QuizBadAnswers) * 100

				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:  strconv.Itoa(i+1) + ". Ratio : " + strconv.Itoa(round(ratio)) + "%",
					Value: "<@" + userData.UserID + "> " + strconv.Itoa(userData.QuizGoodAnswers) + " bonnes réponses",
				})
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

			return nil
		case "badAnswers":
			// Handle the leaderboard by bad answers
			var badAnswerLeaderboard []*db.GlobalUserDataModel
			for _, userData := range leaderboard {
				if userData.QuizBadAnswers > 0 {
					badAnswerLeaderboard = append(badAnswerLeaderboard, &userData)
				}
			}
			// Sort the leaderboard by bad answers
			sort.Slice(badAnswerLeaderboard, func(i, j int) bool {
				return badAnswerLeaderboard[i].QuizBadAnswers > badAnswerLeaderboard[j].QuizBadAnswers
			})
			// Create the embed
			embed := bot_utils.BasicEmbedBuilder(s)
			embed.Title = "Leaderboard des mauvaises réponses"
			embed.Description = "Voici le leaderboard des mauvaises réponses :"
			embed.Fields = []*discordgo.MessageEmbedField{}
			for i, userData := range badAnswerLeaderboard {
				if i >= 10 {
					break
				}
				ratio := float64(userData.QuizBadAnswers) / float64(userData.QuizGoodAnswers+userData.QuizBadAnswers) * 100

				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:  strconv.Itoa(i+1) + ". Ratio : " + strconv.Itoa(round(ratio)) + "%",
					Value: "<@" + userData.UserID + "> " + strconv.Itoa(userData.QuizBadAnswers) + " mauvaises réponses",
				})
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
			return nil
		}
	case "me":
		client, ctx := data.GetDBClient()
		userData, err := client.GlobalUserData.FindUnique(
			db.GlobalUserData.UserID.Equals(i.Member.User.ID),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error fetching user data:", err)
			return err
		}
		if userData == nil {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{bot_utils.WarningEmbed(s, "Aucun quiz trouvé !")},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.ErrorLogger.Println("Error responding to interaction:", err)
				return err
			}
		}

		totalQuizPlayed := userData.QuizBadAnswers + userData.QuizGoodAnswers
		strTotalQuizPlayed := strconv.Itoa(totalQuizPlayed)

		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Mes stats de quiz"
		embed.Description = "Voici vos statistiques de quiz :"
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Nombre de réponses données",
				Value:  strTotalQuizPlayed,
				Inline: true,
			},
			{
				Name:   "Nombre de réponses correctes",
				Value:  strconv.Itoa(userData.QuizGoodAnswers),
				Inline: true,
			},
			{
				Name:   "Nombre de réponses incorrectes",
				Value:  strconv.Itoa(userData.QuizBadAnswers),
				Inline: true,
			},
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

		return nil
	}
	return nil
}
