package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LotoGame struct {
	ent.Schema
}

func (LotoGame) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable(),
		field.String("guild_id"),
		field.String("name").MaxLen(50),
		field.Bool("active").Default(true),
		field.Int("ticket_price").Default(500),
		field.Int("cooldown_minutes").Default(0),
		field.Int("max_tickets_per_purchase").Optional(),
		field.String("message_id").Optional(),
		field.String("channel_id").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LotoGame) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guild_id", "active"),
		index.Fields("guild_id").
			Unique().
			StorageKey("lotogame_one_active_per_guild").
			Annotations(entsql.IndexWhere("active")),
	}
}

func (LotoGame) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("players", LotoPlayer.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tickets", LotoTicket.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("prizes", LotoPrize.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
