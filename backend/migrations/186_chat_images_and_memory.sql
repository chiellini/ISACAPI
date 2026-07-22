-- 内置聊天 Playground：生成/上传图片的服务端持久化 + 会话记忆（长/中/短期）。
--
-- 此前生成图片只落浏览器 IndexedDB（不跨设备、清缓存即丢），此处改为落库（BYTEA），
-- 由带鉴权的接口回读。会话记忆用三列表达：
--   summary          —— 中期滚动摘要（较早若干轮被压缩成的连续叙述）
--   memory           —— 长期稳定事实（用户偏好 / 关键结论 / 实体，跨轮保持）
--   summarized_count —— 已折叠进摘要的「前缀消息条数」，其后的消息按原文进入短期上下文

CREATE TABLE IF NOT EXISTS chat_images (
    id         TEXT        PRIMARY KEY,
    session_id BIGINT      NOT NULL REFERENCES chat_sessions (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL,
    mime       VARCHAR(64) NOT NULL DEFAULT 'image/png',
    byte_size  INTEGER     NOT NULL DEFAULT 0,
    data       BYTEA       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_images_session ON chat_images (session_id);
CREATE INDEX IF NOT EXISTS idx_chat_images_user ON chat_images (user_id);

ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS summary          TEXT    NOT NULL DEFAULT '';
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS memory           TEXT    NOT NULL DEFAULT '';
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS summarized_count INTEGER NOT NULL DEFAULT 0;
