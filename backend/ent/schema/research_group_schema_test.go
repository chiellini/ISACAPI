package schema

import (
	"testing"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/field"
	"github.com/stretchr/testify/require"
)

func TestResearchGroupSchemas(t *testing.T) {
	graph, err := entc.LoadGraph(".", &gen.Config{IDType: &field.TypeInfo{Type: field.TypeInt64}})
	require.NoError(t, err)

	schemas := make(map[string]*load.Schema, len(graph.Schemas))
	for _, item := range graph.Schemas {
		schemas[item.Name] = item
	}

	group := requireSchema(t, schemas, "ResearchGroup")
	requireSchemaFields(t, group, "name", "owner_user_id", "status", "dissolved_at")
	ownerID := requireSchemaField(t, group, "owner_user_id")
	require.True(t, ownerID.Optional)
	require.True(t, ownerID.Nillable)

	member := requireSchema(t, schemas, "ResearchGroupMember")
	requireSchemaFields(t, member,
		"research_group_id", "user_id", "status", "monthly_limit_usd",
		"monthly_usage_usd", "monthly_reserved_usd", "usage_window_start",
	)
	requireHasUniqueIndex(t, member, "user_id")

	audit := requireSchema(t, schemas, "ResearchGroupQuotaAudit")
	for _, name := range []string{
		"research_group_id", "member_id", "actor_user_id", "action",
		"amount_usd", "previous_value_usd", "new_value_usd", "metadata", "created_at",
	} {
		require.True(t, requireSchemaField(t, audit, name).Immutable, "audit field %s must be immutable", name)
	}
	requireNoSchemaEdge(t, audit, "member")
	requireNoSchemaEdge(t, audit, "actor")

	for _, schemaName := range []string{"UsageLog", "BatchImageJob"} {
		item := requireSchema(t, schemas, schemaName)
		for _, name := range []string{"payer_user_id", "research_group_id", "research_group_member_id", "funding_source"} {
			require.True(t, requireSchemaField(t, item, name).Immutable, "%s.%s must be immutable", schemaName, name)
		}
	}
	requireNoSchemaEdge(t, requireSchema(t, schemas, "UsageLog"), "payer")
	requireNoSchemaEdge(t, requireSchema(t, schemas, "UsageLog"), "research_group_member")
}

func requireNoSchemaEdge(t *testing.T, schema *load.Schema, name string) {
	t.Helper()
	for _, schemaEdge := range schema.Edges {
		require.NotEqual(t, name, schemaEdge.Name, "schema %s must not expose snapshot field %s as a live edge", schema.Name, name)
	}
}
