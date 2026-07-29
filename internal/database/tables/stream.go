package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Stream is a Twitch channel tracked by a guild.
//
// twitch_user_id is the tracking key, never the login: Twitch logins can be
// renamed and the old one is then handed to somebody else. twitch_login is kept
// only as a display/lookup cache and is refreshed on every /streamer add.
//
// message_id holds the live notification currently posted for this row. Empty
// means "no live notification on screen": the poller uses it as the started /
// ended state flag, which is what makes a restart mid-stream resume the
// existing message instead of posting a duplicate.
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
