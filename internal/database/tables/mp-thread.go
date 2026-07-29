package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MPThread maps a Discord user to the staff thread mirroring their DMs with the
// bot. Both sides are unique: one user has at most one thread, one thread
// belongs to exactly one user. Rows are dropped when the thread is deleted on
// Discord, and recreated on the user's next DM.
type MPThread struct {
	ent.Schema
}

func (MPThread) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Unique().Immutable(),
		field.String("thread_id").Unique(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (MPThread) Edges() []ent.Edge {
	return nil
}
