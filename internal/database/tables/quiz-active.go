package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ActiveQuiz is a question currently posted in a channel, waiting for answers.
//
// The TS version kept running quizzes in an in-memory map, so a restart left
// dead buttons behind. Persisting them means the answer handler can always
// resolve a message back to its question and its answer order.
//
// `shuffle` is the permutation applied to the answers when the message was
// built, stored as "2,0,3,1": position i in the message holds the answer at
// index shuffle[i] of [good, bad1, bad2, bad3].
type ActiveQuiz struct {
	ent.Schema
}

func (ActiveQuiz) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		// Foreign key of the "question" edge, exposed as a plain column so the
		// answer handler can read it without an extra query.
		field.String("question_id"),
		// One running quiz per message; this is the lookup key of the answer
		// buttons.
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
		// Deleting a finished quiz drops the answers recorded on it (the
		// per-user totals live in quiz_stats and are not touched).
		edge.To("user_answers", QuizUserAnswer.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
