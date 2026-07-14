-- Snapshot the account provider at request submission time. Historical rows stay
-- NULL because deriving their owner from the account's current assignment would
-- rewrite history after an account transfer.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS provider_id BIGINT NULL;

-- Batch image jobs settle asynchronously, so retain the provider assignment that
-- existed when the job was submitted instead of reading the account at settlement.
ALTER TABLE batch_image_jobs
    ADD COLUMN IF NOT EXISTS account_provider_id BIGINT NULL;
