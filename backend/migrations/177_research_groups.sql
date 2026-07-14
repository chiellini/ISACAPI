CREATE TABLE IF NOT EXISTS research_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    dissolved_at TIMESTAMPTZ,
    CONSTRAINT research_groups_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT research_groups_status_check CHECK (status IN ('active', 'paused', 'dissolved')),
    CONSTRAINT research_groups_active_owner_check CHECK (status = 'dissolved' OR owner_user_id IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS research_groups_owner_active_uidx
    ON research_groups (owner_user_id) WHERE status <> 'dissolved';
CREATE INDEX IF NOT EXISTS research_groups_status_idx ON research_groups (status);

CREATE TABLE IF NOT EXISTS research_group_members (
    id BIGSERIAL PRIMARY KEY,
    research_group_id BIGINT NOT NULL REFERENCES research_groups(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    monthly_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    monthly_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    monthly_reserved_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    -- Billing-calendar boundaries are calculated in the configured accounting
    -- timezone by the application and lazily normalized on read.
    usage_window_start TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invited_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    removed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT research_group_members_status_check CHECK (status IN ('pending', 'active', 'paused', 'removed')),
    CONSTRAINT research_group_members_limit_nonnegative CHECK (monthly_limit_usd >= 0),
    CONSTRAINT research_group_members_usage_nonnegative CHECK (monthly_usage_usd >= 0),
    CONSTRAINT research_group_members_reserved_nonnegative CHECK (monthly_reserved_usd >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS research_group_members_group_user_effective_uidx
    ON research_group_members (research_group_id, user_id)
    WHERE status IN ('pending', 'active', 'paused');
CREATE UNIQUE INDEX IF NOT EXISTS research_group_members_user_effective_uidx
    ON research_group_members (user_id)
    WHERE status IN ('pending', 'active', 'paused');
CREATE INDEX IF NOT EXISTS research_group_members_group_status_idx
    ON research_group_members (research_group_id, status);

CREATE TABLE IF NOT EXISTS research_group_quota_audits (
    id BIGSERIAL PRIMARY KEY,
    research_group_id BIGINT NOT NULL REFERENCES research_groups(id) ON DELETE RESTRICT,
    -- Snapshot identifiers deliberately have no FK: deleting a student or
    -- actor must not rewrite an append-only audit record.
    member_id BIGINT,
    actor_user_id BIGINT,
    action VARCHAR(50) NOT NULL,
    amount_usd DECIMAL(20,10),
    previous_value_usd DECIMAL(20,10),
    new_value_usd DECIMAL(20,10),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT research_group_quota_audits_action_not_blank CHECK (btrim(action) <> '')
);

CREATE INDEX IF NOT EXISTS research_group_quota_audits_group_created_idx
    ON research_group_quota_audits (research_group_id, created_at DESC);
CREATE INDEX IF NOT EXISTS research_group_quota_audits_member_created_idx
    ON research_group_quota_audits (member_id, created_at DESC);
CREATE INDEX IF NOT EXISTS research_group_quota_audits_actor_created_idx
    ON research_group_quota_audits (actor_user_id, created_at DESC);

CREATE OR REPLACE FUNCTION prevent_research_group_quota_audit_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'research_group_quota_audits is append-only' USING ERRCODE = '55000';
END;
$$;

DROP TRIGGER IF EXISTS research_group_quota_audits_append_only ON research_group_quota_audits;
CREATE TRIGGER research_group_quota_audits_append_only
    BEFORE UPDATE OR DELETE ON research_group_quota_audits
    FOR EACH ROW EXECUTE FUNCTION prevent_research_group_quota_audit_mutation();

COMMENT ON TABLE research_group_quota_audits IS 'Append-only research group membership and quota operation audit log';
