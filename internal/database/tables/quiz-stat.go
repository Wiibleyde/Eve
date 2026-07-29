package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type QuizStat struct {
	ent.Schema
}

func (QuizStat) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("user_id").Unique(),
		field.Int("good_answers").Default(0),
		field.Int("bad_answers").Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (QuizStat) Edges() []ent.Edge {
	return nil
}
