package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type ActiveQuiz struct {
	ent.Schema
}

func (ActiveQuiz) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("question_id"),
		field.String("message_id").Unique(),
		field.String("channel_id"),
		field.String("guild_id"),
		field.String("shuffle").MaxLen(16),
		field.Time("launched_at").Default(time.Now).Immutable(),
		field.Time("expires_at"),
	}
}

func (ActiveQuiz) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("question", QuizQuestion.Type).
			Ref("active_quizzes").
			Field("question_id").
			Unique().
			Required(),
		edge.To("user_answers", QuizUserAnswer.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
