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

// UserAffiliatePaymentAccount stores an affiliate's encrypted payout account.
type UserAffiliatePaymentAccount struct{ ent.Schema }

func (UserAffiliatePaymentAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_affiliate_payment_accounts"}}
}

func (UserAffiliatePaymentAccount) Mixin() []ent.Mixin { return []ent.Mixin{mixins.TimeMixin{}} }

func (UserAffiliatePaymentAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("type").MaxLen(20),
		field.String("details_encrypted").Sensitive().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("masked_summary").MaxLen(255),
		field.Bool("is_default").Default(false),
	}
}

func (UserAffiliatePaymentAccount) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id"), index.Fields("user_id", "is_default")}
}
