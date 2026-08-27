package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GuildConfig struct {
	ent.Schema
}

type ConfigKey int

const (
	BirthdayChannel ConfigKey = iota
	QuoteChannel
	CalendarURL
	CalendarChannel
	CalendarMessage
	JokesDisabled
	DebugRole
	AIDisabled
)

var configKey = map[ConfigKey]string{
	BirthdayChannel: "birthday.channel",
	QuoteChannel:    "quote.channel",
	CalendarURL:     "calendar.url",
	CalendarChannel: "calendar.channel",
	CalendarMessage: "calendar.message",
	JokesDisabled:   "jokes.disabled",
	DebugRole:       "debug.role",
	AIDisabled:      "ai.disabled",
}

func (k ConfigKey) String() string {
	return configKey[k]
}

func (GuildConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("guild_id"),
		field.String("key"),
		field.String("value"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (GuildConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guild_id", "key").Unique(),
	}
}

func (GuildConfig) Edges() []ent.Edge {
	return nil
}
