package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Quiz struct {
	ent.Schema
}

func (Quiz) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("author_id"),
		field.String("title"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Quiz) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("questions", QuizQuestion.Type),
	}
}
