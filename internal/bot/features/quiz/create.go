package quiz

import (
	"context"
	"strings"
	"unicode/utf8"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/events"
	"github.com/google/uuid"
)

func handleCreate(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := requireGuild(e)
	if !ok {
		return
	}

	data := e.SlashCommandInteractionData()
	question := strings.TrimSpace(data.String("question"))
	good := strings.TrimSpace(data.String("answer"))
	bad := [3]string{
		strings.TrimSpace(data.String("bad1")),
		strings.TrimSpace(data.String("bad2")),
		strings.TrimSpace(data.String("bad3")),
	}
	category := strings.TrimSpace(data.String("category"))
	difficulty := strings.TrimSpace(data.String("difficulty"))

	if msg, ok := validateQuestion(question, good, bad, category, difficulty); !ok {
		helpers.RespondEphemeralCard(e, ui.Error(msg))
		return
	}

	ctx := context.Background()
	err := database.Default.Ent().QuizQuestion.Create().
		SetID(uuid.New().String()).
		SetQuestion(question).
		SetGoodAnswer(good).
		SetBadAnswer1(bad[0]).
		SetBadAnswer2(bad[1]).
		SetBadAnswer3(bad[2]).
		SetCategory(category).
		SetDifficulty(difficulty).
		SetAuthorID(e.User().ID.String()).
		SetGuildID(guildID).
		Exec(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			helpers.RespondEphemeralCard(e, ui.Error("Cette question existe déjà."))
			return
		}
		logger.Error("Error creating quiz question", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de l'enregistrement de la question."))
		return
	}

	card := ui.Success("Question ajoutée au quiz !").
		Divider().
		Fields(
			ui.Field{Name: "Question", Value: truncate(question, 1000)},
			ui.Field{Name: "Bonne réponse", Value: truncate(good, 1000)},
		).
		Fields(
			ui.Field{Name: "Catégorie", Value: category, Inline: true},
			ui.Field{Name: "Difficulté", Value: difficulty, Inline: true},
		)
	helpers.RespondEphemeralCard(e, card)
}

func validateQuestion(question, good string, bad [3]string, category, difficulty string) (string, bool) {
	if question == "" || good == "" || category == "" {
		return "La question, la bonne réponse et la catégorie ne peuvent pas être vides.", false
	}
	for _, b := range bad {
		if b == "" {
			return "Les mauvaises réponses ne peuvent pas être vides.", false
		}
	}
	if utf8.RuneCountInString(question) > maxQuestionInput {
		return "La question est trop longue.", false
	}
	if utf8.RuneCountInString(category) > maxCategoryInput {
		return "La catégorie est trop longue.", false
	}

	answers := append([]string{good}, bad[:]...)
	for _, a := range answers {
		if utf8.RuneCountInString(a) > maxAnswerInput {
			return "Chaque réponse doit faire 80 caractères maximum (limite des boutons Discord).", false
		}
	}

	seen := make(map[string]struct{}, len(answers))
	for _, a := range answers {
		key := strings.ToLower(a)
		if _, dup := seen[key]; dup {
			return "Les quatre réponses doivent être différentes.", false
		}
		seen[key] = struct{}{}
	}

	switch difficulty {
	case difficultyEasy, difficultyNormal, difficultyHard:
	default:
		return "Difficulté invalide.", false
	}

	return "", true
}
