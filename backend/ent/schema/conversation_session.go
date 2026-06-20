package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/google/uuid"
)

// ConversationSession 表示一个逻辑会话（对话存档模块）。
//
// 数据模型为 Session → Branch → Event：一个会话可有多个分支（用户编辑旧消息后分叉）。
// 会话身份由共享的 ConversationIdentityResolver 给出，archive_key 为强信号派生的稳定键，
// 低置信度/无信号时为临时键（不与既有会话合并）。
//
// 唯一约束含 context_domain，保证官方 Anthropic 与 Antigravity Claude 等不同上游
// 即便客户端给出相同 session_id 也不会归并到同一会话。
type ConversationSession struct {
	ent.Schema
}

func (ConversationSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "conversation_sessions"},
	}
}

func (ConversationSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (ConversationSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),

		field.Int64("user_id"),
		field.Int64("api_key_id").
			Optional().
			Nillable().
			Comment("可空且不进唯一约束：换 Key 不应拆分会话"),
		// group_id 进唯一约束，故用 NOT NULL，0 表示无分组（避免 NULL 破坏唯一性）。
		field.Int64("group_id").
			Default(0),

		// archive_key：强信号派生的稳定键或临时键，配合 context_domain 唯一。
		field.String("archive_key").
			MaxLen(80).
			NotEmpty(),
		// context_domain：上游隔离域，如 anthropic_native / antigravity_claude / openai_oauth_codex。
		field.String("context_domain").
			MaxLen(40),
		field.String("protocol").
			MaxLen(32),

		field.String("title").
			MaxLen(255).
			Optional(),

		field.Time("started_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_active_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// 当前活跃分支；首个分支创建后回填。
		field.UUID("active_branch_id", uuid.UUID{}).
			Optional().
			Nillable(),

		field.Int("request_count").
			Default(0),
		field.Int64("total_input_tokens").
			Default(0),
		field.Int64("total_output_tokens").
			Default(0),

		field.String("status").
			MaxLen(32).
			Default("active").
			Comment("active / archived / expired"),
	}
}

func (ConversationSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "group_id", "context_domain", "archive_key").Unique(),
		index.Fields("user_id", "last_active_at"),
		index.Fields("group_id"),
		index.Fields("status"),
		index.Fields("context_domain"),
	}
}
