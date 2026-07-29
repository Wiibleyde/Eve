package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LotoTicket is a single ticket sold for a loto game.
//
// number is sequential per game and starts at 1. It is an explicit column (the
// TS version derived it from row order, which broke as soon as tickets were
// removed) and is what the draw records as the winning ticket number.
type LotoTicket struct {
	ent.Schema
}

func (LotoTicket) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("game_id"),
		field.String("player_id"),
		// Discord snowflake of the member who sold the ticket.
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
		// The cascade annotations live on LotoGame.tickets / LotoPlayer.tickets:
		// ent only reads entsql.OnDelete from the assoc (edge.To) side.
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
