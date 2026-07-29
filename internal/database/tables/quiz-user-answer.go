package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// QuizUserAnswer records that a user answered a running quiz.
//
// One answer per user per quiz is enforced by the unique index below, not by
// bookkeeping in memory: the row insert is the lock.
type QuizUserAnswer struct {
	ent.Schema
}

func (QuizUserAnswer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		// Foreign key of the "active_quiz" edge, exposed as a plain column.
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
