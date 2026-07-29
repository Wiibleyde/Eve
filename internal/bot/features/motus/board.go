package motus

import (
	"fmt"
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/database/tables"

	"github.com/disgoorg/disgo/discord"
)

const (
	colorWon  = 0x2ECC71
	colorLost = 0xE74C3C
)

func boardEmbed(word string, attempts []tables.MotusAttempt, state string) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Motus"

	var b strings.Builder
	fmt.Fprintf(&b, "Mot de **%d** lettres, commençant par **%s**.\n\n",
		WordLen(word), FirstLetter(word))
	b.WriteString(RenderLetters(MaskWord(word)))
	b.WriteString("\n")

	for _, attempt := range attempts {
		states := Score(attempt.Word, word)
		fmt.Fprintf(&b, "\n%s\n%s — <@%s>\n",
			RenderStates(states), RenderLetters(attempt.Word), attempt.UserID)
	}

	b.WriteString("\n" + Legend)

	switch state {
	case tables.MotusStateWon:
		embed.Color = colorWon
		fmt.Fprintf(&b, "\n\n**Gagné !** Le mot était **%s** (%d/%d).",
			word, len(attempts), MaxAttempts)
	case tables.MotusStateLost:
		embed.Color = colorLost
		fmt.Fprintf(&b, "\n\n**Perdu !** Le mot était **%s** (%d/%d).",
			word, len(attempts), MaxAttempts)
	default:
		fmt.Fprintf(&b, "\n\nEssais : **%d/%d** — cliquez sur « Essayer » pour proposer un mot.",
			len(attempts), MaxAttempts)
	}

	embed.Description = b.String()
	return embed
}

func boardComponents(disabled bool) []discord.LayoutComponent {
	button := discord.NewPrimaryButton("Essayer", CustomIDTry).
		WithEmoji(discord.NewComponentEmoji("✏️"))
	if disabled {
		button = button.AsDisabled()
	}
	return []discord.LayoutComponent{discord.NewActionRow(button)}
}

func boardMessage(word string) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds:     []discord.Embed{boardEmbed(word, nil, tables.MotusStatePlaying)},
		Components: boardComponents(false),
	}
}

func boardUpdate(word string, attempts []tables.MotusAttempt, state string) discord.MessageUpdate {
	embedList := []discord.Embed{boardEmbed(word, attempts, state)}
	components := boardComponents(state != tables.MotusStatePlaying)
	return discord.MessageUpdate{
		Embeds:     &embedList,
		Components: &components,
	}
}

func attemptFeedback(word string, attempts []tables.MotusAttempt) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Motus"

	last := attempts[len(attempts)-1]
	states := Score(last.Word, word)
	embed.Description = fmt.Sprintf("%s\n%s\n\nEssais : **%d/%d**.",
		RenderStates(states), RenderLetters(last.Word), len(attempts), MaxAttempts)
	return embed
}
