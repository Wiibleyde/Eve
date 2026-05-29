package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Birthday struct {
	ent.Schema
}

func (Birthday) Fields() []ent.Field {
	return []ent.Field{
		field.String("discord_id").Unique().Immutable(),
		field.Time("birthday"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Birthday) Edges() []ent.Edge {
	return nil
}
