-- 对话存档模块（Conversation Archive）
-- 数据模型：Session → Branch → Event，外加 response_id 映射表。
-- 列的 NULL 语义须与 ent schema 对齐（Optional 非 Nillable -> NOT NULL DEFAULT；Nillable -> 可空）。
-- id 由应用层 (uuid.New) 生成，不依赖数据库 uuid 扩展默认值。

CREATE TABLE IF NOT EXISTS conversation_sessions (
    id                  UUID PRIMARY KEY,
    user_id             BIGINT NOT NULL,
    api_key_id          BIGINT,
    group_id            BIGINT NOT NULL DEFAULT 0,
    archive_key         VARCHAR(80) NOT NULL,
    context_domain      VARCHAR(40) NOT NULL,
    protocol            VARCHAR(32) NOT NULL,
    title               VARCHAR(255) NOT NULL DEFAULT '',
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active_branch_id    UUID,
    request_count       INTEGER NOT NULL DEFAULT 0,
    total_input_tokens  BIGINT NOT NULL DEFAULT 0,
    total_output_tokens BIGINT NOT NULL DEFAULT 0,
    status              VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

-- 唯一约束含 context_domain：不同上游域即便 archive_key 相同也不归并。
-- group_id 用 NOT NULL（0=无分组），避免 NULL 破坏唯一性。
CREATE UNIQUE INDEX IF NOT EXISTS uq_conversation_sessions_identity
    ON conversation_sessions(user_id, group_id, context_domain, archive_key);
CREATE INDEX IF NOT EXISTS idx_conversation_sessions_user_last_active
    ON conversation_sessions(user_id, last_active_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_sessions_group
    ON conversation_sessions(group_id);
CREATE INDEX IF NOT EXISTS idx_conversation_sessions_status
    ON conversation_sessions(status);
CREATE INDEX IF NOT EXISTS idx_conversation_sessions_context_domain
    ON conversation_sessions(context_domain);

CREATE TABLE IF NOT EXISTS conversation_branches (
    id               UUID PRIMARY KEY,
    session_id       UUID NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
    parent_branch_id UUID,
    fork_event_id    BIGINT,
    head_event_id    BIGINT,
    event_count      INTEGER NOT NULL DEFAULT 0,
    tail_sequence    INTEGER NOT NULL DEFAULT -1,
    tail_event_hash  VARCHAR(64) NOT NULL DEFAULT '',
    branch_reason    VARCHAR(32) NOT NULL DEFAULT 'initial',
    status           VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_branches_session
    ON conversation_branches(session_id);
CREATE INDEX IF NOT EXISTS idx_conversation_branches_session_status
    ON conversation_branches(session_id, status);
CREATE INDEX IF NOT EXISTS idx_conversation_branches_parent
    ON conversation_branches(parent_branch_id);

CREATE TABLE IF NOT EXISTS conversation_events (
    id                        BIGSERIAL PRIMARY KEY,
    session_id                UUID NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
    branch_id                 UUID NOT NULL REFERENCES conversation_branches(id) ON DELETE CASCADE,
    parent_event_id           BIGINT,
    sequence                  INTEGER NOT NULL,
    request_id                VARCHAR(64) NOT NULL,
    role                      VARCHAR(16) NOT NULL,
    kind                      VARCHAR(32) NOT NULL,
    content_ciphertext        BYTEA,
    content_nonce             BYTEA,
    encryption_key_version    INTEGER NOT NULL DEFAULT 0,
    content_preview           TEXT NOT NULL DEFAULT '',
    event_hash                VARCHAR(64) NOT NULL DEFAULT '',
    model                     VARCHAR(100) NOT NULL DEFAULT '',
    provider                  VARCHAR(32) NOT NULL DEFAULT '',
    upstream_response_id_hash VARCHAR(64) NOT NULL DEFAULT '',
    tool_call_id_hash         VARCHAR(64) NOT NULL DEFAULT '',
    partial                   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 幂等：同一分支同一请求同一序号只写一次（重试安全）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_conversation_events_branch_request_seq
    ON conversation_events(branch_id, request_id, sequence);
CREATE INDEX IF NOT EXISTS idx_conversation_events_session
    ON conversation_events(session_id);
CREATE INDEX IF NOT EXISTS idx_conversation_events_branch_sequence
    ON conversation_events(branch_id, sequence);
CREATE INDEX IF NOT EXISTS idx_conversation_events_request_id
    ON conversation_events(request_id);
CREATE INDEX IF NOT EXISTS idx_conversation_events_event_hash
    ON conversation_events(event_hash);
CREATE INDEX IF NOT EXISTS idx_conversation_events_upstream_response
    ON conversation_events(upstream_response_id_hash);

CREATE TABLE IF NOT EXISTS conversation_response_refs (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL,
    context_domain   VARCHAR(40) NOT NULL,
    response_id_hash VARCHAR(64) NOT NULL,
    session_id       UUID NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
    branch_id        UUID NOT NULL REFERENCES conversation_branches(id) ON DELETE CASCADE,
    tail_event_id    BIGINT,
    durable          BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_conversation_response_refs_lookup
    ON conversation_response_refs(user_id, context_domain, response_id_hash);
CREATE INDEX IF NOT EXISTS idx_conversation_response_refs_session
    ON conversation_response_refs(session_id);
CREATE INDEX IF NOT EXISTS idx_conversation_response_refs_branch
    ON conversation_response_refs(branch_id);
CREATE INDEX IF NOT EXISTS idx_conversation_response_refs_expires
    ON conversation_response_refs(expires_at);
