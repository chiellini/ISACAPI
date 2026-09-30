-- 内置聊天的用户自定义助手（人设 + 默认模型 + 开场白），按用户隔离。
-- 助手只保存用户偏好；模型可用性仍在对话时由聊天策略中间件强制校验，
-- 因此这里对 model 不做外键/枚举约束。

CREATE TABLE IF NOT EXISTS chat_assistants (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL,
    name            VARCHAR(100) NOT NULL,
    description     VARCHAR(500) NOT NULL DEFAULT '',
    system_prompt   TEXT         NOT NULL DEFAULT '',
    model           VARCHAR(100) NOT NULL DEFAULT '',
    opening_message TEXT         NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_assistants_user_updated
    ON chat_assistants (user_id, updated_at DESC);

-- 会话可绑定一个助手（创建时指定）。助手被删除后置空，会话与消息保留。
ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS assistant_id BIGINT REFERENCES chat_assistants(id) ON DELETE SET NULL;
