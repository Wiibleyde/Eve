package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type QuizQuestion struct {
	ent.Schema
}

func (QuizQuestion) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("question").MaxRuneLen(2048).Unique(),
		field.String("good_answer").MaxRuneLen(256),
		field.String("bad_answer_1").MaxRuneLen(256),
		field.String("bad_answer_2").MaxRuneLen(256),
		field.String("bad_answer_3").MaxRuneLen(256),
		field.String("category").MaxRuneLen(128),
		field.String("difficulty").MaxRuneLen(32),
		field.String("author_id").Optional(),
		field.String("guild_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("last_used_at").Optional(),
	}
}

func (QuizQuestion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("active_quizzes", ActiveQuiz.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
