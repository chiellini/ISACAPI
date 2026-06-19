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

// ConversationBranch 表示会话内的一条分支（对话存档模块）。
//
// 用户编辑旧消息后重发会产生分叉：原分支仍是有效历史，应新建分支而非覆盖旧事件：
//
//	A → B → C
//	     └→ B' → D
//
// 每个活跃分支维护游标 (tail_sequence, tail_event_hash, head_event_id)，使正常追加
// 走 O(新增事件) 快速路径；游标失配时才做最长公共前缀比较并建新分支。
type ConversationBranch struct {
	ent.Schema
}

func (ConversationBranch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "conversation_branches"},
	}
}

func (ConversationBranch) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("session_id", uuid.UUID{}),
		field.UUID("parent_branch_id", uuid.UUID{}).
			Optional().
			Nillable(),
		// 分叉点：父分支中发生分叉的事件 id。
		field.Int64("fork_event_id").
			Optional().
			Nillable(),
		// 分支头部（最新）事件 id，即游标 tail_event_id。
		field.Int64("head_event_id").
			Optional().
			Nillable(),

		field.Int("event_count").
			Default(0),
		// 游标快速路径用：尾事件序号与其内容哈希。
		field.Int("tail_sequence").
			Default(-1),
		field.String("tail_event_hash").
			MaxLen(64).
			Optional(),

		field.String("branch_reason").
			MaxLen(32).
			Default("initial").
			Comment("initial / edited_history / continuation_mismatch / provider_boundary / manual_fork / unknown_parent"),
		field.String("status").
			MaxLen(32).
			Default("active"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_active_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ConversationBranch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("session_id", "status"),
		index.Fields("parent_branch_id"),
	}
}
