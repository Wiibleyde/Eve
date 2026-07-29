package quiz

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"Eve/internal/bot/router"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"

	"github.com/disgoorg/disgo/discord"
)

const (
	colorQuiz    = 0x5865F2
	colorCorrect = 0x43B581
	colorWrong   = 0xFF4444
	colorExpired = 0xFFA500
)

const quizTitle = "🧠 Quiz"

const (
	buttonLabelLimit = 80
	mentionListLimit = 900
)

const (
	goodAnswersField = "✅ Bonne(s) réponse(s)"
	badAnswersField  = "❌ Mauvaise(s) réponse(s)"
)

func storedAnswers(q *ent.QuizQuestion) [answerCount]string {
	return [answerCount]string{q.GoodAnswer, q.BadAnswer1, q.BadAnswer2, q.BadAnswer3}
}

func questionCard(q *ent.QuizQuestion, answers [answerCount]string, expiresAt time.Time, good, bad []string) *ui.Card {
	card := ui.New().
		Accent(colorQuiz).
		Title(quizTitle+" — Nouvelle question !").
		Text("```\n"+escapeCodeBlock(q.Question)+"\n```").
		Fields(
			ui.Field{Name: "Catégorie", Value: fallbackText(q.Category), Inline: true},
			ui.Field{Name: "Difficulté", Value: fallbackText(q.Difficulty), Inline: true},
			ui.Field{Name: "Proposée par", Value: authorMention(q.AuthorID), Inline: true},
			ui.Field{Name: "Expiration", Value: fmt.Sprintf("<t:%d:R>", expiresAt.Unix()), Inline: true},
		)

	if len(good) > 0 || len(bad) > 0 {
		card.Divider()
	}
	if len(good) > 0 {
		card.Fields(ui.Field{Name: goodAnswersField, Value: mentionList(good)})
	}
	if len(bad) > 0 {
		card.Fields(ui.Field{Name: badAnswersField, Value: mentionList(bad)})
	}

	return card.Row(answerButtons(answers, q.ID)...)
}

func mentionList(userIDs []string) string {
	var b strings.Builder
	for i, userID := range userIDs {
		separator := ""
		if b.Len() > 0 {
			separator = " "
		}
		mention := separator + "<@" + userID + ">"
		overflow := separator + fmt.Sprintf("… et %d autre(s)", len(userIDs)-i)
		if b.Len()+len(mention) > mentionListLimit-len(overflow) {
			b.WriteString(overflow)
			break
		}
		b.WriteString(mention)
	}
	return b.String()
}

func answerButtons(answers [answerCount]string, questionID string) []discord.InteractiveComponent {
	components := make([]discord.InteractiveComponent, 0, answerCount)
	for i, answer := range answers {
		customID := router.BuildCustomID(ButtonAnswerPrefix, strconv.Itoa(i), questionID)
		components = append(components, discord.NewSecondaryButton(truncate(answer, buttonLabelLimit), customID))
	}
	return components
}

func resultCard(correct bool, goodAnswer string) *ui.Card {
	if correct {
		return ui.New().
			Accent(colorCorrect).
			Title(quizTitle + " — Bonne réponse !").
			Text("Bravo, c'était bien " + inlineCode(goodAnswer) + ".")
	}
	return ui.New().
		Accent(colorWrong).
		Title(quizTitle + " — Mauvaise réponse").
		Text("Dommage, la bonne réponse était " + inlineCode(goodAnswer) + ".")
}

func expiredCard(goodAnswer string) *ui.Card {
	card := ui.New().Accent(colorExpired).Title(quizTitle + " — Terminé")
	if goodAnswer == "" {
		return card.Text("Ce quiz est terminé et sa réponse n'est plus disponible.")
	}
	return card.Text("Ce quiz est terminé. La bonne réponse était " + inlineCode(goodAnswer) + ".")
}

func questionTextFromMessage(msg discord.Message) (string, bool) {
	for _, text := range ui.Texts(msg.Components) {
		if content, ok := codeBlockContent(text); ok && content != "" {
			return content, true
		}
	}
	return "", false
}

func codeBlockContent(s string) (string, bool) {
	const fence = "```"
	start := strings.Index(s, fence)
	if start < 0 {
		return "", false
	}
	rest := s[start+len(fence):]
	end := strings.Index(rest, fence)
	if end < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:end]), true
}

func escapeCodeBlock(s string) string {
	return strings.ReplaceAll(s, "```", "'''")
}

func inlineCode(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "'") + "`"
}

func authorMention(authorID string) string {
	if authorID == "" {
		return "Inconnu"
	}
	return "<@" + authorID + ">"
}

func fallbackText(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func truncate(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit-1]) + "…"
}
