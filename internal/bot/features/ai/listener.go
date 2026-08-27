package ai

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/maintenance"
	"Eve/internal/bot/ui"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
	"Eve/internal/logger"
	"Eve/internal/ollama"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

type speaker struct {
	id   snowflake.ID
	name string
}

type answer struct {
	content  string
	pingable []snowflake.ID
}

const (
	lookupTimeout   = 3 * time.Second
	generateTimeout = 90 * time.Second
	typingInterval  = 8 * time.Second
	warmupTimeout   = 5 * time.Minute
)

const (
	msgEmptyPrompt = "Oui ? " + EmojiNeutral
	msgBusy        = "Je réfléchis déjà à autre chose, réessaie dans quelques secondes."
	msgFailed      = "Impossible de générer une réponse pour le moment."
)

func Attach(client *bot.Client) {
	if !ollama.Enabled() {
		logger.Warn("AI: " + ollama.EnvURL + " is missing, mention replies are disabled")
		return
	}
	client.AddEventListeners(bot.NewListenerFunc(onMessageCreate))
	go warmup()
	logger.Info("AI: mention replies attached", "model", ollama.Default().Model())
}

func warmup() {
	defer recoverPanic("ai warmup")

	ctx, cancel := context.WithTimeout(context.Background(), warmupTimeout)
	defer cancel()

	if err := ollama.Default().Warmup(ctx); err != nil {
		logger.Warn("AI: warmup failed, first reply will be slower", "error", err)
		return
	}
	logger.Info("AI: model warmed up", "model", ollama.Default().Model())
}

func onMessageCreate(e *events.MessageCreate) {
	if e.GuildID == nil {
		return
	}
	msg := e.Message
	if msg.Author.Bot || msg.Author.System || msg.WebhookID != nil {
		return
	}
	if maintenance.Enabled() {
		return
	}
	if msg.MentionEveryone || !addressesSelf(e) {
		return
	}

	guildID := *e.GuildID
	client := e.Client()
	selfID := client.ID()
	channelID := e.ChannelID
	messageID := msg.ID
	author := speaker{id: msg.Author.ID, name: authorName(msg)}
	prompt := cleanPrompt(msg.Content, selfID)

	go func() {
		defer recoverPanic("ai reply")

		if disabledForGuild(guildID) {
			logger.Debug("AI: reply skipped, disabled for guild", "guild", guildID.String())
			return
		}
		if prompt == "" {
			reply(client, channelID, messageID, answer{content: msgEmptyPrompt})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), generateTimeout)
		defer cancel()
		go keepTyping(ctx, client, channelID)

		answer, err := generate(ctx, selfID, channelID, author, prompt)
		if err != nil {
			respondError(client, channelID, messageID, err)
			return
		}
		reply(client, channelID, messageID, answer)
	}()
}

func generate(ctx context.Context, selfID snowflake.ID, channelID snowflake.ID, author speaker, prompt string) (answer, error) {
	now := time.Now()
	ownerID, ownerKnown := helpers.OwnerID()
	question := ollama.Message{Role: ollama.RoleUser, Content: author.prefix() + prompt}

	conversation := append([]ollama.Message{systemMessage(selfID, ownerID, ownerKnown)}, channelHistory.messages(channelID, now)...)
	if grounding, ok := groundingMessage(ctx, prompt); ok {
		conversation = append(conversation, grounding)
	}
	conversation = append(conversation, question)

	raw, err := ollama.Default().Chat(ctx, conversation)
	if err != nil {
		return answer{}, err
	}

	content := cleanReply(raw, selfID)
	if content == "" {
		return answer{}, errors.New("ai: model returned an unusable answer")
	}

	channelHistory.append(channelID, now, author.id, question, ollama.Message{
		Role:    ollama.RoleAssistant,
		Content: content,
	})

	return answer{
		content:  content,
		pingable: pingable(content, allowedTargets(channelID, now, author.id, ownerID, ownerKnown)),
	}, nil
}

