package mpthreads

import (
	"context"
	"fmt"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// User-facing strings (French, per project conventions).
const (
	msgDMClosed  = "⚠️ Impossible d'envoyer ce message à %s : ses messages privés sont fermés."
	msgDMError   = "⚠️ Impossible d'envoyer ce message à %s : Discord a refusé l'envoi."
	msgDMPartial = "⚠️ Message envoyé seulement en partie à %s : la suite n'est pas passée."
	msgThreadFmt = "**%s** :"
)

// relayInbound mirrors a DM into the user's staff thread, recreating or
// reopening the thread once if the post is refused. The caller holds the
// conversation lock (see dispatch).
func relayInbound(client *bot.Client, msg discord.Message) {
	ctx := context.Background()

	body, attachments := inboundPayload(msg)

	// The owner gets pinged on every inbound DM; the allowlist is restricted to
	// them so mentions inside the relayed content itself stay inert.
	mentions := noMentions()
	if ownerID, ok := helpers.OwnerID(); ok {
		body = discord.UserMention(ownerID) + " " + body
		mentions = &discord.AllowedMentions{Users: []snowflake.ID{ownerID}}
	}

	for attempt := range 2 {
		threadID, err := ensureThread(ctx, client, msg.Author)
		if err != nil {
			logger.Error("MP threads: preparing thread", "user", msg.Author.ID.String(), "error", err)
			return
		}

		if _, err = sendChunks(client, threadID, body, attachments, mentions); err == nil {
			return
		}

		action := recoverThread(ctx, client, threadID, err)
		if action == recoveryRetry && attempt == 0 {
			continue
		}
		if action != recoveryStop {
			logger.Error("MP threads: relaying DM to thread",
				"user", msg.Author.ID.String(), "thread", threadID.String(), "error", err)
		}
		return
	}
}

// inboundPayload renders a DM as it appears in the thread: author header,
// content, sticker links, then the attachments that could be re-uploaded (the
// others are appended as links).
func inboundPayload(msg discord.Message) (string, []relayAttachment) {
	attachments, skipped := collectAttachments(msg.Attachments)

	var body strings.Builder
	body.WriteString(fmt.Sprintf(msgThreadFmt, msg.Author.EffectiveName()))
	if content := strings.TrimSpace(msg.Content); content != "" {
		body.WriteString("\n")
		body.WriteString(content)
	}
	appendLines(&body, stickerLinks(msg.StickerItems))
	appendLines(&body, skipped)

	return body.String(), attachments
}

// relayOutbound sends a staff message from the thread to the user's DMs and
// reports in the thread when the DM could not be fully delivered. The caller
// holds the conversation lock (see dispatch).
func relayOutbound(client *bot.Client, msg discord.Message, userID snowflake.ID) {
	if strings.TrimSpace(msg.Content) == "" && len(msg.Attachments) == 0 && len(msg.StickerItems) == 0 {
		return
	}

	channel, err := client.Rest.CreateDMChannel(userID)
	if err != nil {
		logger.Warn("MP threads: opening DM channel failed", "user", userID.String(), "error", err)
		warnInThread(client, msg, userID, deliveryWarning(err, 0))
		return
	}

	attachments, skipped := collectAttachments(msg.Attachments)

	var body strings.Builder
	body.WriteString(strings.TrimSpace(msg.Content))
	appendLines(&body, stickerLinks(msg.StickerItems))
	appendLines(&body, skipped)

	sent, err := sendChunks(client, channel.ID(), body.String(), attachments, usersOnlyMentions())
	if err != nil {
		logger.Warn("MP threads: relaying staff message to DM failed",
			"user", userID.String(), "delivered", sent, "error", err)
		warnInThread(client, msg, userID, deliveryWarning(err, sent))
	}
}

// deliveryWarning picks the wording matching the actual failure. Only a 50007
// really means the user closed their DMs, and once part of a long message is
// through, staff must not be told nothing arrived.
func deliveryWarning(err error, delivered int) string {
	switch {
	case delivered > 0:
		return msgDMPartial
	case rest.IsJSONErrorCode(err,
		rest.JSONErrorCodeCannotSendMessagesToThisUser,
		rest.JSONErrorCodeCannotSendMessagesToThisUserDueToHavingNoMutualGuilds):
		return msgDMClosed
	default:
		return msgDMError
	}
}

func warnInThread(client *bot.Client, msg discord.Message, userID snowflake.ID, format string) {
	warning := discord.MessageCreate{
		Content:         fmt.Sprintf(format, discord.UserMention(userID)),
		AllowedMentions: noMentions(),
	}.WithMessageReferenceByID(msg.ID)

	if _, err := client.Rest.CreateMessage(msg.ChannelID, warning); err != nil {
		logger.Error("MP threads: posting delivery warning", "thread", msg.ChannelID.String(), "error", err)
	}
}

// sendChunks posts a body (split to Discord's length limit) plus the
// attachments on the last chunk, and reports how many messages actually
// reached the channel so a caller can tell a total failure from a partial one.
// A body with no content and no attachment sends nothing.
func sendChunks(client *bot.Client, channelID snowflake.ID, body string, attachments []relayAttachment, mentions *discord.AllowedMentions) (int, error) {
	chunks := splitContent(body)
	if len(chunks) == 0 && len(attachments) == 0 {
		return 0, nil
	}
	if len(chunks) == 0 {
		chunks = []string{""}
	}

	delivered := 0
	for i, chunk := range chunks {
		create := discord.MessageCreate{Content: chunk, AllowedMentions: mentions}
		if i == len(chunks)-1 {
			create.Files = filesOf(attachments)
		}

		_, err := client.Rest.CreateMessage(channelID, create)
		if err == nil {
			delivered++
			continue
		}
		if len(create.Files) == 0 || !isPayloadTooLarge(err) {
			return delivered, err
		}

		// maxUploadBytes only estimates the guild's real limit: rather than lose
		// the message, resend it with the attachments degraded to links.
		var fallback strings.Builder
		fallback.WriteString(chunk)
		appendLines(&fallback, linksOf(attachments))
		retried, err := sendChunks(client, channelID, fallback.String(), nil, mentions)
		return delivered + retried, err
	}
	return delivered, nil
}
