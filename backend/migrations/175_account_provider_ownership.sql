-- Track the user/provider that contributed each account to the shared pool.
-- Existing accounts remain platform-owned (provider_id IS NULL).
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS provider_id BIGINT;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_accounts_provider_id'
    ) THEN
        ALTER TABLE accounts
            ADD CONSTRAINT fk_accounts_provider_id
            FOREIGN KEY (provider_id)
            REFERENCES users(id)
            ON DELETE SET NULL
            NOT VALID;
    END IF;
END $$;

ALTER TABLE accounts
    VALIDATE CONSTRAINT fk_accounts_provider_id;

CREATE INDEX IF NOT EXISTS accounts_provider_id_idx
    ON accounts (provider_id);
