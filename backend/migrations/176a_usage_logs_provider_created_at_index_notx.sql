-- Runs outside a transaction because PostgreSQL forbids concurrent index creation in one.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_provider_created_at
    ON usage_logs (provider_id, created_at)
    WHERE provider_id IS NOT NULL;
