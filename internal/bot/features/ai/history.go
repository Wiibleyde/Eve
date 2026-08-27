package ai

import (
	"sync"
	"time"

	"Eve/internal/ollama"

	"github.com/disgoorg/snowflake/v2"
)

const (
	maxStoredMessages = 12
	conversationTTL   = 30 * time.Minute
)

type conversation struct {
	messages     []ollama.Message
	participants map[snowflake.ID]struct{}
	updated      time.Time
}

type history struct {
	mu            sync.Mutex
	conversations map[snowflake.ID]*conversation
	ttl           time.Duration
	limit         int
}

var channelHistory = newHistory(conversationTTL, maxStoredMessages)

func newHistory(ttl time.Duration, limit int) *history {
	return &history{
		conversations: make(map[snowflake.ID]*conversation),
		ttl:           ttl,
		limit:         limit,
	}
}

func (history *history) messages(channelID snowflake.ID, now time.Time) []ollama.Message {
	history.mu.Lock()
	defer history.mu.Unlock()

	history.prune(now)
	conv, ok := history.conversations[channelID]
	if !ok {
		return nil
	}
	out := make([]ollama.Message, len(conv.messages))
	copy(out, conv.messages)
	return out
}

func (history *history) participants(channelID snowflake.ID, now time.Time) []snowflake.ID {
	history.mu.Lock()
	defer history.mu.Unlock()

	history.prune(now)
	conv, ok := history.conversations[channelID]
	if !ok {
		return nil
	}
	ids := make([]snowflake.ID, 0, len(conv.participants))
	for id := range conv.participants {
		ids = append(ids, id)
	}
	return ids
}

func (history *history) append(channelID snowflake.ID, now time.Time, speaker snowflake.ID, messages ...ollama.Message) {
	history.mu.Lock()
	defer history.mu.Unlock()

	history.prune(now)
	conv, ok := history.conversations[channelID]
	if !ok {
		conv = &conversation{participants: make(map[snowflake.ID]struct{})}
		history.conversations[channelID] = conv
	}
	conv.participants[speaker] = struct{}{}
	conv.messages = append(conv.messages, messages...)
	if len(conv.messages) > history.limit {
		conv.messages = append([]ollama.Message(nil), conv.messages[len(conv.messages)-history.limit:]...)
	}
	conv.updated = now
}

func (history *history) prune(now time.Time) {
	for channelID, conv := range history.conversations {
		if now.Sub(conv.updated) >= history.ttl {
			delete(history.conversations, channelID)
		}
	}
}
