package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type MotusAttempt struct {
	Word   string `json:"word"`
	UserID string `json:"user_id"`
}

const (
	MotusStatePlaying = "playing"
	MotusStateWon     = "won"
	MotusStateLost    = "lost"
)

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
