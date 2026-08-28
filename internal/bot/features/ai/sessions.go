package ai

import (
	"context"
	"sync"
	"time"

	"Eve/internal/gemini"

	"github.com/disgoorg/snowflake/v2"
)

const sessionTTL = 30 * time.Minute

type session struct {
	chat         *gemini.Chat
	participants map[snowflake.ID]struct{}
	updated      time.Time
}

type chatFactory func(ctx context.Context, systemInstruction string) (*gemini.Chat, error)

type sessionStore struct {
	mu       sync.Mutex
	sessions map[snowflake.ID]*session
	ttl      time.Duration
	newChat  chatFactory
}

var channelSessions = newSessionStore(sessionTTL, func(ctx context.Context, systemInstruction string) (*gemini.Chat, error) {
	return gemini.Default().NewChat(ctx, systemInstruction)
})

func newSessionStore(ttl time.Duration, newChat chatFactory) *sessionStore {
	return &sessionStore{
		sessions: make(map[snowflake.ID]*session),
		ttl:      ttl,
		newChat:  newChat,
	}
}

func (store *sessionStore) chatFor(ctx context.Context, channelID snowflake.ID, now time.Time, systemInstruction string) (*gemini.Chat, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.prune(now)
	if sess, ok := store.sessions[channelID]; ok {
		return sess.chat, nil
	}

	chat, err := store.newChat(ctx, systemInstruction)
	if err != nil {
		return nil, err
	}

	store.sessions[channelID] = &session{
		chat:         chat,
		participants: make(map[snowflake.ID]struct{}),
		updated:      now,
	}
	return chat, nil
}

func (store *sessionStore) touch(channelID snowflake.ID, now time.Time, speaker snowflake.ID) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.prune(now)
	sess, ok := store.sessions[channelID]
	if !ok {
		return
	}
	sess.participants[speaker] = struct{}{}
	sess.updated = now
}

func (store *sessionStore) participants(channelID snowflake.ID, now time.Time) []snowflake.ID {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.prune(now)
	sess, ok := store.sessions[channelID]
	if !ok {
		return nil
	}
	ids := make([]snowflake.ID, 0, len(sess.participants))
	for id := range sess.participants {
		ids = append(ids, id)
	}
	return ids
}

func (store *sessionStore) prune(now time.Time) {
	for channelID, sess := range store.sessions {
		if now.Sub(sess.updated) >= store.ttl {
			delete(store.sessions, channelID)
		}
	}
}
