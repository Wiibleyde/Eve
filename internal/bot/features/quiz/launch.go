package quiz

import (
	"context"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/quizquestion"
	"Eve/internal/logger"

	"entgo.io/ent/dialect/sql"
	"github.com/disgoorg/disgo/events"
	"github.com/google/uuid"
)

const Duration = 8 * time.Hour

func handleLaunch(e *events.ApplicationCommandInteractionCreate) {
	guildID, ok := requireGuild(e)
	if !ok {
		return
	}

	ctx := context.Background()
	db := database.Default.Ent()

	question, err := pickQuestion(ctx)
	if err != nil {
		logger.Error("Error picking a quiz question", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de la récupération d'une question."))
		return
	}
	if question == nil {
		helpers.RespondEphemeralCard(e, ui.Error("Aucun quiz n'a été créé pour le moment. Utilisez `/quiz create` pour en ajouter un."))
		return
	}

	if err := db.QuizQuestion.UpdateOneID(question.ID).SetLastUsedAt(time.Now()).Exec(ctx); err != nil {
		logger.Error("Error updating quiz question last_used_at", "error", err, "question", question.ID)
	}

	perm := shufflePermutation()
	answers := applyPermutation(storedAnswers(question), perm)
	expiresAt := time.Now().Add(Duration)

	msg, err := e.Client().Rest.CreateMessage(e.Channel().ID(),
		questionCard(question, answers, expiresAt, nil, nil).MessageCreate())
	if err != nil {
		logger.Error("Error sending quiz message", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Impossible d'envoyer le quiz dans ce salon."))
		return
	}

	err = db.ActiveQuiz.Create().
		SetID(uuid.New().String()).
		SetQuestionID(question.ID).
		SetMessageID(msg.ID.String()).
		SetChannelID(msg.ChannelID.String()).
		SetGuildID(guildID).
		SetShuffle(encodePermutation(perm)).
		SetExpiresAt(expiresAt).
		Exec(ctx)
	if err != nil {
		logger.Error("Error saving active quiz", "error", err)
		if delErr := e.Client().Rest.DeleteMessage(msg.ChannelID, msg.ID); delErr != nil {
			logger.Error("Error deleting orphaned quiz message", "error", delErr)
		}
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors du lancement du quiz."))
		return
	}

	logger.Debug("Quiz launched", "question", question.ID, "message", msg.ID.String(), "guild", guildID)
	helpers.RespondEphemeralCard(e, ui.Success("Quiz lancé ! Les réponses sont acceptées pendant 8 heures."))
}

func pickQuestion(ctx context.Context) (*ent.QuizQuestion, error) {
	questions, err := database.Default.Ent().QuizQuestion.Query().
		Order(
			quizquestion.ByLastUsedAt(sql.OrderAsc(), sql.OrderNullsFirst()),
			func(s *sql.Selector) { s.OrderExpr(sql.Expr("random()")) },
		).
		Limit(1).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, nil
	}
	return questions[0], nil
}
