package quiz

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/quizstat"
	"Eve/internal/logger"

	"entgo.io/ent/dialect/sql"
	"github.com/disgoorg/disgo/events"
)

const leaderboardSize = 10

const minRatioAnswers = 10

func handleLeaderboard(e *events.ApplicationCommandInteractionCreate) {
	if _, ok := requireGuild(e); !ok {
		return
	}

	data := e.SlashCommandInteractionData()
	choice := data.String("choice")

	ctx := context.Background()

	var (
		stats []*ent.QuizStat
		err   error
		title string
		empty string
	)
	switch choice {
	case choiceBestScores:
		title = "Meilleurs scores"
		empty = "Personne n'a encore de bonne réponse."
		stats, err = bestScores(ctx)
	case choiceBestRatios:
		title = "Meilleurs ratios"
		empty = fmt.Sprintf("Personne n'a encore répondu à %d questions.", minRatioAnswers)
		stats, err = bestRatios(ctx)
	case choiceWorstScores:
		title = "Scores les plus bas"
		empty = "Personne n'a encore de mauvaise réponse."
		stats, err = worstScores(ctx)
	default:
		helpers.RespondEphemeralCard(e, ui.Error("Classement inconnu."))
		return
	}
	if err != nil {
		logger.Error("Error building quiz leaderboard", "error", err, "choice", choice)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de la récupération du classement."))
		return
	}
	if len(stats) == 0 {
		helpers.RespondEphemeralCard(e, ui.Error(empty))
		return
	}

	card := ui.New().
		Accent(colorQuiz).
		Title(quizTitle + " — " + title).
		Text(strings.Join(leaderboardLines(stats), "\n"))

	if err := e.CreateMessage(card.MessageCreate()); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func leaderboardLines(stats []*ent.QuizStat) []string {
	lines := make([]string, 0, len(stats))
	for i, s := range stats {
		total := s.GoodAnswers + s.BadAnswers
		ratio := 0.0
		if total > 0 {
			ratio = float64(s.GoodAnswers) / float64(total) * 100
		}
		lines = append(lines, fmt.Sprintf(
			"**%d.** <@%s> — %d bonne(s) / %d mauvaise(s) — %.0f%%",
			i+1, s.UserID, s.GoodAnswers, s.BadAnswers, ratio,
		))
	}
	return lines
}

func bestScores(ctx context.Context) ([]*ent.QuizStat, error) {
	return database.Default.Ent().QuizStat.Query().
		Where(quizstat.GoodAnswersGT(0)).
		Order(quizstat.ByGoodAnswers(sql.OrderDesc())).
		Limit(leaderboardSize).
		All(ctx)
}

func worstScores(ctx context.Context) ([]*ent.QuizStat, error) {
	return database.Default.Ent().QuizStat.Query().
		Where(quizstat.BadAnswersGT(0)).
		Order(quizstat.ByBadAnswers(sql.OrderDesc())).
		Limit(leaderboardSize).
		All(ctx)
}

func bestRatios(ctx context.Context) ([]*ent.QuizStat, error) {
	all, err := database.Default.Ent().QuizStat.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	eligible := make([]*ent.QuizStat, 0, len(all))
	for _, s := range all {
		if s.GoodAnswers+s.BadAnswers >= minRatioAnswers {
			eligible = append(eligible, s)
		}
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		ri, rj := ratio(eligible[i]), ratio(eligible[j])
		if ri != rj {
			return ri > rj
		}
		return eligible[i].GoodAnswers > eligible[j].GoodAnswers
	})

	if len(eligible) > leaderboardSize {
		eligible = eligible[:leaderboardSize]
	}
	return eligible, nil
}

func ratio(s *ent.QuizStat) float64 {
	total := s.GoodAnswers + s.BadAnswers
	if total == 0 {
		return 0
	}
	return float64(s.GoodAnswers) / float64(total)
}
