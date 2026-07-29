package reactions

import (
	"sync"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

// cooldownWindow is the minimum delay between two joke replies in the same
// channel. The TypeScript version had no rate limit at all, which made a
// two-bot or copy-paste loop trivially spammable.
const cooldownWindow = 30 * time.Second

// cooldowns is the process-wide channel cooldown. In-memory only: losing it on
// restart just means one extra joke, nothing worth a table.
var cooldowns = newCooldownMap(cooldownWindow)

// cooldownMap records the last reply time per channel. It is safe for
// concurrent use — gateway events may be dispatched from several goroutines.
type cooldownMap struct {
	mu     sync.Mutex
	last   map[snowflake.ID]time.Time
	window time.Duration
}

func newCooldownMap(window time.Duration) *cooldownMap {
	return &cooldownMap{
		last:   make(map[snowflake.ID]time.Time),
		window: window,
	}
}

// allow reports whether a reply is permitted in the channel at now, and marks
// the channel as used when it is. The check and the mark are one atomic step so
// two simultaneous messages can never both get through.
//
// It is called before the opt-out lookup: reserving the slot early costs a
// muted guild nothing (it never replies anyway) and removes the race.
func (c *cooldownMap) allow(channelID snowflake.ID, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if last, ok := c.last[channelID]; ok && now.Sub(last) < c.window {
		return false
	}
	c.prune(now)
	c.last[channelID] = now
	return true
}

// prune drops expired entries so the map stays bounded by the number of
// recently active channels. Called under lock, and only on the allowed path,
// which is itself rate limited.
func (c *cooldownMap) prune(now time.Time) {
	for channelID, last := range c.last {
		if now.Sub(last) >= c.window {
			delete(c.last, channelID)
		}
	}
}
