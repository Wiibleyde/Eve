package blague

import (
	"context"
	"fmt"
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
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
	maxDescription = 4096
	maxFieldValue  = 1024
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
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgDisabled))
		return
	}

	value := e.SlashCommandInteractionData().String(OptionType)
	label, ok := categoryLabel(value)
	if !ok {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgUnknownType))
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
		editDeferred(e, embeds.ErrorEmbed(MsgFetchFailed), nil)
		return
	}

	editDeferred(e, jokeEmbed(joke, label), []discord.LayoutComponent{
		discord.NewActionRow(discord.NewPrimaryButton(PublishButtonText, CustomIDPublic)),
	})
}

func HandlePublicButton(e *events.ComponentInteractionCreate, _ []string) {
	if len(e.Message.Embeds) == 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgNoEmbed))
		return
	}

	if err := e.DeferUpdateMessage(); err != nil {
		logger.Error("Blague: deferring button update", "error", err)
		return
	}

	_, err := e.Client().Rest.CreateMessage(e.Message.ChannelID, discord.MessageCreate{
		Content:         fmt.Sprintf("Demandée par <@%s>", e.User().ID),
		Embeds:          []discord.Embed{e.Message.Embeds[0]},
		AllowedMentions: &discord.AllowedMentions{Parse: []discord.AllowedMentionType{}},
	})
	if err != nil {
		logger.Error("Blague: publishing joke", "channel", e.Message.ChannelID.String(), "error", err)
		followupEphemeral(e.Client(), e.ApplicationID(), e.Token(), embeds.ErrorEmbed(MsgPublishFailed))
		return
	}

	content := MsgPublished
	noComponents := []discord.LayoutComponent{}
	if _, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{
		Content:    &content,
		Components: &noComponents,
	}); err != nil {
		logger.Error("Blague: removing publish button", "error", err)
	}
}

func jokeEmbed(joke *Joke, label string) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Blague — " + label
	embed.Description = truncate(joke.Joke, maxDescription)

	answer := strings.TrimSpace(joke.Answer)
	value := AnswerFallback
	if answer != "" {
		value = "||" + truncate(answer, maxFieldValue-4) + "||"
	}
	embed.Fields = []discord.EmbedField{{Name: AnswerFieldName, Value: value}}

	if embed.Footer == nil {
		embed.Footer = &discord.EmbedFooter{}
	}
	embed.Footer.Text = FooterDisclaimer
	return embed
}

func editDeferred(e *events.ApplicationCommandInteractionCreate, embed discord.Embed, components []discord.LayoutComponent) {
	embedList := []discord.Embed{embed}
	update := discord.MessageUpdate{Embeds: &embedList}
	if components != nil {
		update.Components = &components
	}
	if _, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), update); err != nil {
		logger.Error("Blague: editing deferred response failed", "error", err)
	}
}

func followupEphemeral(client *bot.Client, appID snowflake.ID, token string, embed discord.Embed) {
	msg := discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}
	if _, err := client.Rest.CreateFollowupMessage(appID, token, msg); err != nil {
		logger.Error("Blague: sending followup", "error", err)
	}
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
