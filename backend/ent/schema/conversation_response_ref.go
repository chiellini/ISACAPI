package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/google/uuid"
)

// ConversationResponseRef 记录上游 response/item id → 会话分支尾部 的映射（对话存档模块）。
//
// 用于解析客户端下一轮带回的 previous_response_id：查到映射即可定位 (session, branch, tail)。
// 只存 id 的哈希，且仅用于归档侧的父引用解析与追踪——绝不据此向请求重新注入 id。
//
// durable 标记该上游 id 是否可持久复用：openai_api_key 的 response_id 通常 durable=true；
// openai_oauth_codex 的 reasoning item 在 store:false 下 durable=false。
type ConversationResponseRef struct {
	ent.Schema
}

func (ConversationResponseRef) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "conversation_response_refs"},
	}
}

func (ConversationResponseRef) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (ConversationResponseRef) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("context_domain").
			MaxLen(40),
		field.String("response_id_hash").
			MaxLen(64),

		field.UUID("session_id", uuid.UUID{}),
		field.UUID("branch_id", uuid.UUID{}),
		field.Int64("tail_event_id").
			Optional().
			Nillable(),

		field.Bool("durable").
			Default(false),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ConversationResponseRef) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "context_domain", "response_id_hash").Unique(),
		index.Fields("session_id"),
		index.Fields("branch_id"),
		index.Fields("expires_at"),
	}
}
