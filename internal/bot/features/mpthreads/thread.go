package mpthreads

import (
	"context"
	"errors"
	"fmt"

	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

var errDisabled = errors.New("mpthreads: feature disabled")

// recovery is what to do after a post into a thread failed.
type recovery int

const (
	recoveryReport recovery = iota // unknown cause, surface the error
	recoveryRetry                  // thread fixed, posting again is worth it
	recoveryStop                   // deliberate state, stop without noise
)

// ensureThread returns the thread mirroring a user's DMs, creating one when the
// user has none. Callers must hold the user's lock (see lockUser) so concurrent
// DMs cannot spawn duplicate threads.
func ensureThread(ctx context.Context, client *bot.Client, user discord.User) (snowflake.ID, error) {
	if threadID, ok := threadForUser(user.ID); ok {
		return threadID, nil
	}

	cfg, ok := config()
	if !ok {
		return 0, errDisabled
	}

	threadID, err := createThread(client, cfg.channelID, user)
	if err != nil {
		return 0, err
	}
	rememberMapping(ctx, user.ID, threadID)
	logger.Info("MP threads: thread created", "user", user.ID.String(), "thread", threadID.String())
	return threadID, nil
}

// recoverThread reacts to a failed post instead of probing the thread state
// before every single DM: a deleted thread is forgotten so the next attempt
// recreates one, an archived thread is reopened.
func recoverThread(ctx context.Context, client *bot.Client, threadID snowflake.ID, cause error) recovery {
	switch {
	case isMissingChannel(cause):
		logger.Warn("MP threads: thread vanished while posting, recreating", "thread", threadID.String())
		forgetThread(ctx, threadID)
		return recoveryRetry

	case isArchivedThread(cause):
		// Only Archived is cleared. Locking a thread is a moderation decision the
		// bridge must not undo, so Discord rejecting this update ends the relay.
		if _, err := client.Rest.UpdateChannel(threadID, discord.GuildThreadUpdate{Archived: ptr(false)}); err != nil {
			logger.Warn("MP threads: reopening archived thread failed, message not relayed",
				"thread", threadID.String(), "error", err)
			return recoveryStop
		}
		return recoveryRetry

	case isLockedThread(cause):
		logger.Info("MP threads: thread locked by a moderator, message not relayed", "thread", threadID.String())
		return recoveryStop

	default:
		return recoveryReport
	}
}

// createThread posts the parent message in the MP channel and opens the public
// thread that will mirror the user's DMs.
func createThread(client *bot.Client, channelID snowflake.ID, user discord.User) (snowflake.ID, error) {
	parent, err := client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Content:         fmt.Sprintf("Messages privés avec %s", user.Mention()),
		AllowedMentions: noMentions(),
	})
	if err != nil {
		return 0, fmt.Errorf("creating thread parent message: %w", err)
	}

	thread, err := client.Rest.CreateThreadFromMessage(channelID, parent.ID, discord.ThreadCreateFromMessage{
		Name:                threadName(user),
		AutoArchiveDuration: discord.AutoArchiveDuration1w,
	})
	if err != nil {
		return 0, fmt.Errorf("creating thread: %w", err)
	}
	return thread.ID(), nil
}

// isMissingChannel reports whether an error means the channel/thread no longer
// exists. Permission errors are deliberately excluded: they are transient
// configuration problems, and treating them as "deleted" would make the bot
// create a new thread on every DM.
func isMissingChannel(err error) bool {
	return rest.IsJSONErrorCode(err, rest.JSONErrorCodeUnknownChannel)
}

func isArchivedThread(err error) bool {
	return rest.IsJSONErrorCode(err, rest.JSONErrorCodeOperationOnArchivedThread)
}

func isLockedThread(err error) bool {
	return rest.IsJSONErrorCode(err, rest.JSONErrorCodeThreadLocked)
}

// isPayloadTooLarge reports whether Discord refused the upload itself, which
// means the guild's real limit is below maxUploadBytes.
func isPayloadTooLarge(err error) bool {
	return rest.IsJSONErrorCode(err, rest.JSONErrorCodeRequestEntityTooLarge)
}
