package quiz

import (
	"context"
	"strconv"
	"time"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/activequiz"
	"Eve/internal/database/ent/quizquestion"
	"Eve/internal/database/ent/quizstat"
	"Eve/internal/database/ent/quizuseranswer"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/events"
	"github.com/google/uuid"
)

// HandleAnswer handles a click on "quiz:answer:<idx>:<questionID>", where idx
// is the displayed button position. The question id segment is absent on
// buttons posted before it was added, hence the two accepted shapes.
func HandleAnswer(e *events.ComponentInteractionCreate, args []string) {
	if len(args) == 0 || len(args) > 2 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Réponse invalide."))
		return
	}
	idx, err := strconv.Atoi(args[0])
	if err != nil || idx < 0 || idx >= answerCount {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Réponse invalide."))
		return
	}
	var questionID string
	if len(args) == 2 {
		questionID = args[1]
	}

	ctx := context.Background()
	db := database.Default.Ent()

	// The running quiz is resolved from the message it was posted on, never by
	// parsing the embed text.
	active, err := db.ActiveQuiz.Query().
		Where(activequiz.MessageID(e.Message.ID.String())).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error("Error loading active quiz", "error", err)
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la récupération du quiz."))
			return
		}
		helpers.RespondEphemeralEmbed(e, expiredEmbed(purgedQuizGoodAnswer(ctx, questionID, e)))
		return
	}

	question, err := db.QuizQuestion.Get(ctx, active.QuestionID)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error("Error loading quiz question", "error", err, "question", active.QuestionID)
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la récupération de la question."))
			return
		}
		helpers.RespondEphemeralEmbed(e, expiredEmbed(""))
		return
	}

	if !time.Now().Before(active.ExpiresAt) {
		helpers.RespondEphemeralEmbed(e, expiredEmbed(question.GoodAnswer))
		return
	}

	perm, err := decodePermutation(active.Shuffle)
	if err != nil {
		logger.Error("Corrupted quiz shuffle", "error", err, "quiz", active.ID, "shuffle", active.Shuffle)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Ce quiz est corrompu, impossible d'enregistrer votre réponse."))
		return
	}

	userID := e.User().ID.String()
	answered, err := db.QuizUserAnswer.Query().
		Where(quizuseranswer.ActiveQuizID(active.ID), quizuseranswer.UserID(userID)).
		Exist(ctx)
	if err != nil {
		logger.Error("Error checking previous quiz answer", "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement de votre réponse."))
		return
	}
	if answered {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Vous avez déjà répondu à ce quiz."))
		return
	}

	correct := perm[idx] == goodAnswerIndex

	// The unique index on (active_quiz_id, user_id) is what actually enforces
	// one answer per user: the check above only saves a round trip.
	err = db.QuizUserAnswer.Create().
		SetID(uuid.New().String()).
		SetActiveQuizID(active.ID).
		SetUserID(userID).
		SetCorrect(correct).
		Exec(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Vous avez déjà répondu à ce quiz."))
			return
		}
		logger.Error("Error saving quiz answer", "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement de votre réponse."))
		return
	}

	// The answer is already recorded: a stats failure must not hide the result.
	if err := recordStat(ctx, userID, correct); err != nil {
		logger.Error("Error updating quiz stats", "error", err, "user", userID)
	}

	helpers.RespondEphemeralEmbed(e, resultEmbed(correct, question.GoodAnswer))
}

// purgedQuizGoodAnswer resolves the good answer of a quiz whose active row is
// gone. Buttons carrying the question id resolve it exactly; older ones fall
// back to a lookup on the question text read from the embed. Returns "" when
// neither works.
func purgedQuizGoodAnswer(ctx context.Context, questionID string, e *events.ComponentInteractionCreate) string {
	if questionID != "" {
		question, err := database.Default.Ent().QuizQuestion.Get(ctx, questionID)
		if err != nil {
			if !ent.IsNotFound(err) {
				logger.Error("Error loading quiz question", "error", err, "question", questionID)
			}
			return ""
		}
		return question.GoodAnswer
	}

	text, ok := questionTextFromMessage(e.Message)
	if !ok {
		return ""
	}
	question, err := database.Default.Ent().QuizQuestion.Query().
		Where(quizquestion.Question(text)).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error("Error looking up quiz question by text", "error", err)
		}
		return ""
	}
	return question.GoodAnswer
}

// recordStat increments the user counters, creating the row on first answer.
//
// ent is generated without the upsert feature here, so this is the same
// update-then-create dance the other features use, with one retry to settle a
// race between two first answers.
func recordStat(ctx context.Context, userID string, correct bool) error {
	db := database.Default.Ent()

	n, err := increment(ctx, userID, correct)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	create := db.QuizStat.Create().SetID(uuid.New().String()).SetUserID(userID)
	if correct {
		create.SetGoodAnswers(1)
	} else {
		create.SetBadAnswers(1)
	}
	if err := create.Exec(ctx); err != nil {
		if !ent.IsConstraintError(err) {
			return err
		}
		// Someone else created the row in between: increment it instead.
		_, err = increment(ctx, userID, correct)
		return err
	}
	return nil
}

func increment(ctx context.Context, userID string, correct bool) (int, error) {
	update := database.Default.Ent().QuizStat.Update().Where(quizstat.UserID(userID))
	if correct {
		update.AddGoodAnswers(1)
	} else {
		update.AddBadAnswers(1)
	}
	return update.Save(ctx)
}
