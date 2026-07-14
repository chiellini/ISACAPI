ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS payer_user_id BIGINT,
    ADD COLUMN IF NOT EXISTS research_group_id BIGINT,
    ADD COLUMN IF NOT EXISTS research_group_member_id BIGINT,
    ADD COLUMN IF NOT EXISTS funding_source VARCHAR(20);

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'usage_logs_funding_source_check') THEN
        ALTER TABLE usage_logs ADD CONSTRAINT usage_logs_funding_source_check
            CHECK (funding_source IS NULL OR funding_source IN ('self', 'research_group')) NOT VALID;
    END IF;
    -- payer_user_id and research_group_member_id are immutable snapshots, not
    -- live ownership links. No FK is added so account/member deletion cannot
    -- erase or rewrite historical attribution.
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_usage_logs_research_group_id') THEN
        ALTER TABLE usage_logs ADD CONSTRAINT fk_usage_logs_research_group_id FOREIGN KEY (research_group_id) REFERENCES research_groups(id) ON DELETE SET NULL NOT VALID;
    END IF;
END $$;

ALTER TABLE usage_logs VALIDATE CONSTRAINT usage_logs_funding_source_check;
ALTER TABLE usage_logs VALIDATE CONSTRAINT fk_usage_logs_research_group_id;

ALTER TABLE batch_image_jobs
    ADD COLUMN IF NOT EXISTS payer_user_id BIGINT,
    ADD COLUMN IF NOT EXISTS research_group_id BIGINT,
    ADD COLUMN IF NOT EXISTS research_group_member_id BIGINT,
    ADD COLUMN IF NOT EXISTS funding_source VARCHAR(20);

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'batch_image_jobs_funding_source_check') THEN
        ALTER TABLE batch_image_jobs ADD CONSTRAINT batch_image_jobs_funding_source_check
            CHECK (funding_source IS NULL OR funding_source IN ('self', 'research_group')) NOT VALID;
    END IF;
    -- payer/member identifiers remain immutable settlement snapshots.
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_batch_image_jobs_research_group_id') THEN
        ALTER TABLE batch_image_jobs ADD CONSTRAINT fk_batch_image_jobs_research_group_id FOREIGN KEY (research_group_id) REFERENCES research_groups(id) ON DELETE SET NULL NOT VALID;
    END IF;
END $$;

ALTER TABLE batch_image_jobs VALIDATE CONSTRAINT batch_image_jobs_funding_source_check;
ALTER TABLE batch_image_jobs VALIDATE CONSTRAINT fk_batch_image_jobs_research_group_id;

CREATE INDEX IF NOT EXISTS batch_image_jobs_payer_created_idx ON batch_image_jobs (payer_user_id, created_at DESC) WHERE payer_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS batch_image_jobs_research_group_created_idx ON batch_image_jobs (research_group_id, created_at DESC) WHERE research_group_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS batch_image_jobs_research_group_member_created_idx ON batch_image_jobs (research_group_member_id, created_at DESC) WHERE research_group_member_id IS NOT NULL;
