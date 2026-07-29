package tables

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LotoGame is one RP lottery run inside a guild.
//
// Only one game per guild may be active at a time, enforced at the DB level by
// a partial unique index on guild_id where active (Postgres-specific
// entsql.IndexWhere).
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
		// 0 / NULL means "no per-purchase limit".
		field.Int("max_tickets_per_purchase").Optional(),
		// Public message holding the embed + buttons, set right after creation.
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

// Edges owns the cascade annotations on purpose: ent resolves entsql.OnDelete
// from the assoc (edge.To) side only — the same annotation on the inverse
// edge.From side is silently ignored.
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
