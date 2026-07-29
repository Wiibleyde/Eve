package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LotoPrize is one ranked prize of a loto game.
//
// position is 0-based and unique per game; the winner columns stay NULL until
// the (one-shot, irreversible) draw fills them in.
type LotoPrize struct {
	ent.Schema
}

func (LotoPrize) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("game_id"),
		field.String("label").MaxLen(255),
		field.Int("position"),
		field.String("winner_player_id").Optional(),
		field.Int("winning_ticket_number").Optional(),
		field.Time("drawn_at").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LotoPrize) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("game_id", "position").Unique(),
	}
}

func (LotoPrize) Edges() []ent.Edge {
	return []ent.Edge{
		// The delete actions live on LotoGame.prizes / LotoPlayer.won_prizes:
		// ent only reads entsql.OnDelete from the assoc (edge.To) side.
		edge.From("game", LotoGame.Type).
			Ref("prizes").
			Unique().
			Required().
			Field("game_id"),
		edge.From("winner", LotoPlayer.Type).
			Ref("won_prizes").
			Unique().
			Field("winner_player_id"),
	}
}
