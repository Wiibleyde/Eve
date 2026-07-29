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

// No author icon URL is set on purpose: the TS version pointed at a signed
// cdn.discordapp.com attachment which expires.
const authorName = "Quiz"

// buttonLabelLimit is the Discord hard limit for a button label, in characters.
const buttonLabelLimit = 80

// storedAnswers returns the answers of a question in storage order: index 0 is
// always the good one.
func storedAnswers(q *ent.QuizQuestion) [answerCount]string {
	return [answerCount]string{q.GoodAnswer, q.BadAnswer1, q.BadAnswer2, q.BadAnswer3}
}

func questionEmbed(q *ent.QuizQuestion, expiresAt time.Time) discord.Embed {
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
		{Name: "Expiration", Value: fmt.Sprintf("<t:%d:R>", expiresAt.Unix())},
	}
	return embed
}

// The question id travels in the custom ID next to the displayed index so that
// a click on a quiz whose active row is gone can still reveal the answer
// without matching on the rendered question text, which is escaped.
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

// questionTextFromMessage reads the question back from the posted embed. It is
// the last-resort fallback for buttons posted before the question id was
// carried in the custom ID; it misses when the question contains a code fence,
// since escapeCodeBlock rewrote it.
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

// escapeCodeBlock neutralises fences so a crafted question cannot break out of
// the code block it is rendered in.
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

// truncate shortens a string to at most limit characters, ellipsis included.
func truncate(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit-1]) + "…"
}
