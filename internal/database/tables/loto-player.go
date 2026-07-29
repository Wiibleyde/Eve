package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LotoPlayer struct {
	ent.Schema
}

func (LotoPlayer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("game_id"),
		field.String("name").MaxLen(50),
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
		edge.To("tickets", LotoTicket.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("won_prizes", LotoPrize.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}
