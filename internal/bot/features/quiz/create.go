package quiz

import (
	"context"
	"strings"
	"unicode/utf8"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
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
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msg))
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
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Cette question existe déjà."))
			return
		}
		logger.Error("Error creating quiz question", "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement de la question."))
		return
	}

	inline := true
	embed := embeds.SuccessEmbed("Question ajoutée au quiz !")
	embed.Author = &discord.EmbedAuthor{Name: authorName}
	embed.Fields = []discord.EmbedField{
		{Name: "Question", Value: truncate(question, 1024)},
		{Name: "Bonne réponse", Value: truncate(good, 1024)},
		{Name: "Catégorie", Value: category, Inline: &inline},
		{Name: "Difficulté", Value: difficulty, Inline: &inline},
	}
	helpers.RespondEphemeralEmbed(e, embed)
}

// validateQuestion enforces what the database columns and the Discord button
// labels require, and returns the French message to show when it fails.
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

	// Duplicate answers would make two buttons indistinguishable, and one of
	// them would be scored wrong for the same text.
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
