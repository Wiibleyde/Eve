package quiz

import (
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const CommandName = "quiz"

const ButtonAnswerPrefix = "quiz:answer"

const (
	subLaunch      = "launch"
	subCreate      = "create"
	subLeaderboard = "leaderboard"
)

const (
	choiceBestScores  = "best_scores"
	choiceBestRatios  = "best_ratios"
	choiceWorstScores = "worst_scores"
)

const (
	difficultyEasy   = "facile"
	difficultyNormal = "normal"
	difficultyHard   = "difficile"
)

const (
	maxQuestionInput = 1000
	maxAnswerInput   = 80
	maxCategoryInput = 64
)

var (
	questionMaxLength = maxQuestionInput
	answerMaxLength   = maxAnswerInput
	categoryMaxLength = maxCategoryInput
	guildOnly         = []discord.InteractionContextType{discord.InteractionContextTypeGuild}
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        CommandName,
		Description: "Quiz de la communauté",
		Contexts:    guildOnly,
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        subLaunch,
				Description: "Lancer une question au hasard dans ce salon",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        subCreate,
				Description: "Ajouter une question au quiz",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "question",
						Description: "La question posée",
						Required:    true,
						MaxLength:   &questionMaxLength,
					},
					discord.ApplicationCommandOptionString{
						Name:        "answer",
						Description: "La bonne réponse",
						Required:    true,
						MaxLength:   &answerMaxLength,
					},
					discord.ApplicationCommandOptionString{
						Name:        "bad1",
						Description: "Une mauvaise réponse",
						Required:    true,
						MaxLength:   &answerMaxLength,
					},
					discord.ApplicationCommandOptionString{
						Name:        "bad2",
						Description: "Une autre mauvaise réponse",
						Required:    true,
						MaxLength:   &answerMaxLength,
					},
					discord.ApplicationCommandOptionString{
						Name:        "bad3",
						Description: "Une dernière mauvaise réponse",
						Required:    true,
						MaxLength:   &answerMaxLength,
					},
					discord.ApplicationCommandOptionString{
						Name:        "category",
						Description: "La catégorie de la question",
						Required:    true,
						MaxLength:   &categoryMaxLength,
					},
					discord.ApplicationCommandOptionString{
						Name:        "difficulty",
						Description: "La difficulté de la question",
						Required:    true,
						Choices: []discord.ApplicationCommandOptionChoiceString{
							{Name: "Facile", Value: difficultyEasy},
							{Name: "Normal", Value: difficultyNormal},
							{Name: "Difficile", Value: difficultyHard},
						},
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        subLeaderboard,
				Description: "Afficher le classement du quiz",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "choice",
						Description: "Le classement à afficher",
						Required:    true,
						Choices: []discord.ApplicationCommandOptionChoiceString{
							{Name: "Meilleurs scores", Value: choiceBestScores},
							{Name: "Meilleurs ratios", Value: choiceBestRatios},
							{Name: "Scores les plus bas", Value: choiceWorstScores},
						},
					},
				},
			},
		},
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return
	}
	switch *data.SubCommandName {
	case subLaunch:
		handleLaunch(e)
	case subCreate:
		handleCreate(e)
	case subLeaderboard:
		handleLeaderboard(e)
	}
}

func requireGuild(e *events.ApplicationCommandInteractionCreate) (string, bool) {
	guildID := e.GuildID()
	if guildID == nil {
		helpers.RespondEphemeralCard(e, ui.Error("Cette commande doit être utilisée dans un serveur."))
		return "", false
	}
	return guildID.String(), true
}
