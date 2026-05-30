package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type QuizAnswer struct {
	ent.Schema
}

func (QuizAnswer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("answer_text"),
		field.Bool("is_valid").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (QuizAnswer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("question", QuizQuestion.Type).
			Ref("answers").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user_answers", QuizUserAnswer.Type),
	}
}
