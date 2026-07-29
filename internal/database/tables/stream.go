package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Stream struct {
	ent.Schema
}

func (Stream) Fields() []ent.Field {
	return []ent.Field{
		field.String("guild_id"),
		field.String("channel_id"),
		field.String("role_id").Optional(),
		field.String("twitch_user_id"),
		field.String("twitch_login"),
		field.String("message_id").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Stream) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guild_id", "twitch_user_id").Unique(),
	}
}

func (Stream) Edges() []ent.Edge {
	return nil
}
