package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Quote struct {
	ent.Schema
}

func (Quote) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("guild_id"),
		field.String("author_id"),
		field.String("quote"),
		field.String("context").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}
