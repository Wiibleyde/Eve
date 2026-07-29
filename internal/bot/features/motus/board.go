package motus

import (
	"fmt"
	"strings"

	"Eve/internal/bot/ui"
	"Eve/internal/database/tables"

	"github.com/disgoorg/disgo/discord"
)

const (
	colorWon  = 0x2ECC71
	colorLost = 0xE74C3C
)

const boardTitle = "🟥 Motus"

func boardCard(word string, attempts []tables.MotusAttempt, state string) *ui.Card {
	card := ui.New().Title(boardTitle)

	card.Textf("Mot de **%d** lettres, commençant par **%s**.\n%s",
		WordLen(word), FirstLetter(word), RenderLetters(MaskWord(word)))

	if len(attempts) > 0 {
		var b strings.Builder
		for _, attempt := range attempts {
			fmt.Fprintf(&b, "%s\n%s — <@%s>\n\n",
				RenderStates(Score(attempt.Word, word)), RenderLetters(attempt.Word), attempt.UserID)
		}
		card.Divider().Text(strings.TrimRight(b.String(), "\n"))
	}

	card.Divider().Subtext(Legend)

	switch state {
	case tables.MotusStateWon:
		card.Accent(colorWon).Textf("🎉 **Gagné !** Le mot était **%s** (%d/%d).", word, len(attempts), MaxAttempts)
	case tables.MotusStateLost:
		card.Accent(colorLost).Textf("💀 **Perdu !** Le mot était **%s** (%d/%d).", word, len(attempts), MaxAttempts)
	default:
		card.Textf("Essais : **%d/%d** — cliquez sur « Essayer » pour proposer un mot.", len(attempts), MaxAttempts)
	}

	return card.Row(tryButton(state != tables.MotusStatePlaying))
}

func tryButton(disabled bool) discord.ButtonComponent {
	button := discord.NewPrimaryButton("Essayer", CustomIDTry).
		WithEmoji(discord.NewComponentEmoji("✏️"))
	if disabled {
		return button.AsDisabled()
	}
	return button
}

func boardMessage(word string) discord.MessageCreate {
	return boardCard(word, nil, tables.MotusStatePlaying).MessageCreate()
}

func boardUpdate(word string, attempts []tables.MotusAttempt, state string) discord.MessageUpdate {
	return boardCard(word, attempts, state).MessageUpdate()
}

func attemptFeedback(word string, attempts []tables.MotusAttempt) *ui.Card {
	last := attempts[len(attempts)-1]
	return ui.New().
		Title(boardTitle).
		Textf("%s\n%s", RenderStates(Score(last.Word, word)), RenderLetters(last.Word)).
		Subtext(fmt.Sprintf("Essais : %d/%d", len(attempts), MaxAttempts))
}
