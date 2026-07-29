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
	"Eve/internal/database/ent"
	"Eve/internal/logger"
	"Eve/internal/twitch"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

const (
	// PollInterval is how often the live set is refreshed. The TS version used
	// a 13s cron, which hammered Helix for no benefit: a notification landing
	// within a minute of going live is perfectly fine.
	PollInterval = 60 * time.Second

	// ViewerEditInterval is the minimum delay between two edits triggered by
	// the viewer count alone. Title and game changes are not throttled — those
	// are real events; a viewer count drifting by three is not worth an edit.
	ViewerEditInterval = 5 * time.Minute

	// tickTimeout bounds one whole poll cycle, staying under PollInterval so a
	// stuck cycle cannot pile up on the next one.
	tickTimeout = 45 * time.Second

	// userCacheTTL is how long a Twitch profile (avatar, offline banner) is
	// reused. Profiles barely ever change and this keeps /users calls rare.
	userCacheTTL = 6 * time.Hour

	// offlineTolerance is how many consecutive ticks must miss a stream before
	// it is declared over. Helix is eventually consistent and intermittently
	// omits a running stream for one poll; ending on the first miss would post a
	// brand new notification (and a second role ping) on the very next tick.
	offlineTolerance = 2
)

// trackState is the in-memory view of one tracked row. The posted message ID is
// deliberately NOT kept here: the database row is the single source of truth,
// so a channel change or a manual edit cannot desync the poller.
type trackState struct {
	// seen reports whether the fields below describe a stream we actually
	// observed. False right after a restart, when only the message ID is known.
	seen      bool
	streamID  string
	title     string
	game      string
	viewers   int
	startedAt time.Time
	lastEdit  time.Time
	// missed counts consecutive ticks where Helix did not return the stream.
	missed int
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

// pollerOnce keeps a single poller alive for the whole process: StartPoller is
// called from the Ready listener, and Discord re-dispatches Ready on every
// non-resumable reconnect. A second poller would start with an empty state map
// and race the first one into duplicate notifications and duplicate role pings.
var pollerOnce sync.Once

// StartPoller launches the Twitch polling goroutine.
//
// The first tick runs immediately and doubles as the startup reconciliation:
// with an empty state map, a channel that is already live resumes its stored
// message instead of posting a duplicate notification.
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

	// Forget rows removed by /streamer remove so the map cannot grow forever.
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

	// A different stream ID on a row that still has a live message means the
	// previous stream ended between two ticks and a new one already started.
	// Close the old notification before opening a new one.
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
	content, mentions := roleMention(row.RoleID)

	msg, err := p.client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Content:         content,
		Embeds:          []discord.Embed{liveEmbed(s, user, hasUser)},
		AllowedMentions: mentions,
	})
	if err != nil {
		logger.Error("Sending Twitch live notification",
			"guild", row.GuildID, "channel", row.ChannelID, "login", row.TwitchLogin, "error", err)
		return
	}

	messageID := msg.ID.String()
	if err := entClient().Stream.UpdateOneID(row.ID).SetMessageID(messageID).Exec(ctx); err != nil {
		// The message is out but unrecorded: without the ID the next tick would
		// post a duplicate, so delete it and let the next tick start over.
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
		// Nothing worth an edit; keep the stream ID in sync all the same.
		st.streamID = s.ID
		return
	}

	user, hasUser := p.userInfo(ctx, row.TwitchUserID)
	if !p.editMessage(ctx, row, liveEmbed(s, user, hasUser)) {
		return
	}

	st.observe(s)
	st.lastEdit = time.Now()
	logger.Debug("Twitch live notification updated",
		"guild", row.GuildID, "login", s.UserLogin, "resume", resume, "viewers", s.ViewerCount)
}

// shouldEdit decides whether the live embed is worth rewriting.
//
//   - a state with no observation yet means the bot restarted mid-stream: edit
//     once to pick up whatever changed while it was down;
//   - a new title or a new game is a real event, edited immediately;
//   - a viewer count drifting on its own is cosmetic and is only allowed to
//     spend an edit every ViewerEditInterval.
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
	p.editMessage(ctx, row, endedEmbed(st, row.TwitchLogin, user, hasUser))

	// Clear the reference even when the edit failed: the stream is over either
	// way, and keeping the ID would block the next notification.
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

// editMessage rewrites the notification embed. It reports whether the edit
// succeeded. A message that no longer exists clears the stored ID so the next
// tick can post a fresh notification instead of retrying forever.
func (p *poller) editMessage(ctx context.Context, row *ent.Stream, embed discord.Embed) bool {
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

	// The role was already pinged by the original message; an edit must never
	// ping again.
	_, err = p.client.Rest.UpdateMessage(channelID, messageID, discord.NewMessageUpdate().
		WithEmbeds(embed).
		WithAllowedMentions(noMentions()),
	)
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
		// Reset in place rather than dropping the entry: the caller still holds
		// this pointer and would otherwise write to a state nobody can read back.
		if st, ok := p.states[row.ID]; ok {
			st.reset()
		}
		return false
	}

	logger.Error("Editing Twitch live notification",
		"guild", row.GuildID, "login", row.TwitchLogin, "error", err)
	return false
}

// userInfo returns the cached Twitch profile of a user, refreshing it when the
// entry expired. A failure is not fatal: the embed just loses its avatar.
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

// roleMention builds the notification content and the matching allowed
// mentions. Mentions are always restricted to the configured role: an embed
// title is user-controlled text and must never be able to ping anyone.
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

// noMentions is the "ping nobody" policy. Parse is an explicit empty array, not
// the zero value: a nil Parse marshals as JSON null, which Discord is free to
// read as "default parsing" rather than "no mention types".
func noMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{Parse: []discord.AllowedMentionType{}}
}

// closeNotifications marks the live notifications of rows the poller will never
// see again — deleted by /streamer remove, or moved to another channel — as
// «Stream terminé». Nothing else would: the poller only knows about rows it can
// still read.
//
// The work is detached because the interaction it comes from has already been
// answered and must not wait on these REST calls.
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

	_, err = client.Rest.UpdateMessage(channelID, messageID, discord.NewMessageUpdate().
		WithEmbeds(endedEmbed(nil, row.TwitchLogin, twitch.User{}, false)).
		WithAllowedMentions(noMentions()),
	)
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
