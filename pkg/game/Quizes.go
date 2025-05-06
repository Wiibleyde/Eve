package game

import (
	"main/pkg/bot_utils"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	QuizMaxTime = 8 // in hours
)

type Quiz struct {
	Question       string
	Answer         string
	BadAnswers     [3]string
	ShuffleAnswers [4]string
	Category       string
	Difficulty     string
	Author         string
	CreatedAt      time.Time
	RightUsers     []string
	WrongUsers     []string
}

type Quizes struct {
	Quizes map[string]*Quiz
}

var quizManager Quizes = Quizes{
	Quizes: make(map[string]*Quiz),
}

func GetQuiz(messageID string) *Quiz {
	if quiz, ok := quizManager.Quizes[messageID]; ok {
		return quiz
	}
	return nil
}

func AddQuiz(messageID string, quiz *Quiz) {
	if _, ok := quizManager.Quizes[messageID]; !ok {
		quizManager.Quizes[messageID] = quiz
	}
}

func RemoveQuiz(messageID string) {
	delete(quizManager.Quizes, messageID)
}

func (q *Quiz) IsActive() bool {
	return q.CreatedAt.Add(time.Duration(QuizMaxTime) * time.Hour).After(time.Now())
}

func (q *Quiz) GenerateEmbed(s *discordgo.Session) *discordgo.MessageEmbed {
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
			Value:  "<t:" + strconv.FormatInt(q.CreatedAt.Unix()+(QuizMaxTime*3600), 10) + ":R>",
			Inline: true,
		},
	}

	// Add the author of the quiz
	if q.Author != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Auteur",
			Value:  "<@" + q.Author + ">",
			Inline: true,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Auteur",
			Value:  "Inconnu",
			Inline: true,
		})
	}

	if q.RightUsers != nil {
		value := ""
		for _, userID := range q.RightUsers {
			value += "<@" + userID + ">\n"
		}
		// Append all users who answered correctly
		if len(q.RightUsers) > 0 {
			value := ""
			for _, userID := range q.RightUsers {
				value += "<@" + userID + ">\n"
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Bonne réponse",
				Value:  value,
				Inline: true,
			})
		}
	}

	if q.WrongUsers != nil {
		value := ""
		for _, userID := range q.WrongUsers {
			value += "<@" + userID + ">\n"
		}
		// Append all users who answered incorrectly
		if len(q.WrongUsers) > 0 {
			value := ""
			for _, userID := range q.WrongUsers {
				value += "<@" + userID + ">\n"
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Mauvaise réponse",
				Value:  value,
				Inline: true,
			})
		}
	}

	return embed
}

func (q *Quiz) GenerateComponents() []discordgo.MessageComponent {
	buttons := make([]discordgo.MessageComponent, 0, 4)
	for i := 0; i < 4; i++ {
		buttons = append(buttons, &discordgo.Button{
			Label:    "Réponse " + strconv.Itoa(i+1),
			Style:    discordgo.PrimaryButton,
			CustomID: "quizAnswerButton--" + strconv.Itoa(i),
		})
	}

	actionRow := &discordgo.ActionsRow{
		Components: buttons,
	}

	return []discordgo.MessageComponent{
		actionRow,
	}
}
