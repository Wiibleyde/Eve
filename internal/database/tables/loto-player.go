package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LotoPlayer is an RP character taking part in a loto game.
//
// The name is the RP name and is intentionally case-sensitive: "John Doe" and
// "john doe" are two different participants, which the public embed warns about.
type LotoPlayer struct {
	ent.Schema
}

func (LotoPlayer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("game_id"),
		field.String("name").MaxLen(50),
		// last_play drives the per-player purchase cooldown.
		field.Time("last_play").Default(time.Now),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LotoPlayer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("game_id", "name").Unique(),
	}
}

func (LotoPlayer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", LotoGame.Type).
			Ref("players").
			Unique().
			Required().
			Field("game_id"),
		// Cascade / set-null live here: ent only honours entsql.OnDelete on the
		// assoc (edge.To) side of a relation.
		edge.To("tickets", LotoTicket.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("won_prizes", LotoPrize.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}
