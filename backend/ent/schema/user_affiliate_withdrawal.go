package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserAffiliateWithdrawal stores the payout destination snapshot and lifecycle.
type UserAffiliateWithdrawal struct{ ent.Schema }

func (UserAffiliateWithdrawal) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_affiliate_withdrawals"}}
}

func (UserAffiliateWithdrawal) Mixin() []ent.Mixin { return []ent.Mixin{mixins.TimeMixin{}} }

func (UserAffiliateWithdrawal) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Float("amount").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("status").MaxLen(20).Default("submitted"),
		field.Int64("payment_account_id").Optional().Nillable(),
		field.String("payment_account_type").MaxLen(20),
		field.String("payment_details_encrypted").Sensitive().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("payment_account_summary").MaxLen(255),
		field.String("idempotency_key").Optional().Nillable().Sensitive().MaxLen(128),
		field.Time("submitted_at"),
		field.Time("reviewed_at").Optional().Nillable(),
		field.Int64("reviewed_by").Optional().Nillable(),
		field.Time("paid_at").Optional().Nillable(),
		field.Int64("paid_by").Optional().Nillable(),
		field.Time("canceled_at").Optional().Nillable(),
		field.String("cancel_reason").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("reject_reason").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("actual_currency").Optional().Nillable().MaxLen(20),
		field.Float("actual_amount").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("exchange_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("external_reference").Optional().Nillable().MaxLen(255),
	}
}

func (UserAffiliateWithdrawal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("status", "created_at"),
		index.Fields("external_reference"),
	}
}
