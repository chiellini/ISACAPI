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

// ResearchGroupMember links an independently authenticated student to a
// research group and tracks the student's monthly sponsored quota.
type ResearchGroupMember struct {
	ent.Schema
}

func (ResearchGroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "research_group_members"}}
}

func (ResearchGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("research_group_id"),
		field.Int64("user_id"),
		field.String("status").MaxLen(20).Default(domain.ResearchGroupMemberStatusPending).
			Validate(func(v string) error {
				switch v {
				case domain.ResearchGroupMemberStatusPending, domain.ResearchGroupMemberStatusActive, domain.ResearchGroupMemberStatusPaused, domain.ResearchGroupMemberStatusRemoved:
					return nil
				default:
					return fmt.Errorf("invalid research group member status %q", v)
				}
			}),
		field.Float("monthly_limit_usd").Default(0).SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("monthly_usage_usd").Default(0).SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("monthly_reserved_usd").Default(0).SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Time("usage_window_start").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("invited_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("accepted_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("paused_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("removed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ResearchGroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("research_group", ResearchGroup.Type).Ref("members").Field("research_group_id").Required().Unique(),
		edge.From("user", User.Type).Ref("research_group_memberships").Field("user_id").Required().Unique(),
	}
}

func (ResearchGroupMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("research_group_id", "user_id").Unique().Annotations(entsql.IndexWhere("status IN ('pending', 'active', 'paused')")),
		index.Fields("user_id").Unique().Annotations(entsql.IndexWhere("status IN ('pending', 'active', 'paused')")),
		index.Fields("research_group_id", "status"),
	}
}
