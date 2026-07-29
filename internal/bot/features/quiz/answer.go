package quiz

import (
	"context"
	"strconv"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
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

func HandleAnswer(e *events.ComponentInteractionCreate, args []string) {
	if len(args) == 0 || len(args) > 2 {
		helpers.RespondEphemeralCard(e, ui.Error("Réponse invalide."))
		return
	}
	idx, err := strconv.Atoi(args[0])
	if err != nil || idx < 0 || idx >= answerCount {
		helpers.RespondEphemeralCard(e, ui.Error("Réponse invalide."))
		return
	}
	var questionID string
	if len(args) == 2 {
		questionID = args[1]
	}

	ctx := context.Background()
	db := database.Default.Ent()

	active, err := db.ActiveQuiz.Query().
		Where(activequiz.MessageID(e.Message.ID.String())).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error("Error loading active quiz", "error", err)
			helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de la récupération du quiz."))
			return
		}
		helpers.RespondEphemeralCard(e, expiredCard(purgedQuizGoodAnswer(ctx, questionID, e)))
		return
	}

	question, err := db.QuizQuestion.Get(ctx, active.QuestionID)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error("Error loading quiz question", "error", err, "question", active.QuestionID)
			helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de la récupération de la question."))
			return
		}
		helpers.RespondEphemeralCard(e, expiredCard(""))
		return
	}

	if !time.Now().Before(active.ExpiresAt) {
		helpers.RespondEphemeralCard(e, expiredCard(question.GoodAnswer))
		return
	}

	perm, err := decodePermutation(active.Shuffle)
	if err != nil {
		logger.Error("Corrupted quiz shuffle", "error", err, "quiz", active.ID, "shuffle", active.Shuffle)
		helpers.RespondEphemeralCard(e, ui.Error("Ce quiz est corrompu, impossible d'enregistrer votre réponse."))
		return
	}

	userID := e.User().ID.String()
	answered, err := db.QuizUserAnswer.Query().
		Where(quizuseranswer.ActiveQuizID(active.ID), quizuseranswer.UserID(userID)).
		Exist(ctx)
	if err != nil {
		logger.Error("Error checking previous quiz answer", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de l'enregistrement de votre réponse."))
		return
	}
	if answered {
		helpers.RespondEphemeralCard(e, ui.Error("Vous avez déjà répondu à ce quiz."))
		return
	}

	correct := perm[idx] == goodAnswerIndex

	err = db.QuizUserAnswer.Create().
		SetID(uuid.New().String()).
		SetActiveQuizID(active.ID).
		SetUserID(userID).
		SetCorrect(correct).
		Exec(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			helpers.RespondEphemeralCard(e, ui.Error("Vous avez déjà répondu à ce quiz."))
			return
		}
		logger.Error("Error saving quiz answer", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de l'enregistrement de votre réponse."))
		return
	}

	if err := recordStat(ctx, userID, correct); err != nil {
		logger.Error("Error updating quiz stats", "error", err, "user", userID)
	}

	helpers.RespondEphemeralCard(e, resultCard(correct, question.GoodAnswer))

	refreshQuizMessage(ctx, e, active, question, applyPermutation(storedAnswers(question), perm))
}

func refreshQuizMessage(ctx context.Context, e *events.ComponentInteractionCreate, active *ent.ActiveQuiz, question *ent.QuizQuestion, answers [answerCount]string) {
	good, bad, err := answerers(ctx, active.ID)
	if err != nil {
		logger.Error("Error loading quiz answerers", "error", err, "quiz", active.ID)
		return
	}

	card := questionCard(question, answers, active.ExpiresAt, good, bad)
	if _, err := e.Client().Rest.UpdateMessage(e.Message.ChannelID, e.Message.ID, card.MessageUpdate()); err != nil {
		logger.Error("Error updating quiz message", "error", err, "quiz", active.ID)
	}
}

func answerers(ctx context.Context, activeQuizID string) ([]string, []string, error) {
	answers, err := database.Default.Ent().QuizUserAnswer.Query().
		Where(quizuseranswer.ActiveQuizID(activeQuizID)).
		Order(quizuseranswer.ByAnsweredAt()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	var good, bad []string
	for _, answer := range answers {
		if answer.Correct {
			good = append(good, answer.UserID)
			continue
		}
		bad = append(bad, answer.UserID)
	}
	return good, bad, nil
}

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
