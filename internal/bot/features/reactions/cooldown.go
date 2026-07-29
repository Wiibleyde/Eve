package reactions

import (
	"sync"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

const cooldownWindow = 30 * time.Second

var cooldowns = newCooldownMap(cooldownWindow)

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

func (c *cooldownMap) prune(now time.Time) {
	for channelID, last := range c.last {
		if now.Sub(last) >= c.window {
			delete(c.last, channelID)
		}
	}
}
