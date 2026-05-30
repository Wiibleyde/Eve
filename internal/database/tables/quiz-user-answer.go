package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type QuizUserAnswer struct {
	ent.Schema
}

func (QuizUserAnswer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("user_id"),
		field.Time("answered_at").Default(time.Now).Immutable(),
	}
}

func (QuizUserAnswer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("question", QuizQuestion.Type).
			Ref("user_answers").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("answer", QuizAnswer.Type).
			Ref("user_answers").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
