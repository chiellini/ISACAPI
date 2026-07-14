package schema

import (
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ResearchGroup is a funding group owned by one user. It intentionally does
// not own API keys: students keep their own identity and keys.
type ResearchGroup struct {
	ent.Schema
}

func (ResearchGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "research_groups"}}
}

func (ResearchGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(100).NotEmpty(),
		field.Int64("owner_user_id").Optional().Nillable(),
		field.String("status").MaxLen(20).Default(domain.ResearchGroupStatusActive).
			Validate(func(v string) error {
				switch v {
				case domain.ResearchGroupStatusActive, domain.ResearchGroupStatusPaused, domain.ResearchGroupStatusDissolved:
					return nil
				default:
					return fmt.Errorf("invalid research group status %q", v)
				}
			}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("dissolved_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ResearchGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("owned_research_groups").Field("owner_user_id").Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("members", ResearchGroupMember.Type),
		edge.To("quota_audits", ResearchGroupQuotaAudit.Type),
		edge.To("funded_usage_logs", UsageLog.Type),
	}
}

func (ResearchGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_user_id").Unique().Annotations(entsql.IndexWhere("status <> 'dissolved'")),
		index.Fields("status"),
	}
}
