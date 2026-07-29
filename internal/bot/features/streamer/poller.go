package streamer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"Eve/internal/bot/maintenance"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"
	"Eve/internal/logger"
	"Eve/internal/twitch"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

const (
	PollInterval = 60 * time.Second

	ViewerEditInterval = 5 * time.Minute

	tickTimeout = 45 * time.Second

	userCacheTTL = 6 * time.Hour

	offlineTolerance = 2
)

type trackState struct {
	seen      bool
	streamID  string
	title     string
	game      string
	viewers   int
	startedAt time.Time
	lastEdit  time.Time
	missed    int
}

func (s *trackState) reset() { *s = trackState{} }

func (s *trackState) observe(stream twitch.Stream) {
	s.seen = true
	s.streamID = stream.ID
	s.title = stream.Title
	s.game = stream.GameName
	s.viewers = stream.ViewerCount
	s.startedAt = stream.StartedAt
	s.missed = 0
}

type cachedUser struct {
	user    twitch.User
	fetched time.Time
}

type poller struct {
	client *bot.Client
	states map[int]*trackState
	users  map[string]cachedUser
}

var pollerOnce sync.Once

func StartPoller(client *bot.Client) {
	if !Enabled() {
		warnDisabled()
		return
	}

	pollerOnce.Do(func() {
		go func() {
			p := &poller{
				client: client,
				states: make(map[int]*trackState),
				users:  make(map[string]cachedUser),
			}
			p.tick()

			ticker := time.NewTicker(PollInterval)
			defer ticker.Stop()
			for range ticker.C {
				p.tick()
			}
		}()
	})
}

func (p *poller) tick() {
	if maintenance.Enabled() {
		logger.Debug("Twitch poll skipped: maintenance mode")
		return
	}
	db := entClient()
	if db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), tickTimeout)
	defer cancel()

	rows, err := db.Stream.Query().All(ctx)
	if err != nil {
		logger.Error("Querying tracked streams", "error", err)
		return
	}
	if len(rows) == 0 {
		clear(p.states)
		return
	}

	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.TwitchUserID)
	}

	streams, err := helixClient().GetStreamsByUserIDs(ctx, userIDs)
	if err != nil {
		logger.Error("Fetching Twitch streams", "error", err)
		return
	}

	live := make(map[string]twitch.Stream, len(streams))
	for _, s := range streams {
		if s.IsLive() {
			live[s.UserID] = s
		}
	}

	tracked := make(map[int]struct{}, len(rows))
	for _, row := range rows {
		tracked[row.ID] = struct{}{}
		if s, ok := live[row.TwitchUserID]; ok {
			p.handleLive(ctx, row, s)
			continue
		}
		p.handleOffline(ctx, row)
	}

	for id := range p.states {
		if _, ok := tracked[id]; !ok {
			delete(p.states, id)
		}
	}
}

func (p *poller) handleLive(ctx context.Context, row *ent.Stream, s twitch.Stream) {
	st, ok := p.states[row.ID]
	if !ok {
		st = &trackState{}
		p.states[row.ID] = st
	}
	st.missed = 0

	if row.MessageID != "" && st.seen && st.streamID != s.ID {
		p.endNotification(ctx, row, st)
	}

	if row.MessageID == "" {
		p.startNotification(ctx, row, s, st)
		return
	}
	p.updateNotification(ctx, row, s, st)
}

func (p *poller) handleOffline(ctx context.Context, row *ent.Stream) {
	st := p.states[row.ID]
	if row.MessageID == "" {
		if st != nil {
			delete(p.states, row.ID)
		}
		return
	}

	if st == nil {
		st = &trackState{}
		p.states[row.ID] = st
	}
	st.missed++
	if st.missed < offlineTolerance {
		logger.Debug("Twitch stream missing from the live set, waiting for confirmation",
			"guild", row.GuildID, "login", row.TwitchLogin, "misses", st.missed)
		return
	}

	p.endNotification(ctx, row, st)
	delete(p.states, row.ID)
}

func (p *poller) startNotification(ctx context.Context, row *ent.Stream, s twitch.Stream, st *trackState) {
	channelID, err := snowflake.Parse(row.ChannelID)
	if err != nil {
		logger.Error("Invalid notification channel ID", "guild", row.GuildID, "channel", row.ChannelID, "error", err)
		return
	}

	user, hasUser := p.userInfo(ctx, row.TwitchUserID)
	mention, mentions := roleMention(row.RoleID)

	create := liveCard(s, user, hasUser, mention).MessageCreate()
	create.AllowedMentions = mentions

	msg, err := p.client.Rest.CreateMessage(channelID, create)
	if err != nil {
		logger.Error("Sending Twitch live notification",
			"guild", row.GuildID, "channel", row.ChannelID, "login", row.TwitchLogin, "error", err)
		return
	}

	messageID := msg.ID.String()
	if err := entClient().Stream.UpdateOneID(row.ID).SetMessageID(messageID).Exec(ctx); err != nil {
		logger.Error("Storing live notification message ID", "guild", row.GuildID, "login", row.TwitchLogin, "error", err)
		if delErr := p.client.Rest.DeleteMessage(channelID, msg.ID); delErr != nil {
			logger.Error("Deleting orphaned live notification", "channel", row.ChannelID, "error", delErr)
		}
		return
	}

	row.MessageID = messageID
	st.observe(s)
	st.lastEdit = time.Now()

	logger.Info("Twitch live notification sent",
		"guild", row.GuildID, "login", s.UserLogin, "game", s.GameName, "viewers", s.ViewerCount)
}

