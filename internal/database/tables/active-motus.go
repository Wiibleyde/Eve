package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MotusAttempt is a single guess made during a Motus game. It is stored inside
// the JSON "attempts" column of ActiveMotus so the whole board can be rebuilt
// after a restart.
type MotusAttempt struct {
	// Word is the normalized guess (uppercase, no diacritics).
	Word string `json:"word"`
	// UserID is the Discord snowflake of the player who submitted the guess.
	UserID string `json:"user_id"`
}

// Motus game states, stored as-is in the "state" column.
const (
	MotusStatePlaying = "playing"
	MotusStateWon     = "won"
	MotusStateLost    = "lost"
)

// ActiveMotus is a Motus game in progress, keyed by the board message it is
// attached to. Keeping games in the database instead of an in-memory map means
// a restart does not kill running games.
type ActiveMotus struct {
	ent.Schema
}

func (ActiveMotus) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("message_id").Unique().Immutable(),
		field.String("channel_id").Immutable(),
		field.String("guild_id").Immutable(),
		field.String("word").Immutable(),
		field.JSON("attempts", []MotusAttempt{}).Default([]MotusAttempt{}),
		field.String("state").Default(MotusStatePlaying),
		field.String("started_by").Immutable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("expires_at"),
	}
}

func (ActiveMotus) Edges() []ent.Edge {
	return nil
}
