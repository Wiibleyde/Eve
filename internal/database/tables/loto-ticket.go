package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LotoTicket struct {
	ent.Schema
}

func (LotoTicket) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("game_id"),
		field.String("player_id"),
		field.String("seller_id"),
		field.Int("number"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (LotoTicket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("game_id", "number").Unique(),
		index.Fields("player_id"),
	}
}

func (LotoTicket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", LotoGame.Type).
			Ref("tickets").
			Unique().
			Required().
			Field("game_id"),
		edge.From("player", LotoPlayer.Type).
			Ref("tickets").
			Unique().
			Required().
			Field("player_id"),
	}
}
