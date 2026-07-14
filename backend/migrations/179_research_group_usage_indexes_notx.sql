CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_logs_payer_created_idx
    ON usage_logs (payer_user_id, created_at DESC)
    WHERE payer_user_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_logs_research_group_created_idx
    ON usage_logs (research_group_id, created_at DESC)
    WHERE research_group_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_logs_research_group_member_created_idx
    ON usage_logs (research_group_member_id, created_at DESC)
    WHERE research_group_member_id IS NOT NULL;
