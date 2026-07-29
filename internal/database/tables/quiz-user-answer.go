package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type QuizUserAnswer struct {
	ent.Schema
}

func (QuizUserAnswer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("active_quiz_id"),
		field.String("user_id"),
		field.Bool("correct").Default(false),
		field.Time("answered_at").Default(time.Now).Immutable(),
	}
}

func (QuizUserAnswer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("active_quiz_id", "user_id").Unique(),
	}
}

func (QuizUserAnswer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("active_quiz", ActiveQuiz.Type).
			Ref("user_answers").
			Field("active_quiz_id").
			Unique().
			Required(),
	}
}
