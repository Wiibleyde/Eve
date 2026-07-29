package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type MPThread struct {
	ent.Schema
}

func (MPThread) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Unique().Immutable(),
		field.String("thread_id").Unique(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (MPThread) Edges() []ent.Edge {
	return nil
}
