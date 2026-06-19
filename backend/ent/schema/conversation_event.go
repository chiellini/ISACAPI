package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ConversationEvent 表示某个分支上的一个事件（对话存档模块）。
//
// 事件只追加、不覆盖。内容默认 AES-256-GCM 加密后存 content_ciphertext，明文绝不落库；
// content_preview 仅存脱敏短预览。event_hash 为规范化内容指纹，用于分支去重；
// upstream_response_id_hash / tool_call_id_hash 保存上游稳定 id 的哈希用于追踪与父引用解析，
// 但绝不用于把这些 id 重新注入请求。
type ConversationEvent struct {
	ent.Schema
}

func (ConversationEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "conversation_events"},
	}
}

func (ConversationEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("session_id", uuid.UUID{}),
		field.UUID("branch_id", uuid.UUID{}),
		field.Int64("parent_event_id").
			Optional().
			Nillable(),
		field.Int("sequence").
			Comment("分支内单调递增序号，配合 request_id 幂等"),
		field.String("request_id").
			MaxLen(64),

		field.String("role").
			MaxLen(16).
			Comment("system / user / assistant / tool"),
		field.String("kind").
			MaxLen(32).
			Comment("message / tool_call / tool_result / image_reference / file_reference / error"),

		field.Bytes("content_ciphertext").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "bytea"}),
		field.Bytes("content_nonce").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "bytea"}),
		field.Int("encryption_key_version").
			Default(0),
		field.String("content_preview").
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("event_hash").
			MaxLen(64).
			Optional(),

		field.String("model").
			MaxLen(100).
			Optional(),
		field.String("provider").
			MaxLen(32).
			Optional(),
		field.String("upstream_response_id_hash").
			MaxLen(64).
			Optional(),
		field.String("tool_call_id_hash").
			MaxLen(64).
			Optional(),

		field.Bool("partial").
			Default(false).
			Comment("客户端中途断开导致内容不完整时为 true"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ConversationEvent) Indexes() []ent.Index {
	return []ent.Index{
		// 幂等：同一分支同一请求同一序号只写一次。
		index.Fields("branch_id", "request_id", "sequence").Unique(),
		index.Fields("session_id"),
		index.Fields("branch_id", "sequence"),
		index.Fields("request_id"),
		index.Fields("event_hash"),
		index.Fields("upstream_response_id_hash"),
	}
}
