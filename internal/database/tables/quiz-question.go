package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// QuizQuestion is one community quiz question with its four answers stored
// inline.
//
// The previous design normalised answers into a `quiz_answer` table, which
// bought nothing: a question always has exactly one good answer and three bad
// ones. Inlining them removes a join and makes the shuffle permutation stored
// on ActiveQuiz meaningful (index 0 is always the good answer).
type QuizQuestion struct {
	ent.Schema
}

// Lengths are MaxRuneLen, not MaxLen: MaxLen validates a byte count while the
// varchar(n) column it generates and the Discord option limits both count
// characters, so a multi-byte input would pass every check and still be
// rejected by ent at insert time.
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
		// Empty when the author is unknown (imported question, deleted user).
		field.String("author_id").Optional(),
		// Where the question was added. The pool is global — questions are
		// unique across guilds and quiz_stat has no guild_id — so this is
		// provenance only and is never used to filter the pick.
		field.String("guild_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
		// NULL until the question is launched once; drives the LRU pick.
		field.Time("last_used_at").Optional(),
	}
}

func (QuizQuestion) Edges() []ent.Edge {
	return []ent.Edge{
		// The ON DELETE annotation is only honoured on the edge.To side by ent
		// codegen: deleting a question drops the quizzes that used it.
		edge.To("active_quizzes", ActiveQuiz.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
