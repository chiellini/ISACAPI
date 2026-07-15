package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserAffiliateAdminAudit is an append-only audit trail for affiliate writes.
type UserAffiliateAdminAudit struct{ ent.Schema }

func (UserAffiliateAdminAudit) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_affiliate_admin_audits"}}
}

func (UserAffiliateAdminAudit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("operator_user_id"),
		field.Int64("target_user_id").Optional().Nillable(),
		field.Int64("withdrawal_id").Optional().Nillable(),
		field.String("action").MaxLen(64),
		field.String("idempotency_key").Optional().Nillable().MaxLen(128),
		field.String("detail").Default("{}").SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (UserAffiliateAdminAudit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("target_user_id", "created_at"),
		index.Fields("withdrawal_id", "created_at"),
		index.Fields("operator_user_id", "created_at"),
	}
}
