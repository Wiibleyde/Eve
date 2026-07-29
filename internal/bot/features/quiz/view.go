package quiz

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/router"
	"Eve/internal/database/ent"

	"github.com/disgoorg/disgo/discord"
)

const (
	colorQuiz    = 0x5865F2
	colorCorrect = 0x43B581
	colorWrong   = 0xFF4444
	colorExpired = 0xFFA500
)

const authorName = "Quiz"

const (
	buttonLabelLimit = 80
	fieldValueLimit  = 1024
)

const (
	goodAnswersField = "Bonne(s) réponse(s)"
	badAnswersField  = "Mauvaise(s) réponse(s)"
)

func storedAnswers(q *ent.QuizQuestion) [answerCount]string {
	return [answerCount]string{q.GoodAnswer, q.BadAnswer1, q.BadAnswer2, q.BadAnswer3}
}

func questionEmbed(q *ent.QuizQuestion, expiresAt time.Time, good, bad []string) discord.Embed {
	inline := true

	embed := embeds.BaseEmbed()
	embed.Color = colorQuiz
	embed.Author = &discord.EmbedAuthor{Name: authorName}
	embed.Title = "Nouvelle question !"
	embed.Description = "```\n" + escapeCodeBlock(q.Question) + "\n```"
	embed.Fields = []discord.EmbedField{
		{Name: "Catégorie", Value: fallbackText(q.Category), Inline: &inline},
		{Name: "Difficulté", Value: fallbackText(q.Difficulty), Inline: &inline},
		{Name: "Proposée par", Value: authorMention(q.AuthorID), Inline: &inline},
		{Name: "Expiration", Value: fmt.Sprintf("<t:%d:R>", expiresAt.Unix()), Inline: &inline},
	}
	if len(good) > 0 {
		embed.Fields = append(embed.Fields, discord.EmbedField{Name: goodAnswersField, Value: mentionList(good), Inline: &inline})
	}
	if len(bad) > 0 {
		embed.Fields = append(embed.Fields, discord.EmbedField{Name: badAnswersField, Value: mentionList(bad), Inline: &inline})
	}
	return embed
}

func mentionList(userIDs []string) string {
	var b strings.Builder
	for i, userID := range userIDs {
		separator := ""
		if b.Len() > 0 {
			separator = "\n"
		}
		mention := separator + "<@" + userID + ">"
		overflow := separator + fmt.Sprintf("… et %d autre(s)", len(userIDs)-i)
		if b.Len()+len(mention) > fieldValueLimit-len(overflow) {
			b.WriteString(overflow)
			break
		}
		b.WriteString(mention)
	}
	return b.String()
}

func answerButtons(answers [answerCount]string, questionID string) discord.ActionRowComponent {
	components := make([]discord.InteractiveComponent, 0, answerCount)
	for i, answer := range answers {
		customID := router.BuildCustomID(ButtonAnswerPrefix, strconv.Itoa(i), questionID)
		components = append(components, discord.NewSecondaryButton(truncate(answer, buttonLabelLimit), customID))
	}
	return discord.NewActionRow(components...)
}

func resultEmbed(correct bool, goodAnswer string) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Author = &discord.EmbedAuthor{Name: authorName}
	if correct {
		embed.Color = colorCorrect
		embed.Title = "Bonne réponse !"
		embed.Description = "Bravo, c'était bien " + inlineCode(goodAnswer) + "."
		return embed
	}
	embed.Color = colorWrong
	embed.Title = "Mauvaise réponse"
	embed.Description = "Dommage, la bonne réponse était " + inlineCode(goodAnswer) + "."
	return embed
}

func expiredEmbed(goodAnswer string) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Author = &discord.EmbedAuthor{Name: authorName}
	embed.Color = colorExpired
	embed.Title = "Quiz terminé"
	if goodAnswer == "" {
		embed.Description = "Ce quiz est terminé et sa réponse n'est plus disponible."
		return embed
	}
	embed.Description = "Ce quiz est terminé. La bonne réponse était " + inlineCode(goodAnswer) + "."
	return embed
}

func questionTextFromMessage(msg discord.Message) (string, bool) {
	for _, embed := range msg.Embeds {
		if text, ok := codeBlockContent(embed.Description); ok && text != "" {
			return text, true
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