func allowedTargets(channelID snowflake.ID, now time.Time, authorID snowflake.ID, ownerID snowflake.ID, ownerKnown bool) map[snowflake.ID]struct{} {
	allowed := map[snowflake.ID]struct{}{authorID: {}}
	if ownerKnown {
		allowed[ownerID] = struct{}{}
	}
	for _, id := range channelHistory.participants(channelID, now) {
		allowed[id] = struct{}{}
	}
	return allowed
}

func pingable(content string, allowed map[snowflake.ID]struct{}) []snowflake.ID {
	ids := make([]snowflake.ID, 0, len(allowed))
	for _, id := range mentionedUsers(content) {
		if _, ok := allowed[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func keepTyping(ctx context.Context, client *bot.Client, channelID snowflake.ID) {
	defer recoverPanic("ai typing")

	ticker := time.NewTicker(typingInterval)
	defer ticker.Stop()

	for {
		if err := client.Rest.SendTyping(channelID); err != nil {
			logger.Debug("AI: typing indicator failed", "channel", channelID.String(), "error", err)
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func addressesSelf(e *events.MessageCreate) bool {
	selfID := e.Client().ID()
	for _, user := range e.Message.Mentions {
		if user.ID == selfID {
			return true
		}
	}
	referenced := e.Message.ReferencedMessage
	return referenced != nil && referenced.Author.ID == selfID
}

func (s speaker) prefix() string {
	return s.name + " (" + mention(s.id) + ") : "
}

func authorName(msg discord.Message) string {
	if msg.Member != nil {
		return msg.Member.EffectiveName()
	}
	return msg.Author.EffectiveName()
}

func disabledForGuild(guildID snowflake.ID) bool {
	if database.Default == nil {
		logger.Warn("AI: database unavailable, staying silent")
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), lookupTimeout)
	defer cancel()

	cfg, err := database.Default.Ent().GuildConfig.Query().
		Where(
			guildconfig.GuildID(guildID.String()),
			guildconfig.Key(tables.AIDisabled.String()),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false
		}
		logger.Error("Error reading ai.disabled", "guild", guildID.String(), "error", err)
		return true
	}
	return parseBool(cfg.Value)
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "oui", "yes", "on":
		return true
	default:
		return false
	}
}

func respondError(client *bot.Client, channelID snowflake.ID, messageID snowflake.ID, err error) {
	if errors.Is(err, ollama.ErrBusy) {
		logger.Debug("AI: generation refused, queue is full", "channel", channelID.String())
		replyCard(client, channelID, messageID, ui.Warning("Trop de demandes", msgBusy))
		return
	}
	logger.Error("AI: generation failed", "channel", channelID.String(), "error", err)
	replyCard(client, channelID, messageID, ui.Error(msgFailed))
}

func reply(client *bot.Client, channelID snowflake.ID, messageID snowflake.ID, a answer) {
	send(client, channelID, discord.MessageCreate{
		Content: a.content,
		AllowedMentions: &discord.AllowedMentions{
			Users:       a.pingable,
			RepliedUser: len(a.pingable) == 0,
		},
		MessageReference: reference(channelID, messageID),
	})
}

func replyCard(client *bot.Client, channelID snowflake.ID, messageID snowflake.ID, card *ui.Card) {
	message := card.MessageCreate()
	message.MessageReference = reference(channelID, messageID)
	send(client, channelID, message)
}

func send(client *bot.Client, channelID snowflake.ID, message discord.MessageCreate) {
	if _, err := client.Rest.CreateMessage(channelID, message); err != nil {
		logger.Error("AI: sending reply failed", "channel", channelID.String(), "error", err)
	}
}

func reference(channelID snowflake.ID, messageID snowflake.ID) *discord.MessageReference {
	return &discord.MessageReference{
		Type:      discord.MessageReferenceTypeDefault,
		MessageID: &messageID,
		ChannelID: &channelID,
	}
}

func recoverPanic(what string) {
	rec := recover()
	if rec == nil {
		return
	}
	logger.Error("Panic in AI listener",
		"what", what,
		"panic", fmt.Sprint(rec),
		"stack", strings.ReplaceAll(string(debug.Stack()), "\n", " | "),
	)
}
