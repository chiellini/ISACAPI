-- Repair the final state left by the legacy custom runner, which executed both
-- goose Up and Down sections in migrations 019, 024, and 037.

-- Restore the active WeChat attribute. Reuse the original soft-deleted row when
-- possible so existing references keep the same attribute id.
DO $$
DECLARE
    wechat_attribute_id BIGINT;
BEGIN
    SELECT id
    INTO wechat_attribute_id
    FROM user_attribute_definitions
    WHERE key = 'wechat'
      AND deleted_at IS NULL
    ORDER BY id
    LIMIT 1;

    IF wechat_attribute_id IS NULL THEN
        SELECT id
        INTO wechat_attribute_id
        FROM user_attribute_definitions
        WHERE key = 'wechat'
        ORDER BY id
        LIMIT 1
        FOR UPDATE;

        IF wechat_attribute_id IS NULL THEN
            INSERT INTO user_attribute_definitions (
                key, name, description, type, options, required, validation,
                placeholder, display_order, enabled, created_at, updated_at
            ) VALUES (
                'wechat', '微信', '用户微信号', 'text', '[]'::jsonb, FALSE,
                '{}'::jsonb, '请输入微信号', 0, TRUE, NOW(), NOW()
            )
            RETURNING id INTO wechat_attribute_id;
        ELSE
            UPDATE user_attribute_definitions
            SET deleted_at = NULL,
                enabled = TRUE,
                updated_at = NOW()
            WHERE id = wechat_attribute_id;
        END IF;
    END IF;

    -- The broken 019 migration restored users.wechat after deleting its copied
    -- attribute values. Correct installations no longer have this column.
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'users'
          AND column_name = 'wechat'
    ) THEN
        EXECUTE $backfill$
            INSERT INTO user_attribute_values (
                user_id, attribute_id, value, created_at, updated_at
            )
            SELECT u.id, $1, BTRIM(u.wechat), NOW(), NOW()
            FROM users u
            WHERE NULLIF(BTRIM(u.wechat), '') IS NOT NULL
              AND u.deleted_at IS NULL
            ON CONFLICT (user_id, attribute_id) DO NOTHING
        $backfill$ USING wechat_attribute_id;
    END IF;
END $$;

ALTER TABLE users DROP COLUMN IF EXISTS wechat;

-- Restore the legacy Gemini Code Assist tier marker removed by migration 024's
-- Down section. Preserve every existing non-empty tier id.
UPDATE accounts
SET credentials = jsonb_set(credentials, '{tier_id}', '"LEGACY"'::jsonb, TRUE)
WHERE platform = 'gemini'
  AND type = 'oauth'
  AND jsonb_typeof(credentials) = 'object'
  AND NULLIF(BTRIM(credentials->>'tier_id'), '') IS NULL
  AND (
      credentials->>'oauth_type' = 'code_assist'
      OR (
          credentials->>'oauth_type' IS NULL
          AND credentials->>'project_id' IS NOT NULL
      )
  );

-- Restore the table dropped by migration 037's Down section.
CREATE TABLE IF NOT EXISTS ops_alert_silences (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    platform VARCHAR(64) NOT NULL,
    group_id BIGINT,
    region VARCHAR(64),
    until TIMESTAMPTZ NOT NULL,
    reason TEXT,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_alert_silences_lookup
    ON ops_alert_silences (rule_id, platform, group_id, region, until);