func (p *poller) updateNotification(ctx context.Context, row *ent.Stream, s twitch.Stream, st *trackState) {
	resume := !st.seen
	if !shouldEdit(st, s, time.Now()) {
		st.streamID = s.ID
		return
	}

	user, hasUser := p.userInfo(ctx, row.TwitchUserID)
	mention, _ := roleMention(row.RoleID)
	if !p.editMessage(ctx, row, liveCard(s, user, hasUser, mention)) {
		return
	}

	st.observe(s)
	st.lastEdit = time.Now()
	logger.Debug("Twitch live notification updated",
		"guild", row.GuildID, "login", s.UserLogin, "resume", resume, "viewers", s.ViewerCount)
}

func shouldEdit(st *trackState, s twitch.Stream, now time.Time) bool {
	if !st.seen {
		return true
	}
	if st.title != s.Title || st.game != s.GameName {
		return true
	}
	if st.viewers == s.ViewerCount {
		return false
	}
	return now.Sub(st.lastEdit) >= ViewerEditInterval
}

func (p *poller) endNotification(ctx context.Context, row *ent.Stream, st *trackState) {
	if row.MessageID == "" {
		return
	}

	user, hasUser := p.userInfo(ctx, row.TwitchUserID)
	p.editMessage(ctx, row, endedCard(st, row.TwitchLogin, user, hasUser))

	if err := entClient().Stream.UpdateOneID(row.ID).ClearMessageID().Exec(ctx); err != nil {
		logger.Error("Clearing live notification message ID", "guild", row.GuildID, "login", row.TwitchLogin, "error", err)
		return
	}
	row.MessageID = ""
	if st != nil {
		st.reset()
	}

	logger.Info("Twitch stream ended", "guild", row.GuildID, "login", row.TwitchLogin)
}

func (p *poller) editMessage(ctx context.Context, row *ent.Stream, card *ui.Card) bool {
	channelID, err := snowflake.Parse(row.ChannelID)
	if err != nil {
		logger.Error("Invalid notification channel ID", "guild", row.GuildID, "channel", row.ChannelID, "error", err)
		return false
	}
	messageID, err := snowflake.Parse(row.MessageID)
	if err != nil {
		logger.Error("Invalid notification message ID", "guild", row.GuildID, "message", row.MessageID, "error", err)
		return false
	}

	_, err = p.client.Rest.UpdateMessage(channelID, messageID, card.MessageUpdate())
	if err == nil {
		return true
	}

	if isNotFound(err) {
		logger.Warn("Live notification message is gone, dropping its reference",
			"guild", row.GuildID, "login", row.TwitchLogin)
		if clearErr := entClient().Stream.UpdateOneID(row.ID).ClearMessageID().Exec(ctx); clearErr != nil {
			logger.Error("Clearing live notification message ID", "guild", row.GuildID, "error", clearErr)
		}
		row.MessageID = ""
		if st, ok := p.states[row.ID]; ok {
			st.reset()
		}
		return false
	}

	logger.Error("Editing Twitch live notification",
		"guild", row.GuildID, "login", row.TwitchLogin, "error", err)
	return false
}

func (p *poller) userInfo(ctx context.Context, userID string) (twitch.User, bool) {
	if cached, ok := p.users[userID]; ok && time.Since(cached.fetched) < userCacheTTL {
		return cached.user, true
	}

	user, found, err := helixClient().GetUserByID(ctx, userID)
	if err != nil {
		logger.Debug("Fetching Twitch user profile", "twitch_user", userID, "error", err)
		return twitch.User{}, false
	}
	if !found {
		return twitch.User{}, false
	}

	p.users[userID] = cachedUser{user: user, fetched: time.Now()}
	return user, true
}

func roleMention(roleID string) (string, *discord.AllowedMentions) {
	if roleID == "" {
		return "", noMentions()
	}
	id, err := snowflake.Parse(roleID)
	if err != nil {
		logger.Error("Invalid notification role ID", "role", roleID, "error", err)
		return "", noMentions()
	}
	mentions := noMentions()
	mentions.Roles = []snowflake.ID{id}
	return "<@&" + roleID + ">", mentions
}

func noMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{Parse: []discord.AllowedMentionType{}}
}

func closeNotifications(client *bot.Client, rows ...*ent.Stream) {
	pending := make([]*ent.Stream, 0, len(rows))
	for _, row := range rows {
		if row.MessageID != "" {
			pending = append(pending, row)
		}
	}
	if len(pending) == 0 {
		return
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("Panic while closing Twitch live notifications",
					"panic", fmt.Sprint(rec),
					"stack", strings.ReplaceAll(string(debug.Stack()), "\n", " | "),
				)
			}
		}()
		for _, row := range pending {
			closeNotification(client, row)
		}
	}()
}

func closeNotification(client *bot.Client, row *ent.Stream) {
	channelID, err := snowflake.Parse(row.ChannelID)
	if err != nil {
		logger.Error("Invalid notification channel ID", "guild", row.GuildID, "channel", row.ChannelID, "error", err)
		return
	}
	messageID, err := snowflake.Parse(row.MessageID)
	if err != nil {
		logger.Error("Invalid notification message ID", "guild", row.GuildID, "message", row.MessageID, "error", err)
		return
	}

	_, err = client.Rest.UpdateMessage(channelID, messageID,
		endedCard(nil, row.TwitchLogin, twitch.User{}, false).MessageUpdate())
	if err != nil && !isNotFound(err) {
		logger.Error("Closing Twitch live notification",
			"guild", row.GuildID, "login", row.TwitchLogin, "error", err)
		return
	}
	logger.Info("Twitch live notification closed", "guild", row.GuildID, "login", row.TwitchLogin)
}

func isNotFound(err error) bool {
	var restErr *rest.Error
	if !errors.As(err, &restErr) || restErr.Response == nil {
		return false
	}
	return restErr.Response.StatusCode == http.StatusNotFound
}
