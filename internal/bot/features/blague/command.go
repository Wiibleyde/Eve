// Package blague serves random French jokes from blagues-api.fr.
//
// /blague answers ephemerally with the joke as the embed description and the
// punchline hidden behind a ||spoiler|| field, plus a "Rendre publique" button.
//
// The button is stateless: it re-reads the embed off the interaction message
// (Discord ships the full message with every component interaction), reposts it
// publicly with an attribution line, then strips the button from the ephemeral.
// Nothing is stored between the two interactions — no cache, no database, no
// joke ID smuggled through the custom ID.
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
	CommandName = "blague"
	// OptionType is the required category option.
	OptionType = "type"
	// CustomIDPublic is the "<feature>:<action>" custom ID of the publish
	// button. It carries no data segment on purpose.
	CustomIDPublic = "blague:public"
)

// User-facing strings (French, per project conventions).
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

// Discord payload limits we clamp to before sending.
const (
	maxDescription = 4096
	maxFieldValue  = 1024
)

// command is only published when a token is configured; see Commands.
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

// Commands returns the command set to register on Discord.
//
// It is a function, not a package var, because it depends on BLAGUE_API_TOKEN
// which only exists once godotenv has run — package var initialisers execute
// before main(). Without a token nothing is returned, so the command is never
// published rather than published and broken (the TypeScript version's bug).
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

	// The HTTP call can take up to RequestTimeout, well past the 3s ACK
	// deadline, so acknowledge first and edit the deferred response once the
	// joke is in hand.
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

// HandlePublicButton reposts the joke publicly and removes the button from the
// ephemeral. The embed is read straight off the interaction message, so the
// handler holds no state at all.
func HandlePublicButton(e *events.ComponentInteractionCreate, _ []string) {
	if len(e.Message.Embeds) == 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgNoEmbed))
		return
	}

	// Deferred (not visible) rather than DeferCreateMessage: acknowledging
	// immediately frees up the REST call below from the 3s ACK budget without
	// flashing a loading state on the button.
	if err := e.DeferUpdateMessage(); err != nil {
		logger.Error("Blague: deferring button update", "error", err)
		return
	}

	// Publish first: if it fails the ephemeral keeps its button and the user
	// can simply try again.
	_, err := e.Client().Rest.CreateMessage(e.Message.ChannelID, discord.MessageCreate{
		Content: fmt.Sprintf("Demandée par <@%s>", e.User().ID),
		Embeds:  []discord.Embed{e.Message.Embeds[0]},
		// The attribution renders as a mention but must not ping.
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
		// Leave room for the surrounding spoiler markers.
		value = "||" + truncate(answer, maxFieldValue-4) + "||"
	}
	embed.Fields = []discord.EmbedField{{Name: AnswerFieldName, Value: value}}

	if embed.Footer == nil {
		embed.Footer = &discord.EmbedFooter{}
	}
	embed.Footer.Text = FooterDisclaimer
	return embed
}

// editDeferred replaces a DeferCreateMessage response with an embed and,
// optionally, action row components.
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

// followupEphemeral sends an ephemeral embed as a followup to an
// already-acknowledged interaction (used after DeferUpdateMessage, which
// leaves no in-place response to edit for an error).
func followupEphemeral(client *bot.Client, appID snowflake.ID, token string, embed discord.Embed) {
	msg := discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}
	if _, err := client.Rest.CreateFollowupMessage(appID, token, msg); err != nil {
		logger.Error("Blague: sending followup", "error", err)
	}
}

// truncate clamps s to max runes, appending an ellipsis when it cuts.
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
