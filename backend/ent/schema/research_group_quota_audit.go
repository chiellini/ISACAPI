package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ResearchGroupQuotaAudit is an append-only record of membership and quota operations.
type ResearchGroupQuotaAudit struct{ ent.Schema }

func (ResearchGroupQuotaAudit) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "research_group_quota_audits"}}
}

func (ResearchGroupQuotaAudit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("research_group_id").Immutable(),
		field.Int64("member_id").Optional().Nillable().Immutable(),
		field.Int64("actor_user_id").Optional().Nillable().Immutable(),
		field.String("action").MaxLen(50).NotEmpty().Immutable(),
		field.Float("amount_usd").Optional().Nillable().Immutable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("previous_value_usd").Optional().Nillable().Immutable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("new_value_usd").Optional().Nillable().Immutable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.JSON("metadata", map[string]any{}).Optional().Immutable().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ResearchGroupQuotaAudit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("research_group", ResearchGroup.Type).Ref("quota_audits").Field("research_group_id").Required().Unique().Immutable(),
	}
}

func (ResearchGroupQuotaAudit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("research_group_id", "created_at"),
		index.Fields("member_id", "created_at"),
		index.Fields("actor_user_id", "created_at"),
	}
}
