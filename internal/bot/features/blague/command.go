package blague

import (
	"context"
	"fmt"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const (
	CommandName    = "blague"
	OptionType     = "type"
	CustomIDPublic = "blague:public"
)

const (
	MsgDisabled       = "La fonctionnalité blagues est désactivée : aucun jeton d'API n'est configuré."
	MsgUnknownType    = "Type de blague inconnu."
	MsgFetchFailed    = "Impossible de récupérer une blague pour le moment. Réessayez plus tard."
	MsgNoEmbed        = "Impossible de retrouver la blague à publier."
	MsgPublishFailed  = "Impossible de publier la blague dans ce salon."
	MsgPublished      = "Blague publiée !"
	AnswerFieldName   = "Réponse :"
	AnswerFallback    = "*(aucune réponse fournie)*"
	PublishButtonText = "Rendre publique"
	FooterDisclaimer  = "⚠️ Eve et ses développeurs ne sont pas responsables des blagues proposées. ⚠️"
)

const (
	jokeColor = 0xFEE75C
	maxJoke   = 2000
	maxAnswer = 800
)

var command = discord.SlashCommandCreate{
	Name:        CommandName,
	Description: "Obtenir une blague aléatoire",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        OptionType,
			Description: "Catégorie de la blague",
			Required:    true,
			Choices:     typeChoices(),
		},
	},
}

func typeChoices() []discord.ApplicationCommandOptionChoiceString {
	choices := make([]discord.ApplicationCommandOptionChoiceString, 0, len(categories))
	for _, c := range categories {
		choices = append(choices, discord.ApplicationCommandOptionChoiceString{
			Name:  c.Label,
			Value: c.Value,
		})
	}
	return choices
}

func Commands() []discord.ApplicationCommandCreate {
	if !Enabled() {
		return nil
	}
	return []discord.ApplicationCommandCreate{command}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	if !Enabled() {
		helpers.RespondEphemeralCard(e, ui.Error(MsgDisabled))
		return
	}

	value := e.SlashCommandInteractionData().String(OptionType)
	label, ok := categoryLabel(value)
	if !ok {
		helpers.RespondEphemeralCard(e, ui.Error(MsgUnknownType))
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Blague: deferring interaction", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	joke, err := fetchRandom(ctx, value)
	if err != nil {
		logger.Error("Blague: fetching joke", "category", value, "error", err)
		editDeferred(e, ui.Error(MsgFetchFailed))
		return
	}

	editDeferred(e, jokeCard(joke, label).Row(
		discord.NewPrimaryButton(PublishButtonText, CustomIDPublic).
			WithEmoji(discord.NewComponentEmoji("📢")),
	))
}

func HandlePublicButton(e *events.ComponentInteractionCreate, _ []string) {
	joke := publishedCard(e.Message.Components, e.User().ID)
	if joke == nil {
		helpers.RespondEphemeralCard(e, ui.Error(MsgNoEmbed))
		return
	}

	if err := e.DeferUpdateMessage(); err != nil {
		logger.Error("Blague: deferring button update", "error", err)
		return
	}

	if _, err := e.Client().Rest.CreateMessage(e.Message.ChannelID, joke.MessageCreate()); err != nil {
		logger.Error("Blague: publishing joke", "channel", e.Message.ChannelID.String(), "error", err)
		helpers.FollowupEphemeralCard(e.Client(), e.ApplicationID(), e.Token(), ui.Error(MsgPublishFailed))
		return
	}

	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), ui.Success(MsgPublished))
}

func jokeCard(joke *Joke, label string) *ui.Card {
	answer := AnswerFallback
	if trimmed := strings.TrimSpace(joke.Answer); trimmed != "" {
		answer = "||" + truncate(trimmed, maxAnswer) + "||"
	}

	return ui.New().
		Accent(jokeColor).
		Title("😂 Blague — " + label).
		Text(truncate(joke.Joke, maxJoke)).
		Divider().
		Fields(ui.Field{Name: AnswerFieldName, Value: answer}).
		Footer(FooterDisclaimer)
}

func publishedCard(components []discord.LayoutComponent, requester snowflake.ID) *ui.Card {
	texts := ui.Texts(components)
	if len(texts) < 2 {
		return nil
	}

	card := ui.New().Accent(jokeColor)
	for _, text := range texts {
		if strings.HasPrefix(text, "## ") {
			card.Title(strings.TrimPrefix(text, "## "))
			continue
		}
		if strings.HasPrefix(text, "-# ") {
			continue
		}
		card.Text(text)
	}
	return card.Divider().
		Subtext(fmt.Sprintf("Demandée par <@%s>", requester)).
		Footer(FooterDisclaimer)
}

func editDeferred(e *events.ApplicationCommandInteractionCreate, card *ui.Card) {
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), card)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}
