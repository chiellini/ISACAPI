ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS agent_status VARCHAR(16) NOT NULL DEFAULT 'inactive',
    ADD COLUMN IF NOT EXISTS agent_status_updated_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS agent_status_updated_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS aff_withdrawal_pending DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS aff_debt DECIMAL(20,8) NOT NULL DEFAULT 0;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_affiliates_agent_status_check') THEN
        ALTER TABLE user_affiliates ADD CONSTRAINT user_affiliates_agent_status_check
            CHECK (agent_status IN ('inactive', 'active', 'suspended'));
    END IF;
END $$;

UPDATE user_affiliates ua
SET agent_status = 'active', agent_status_updated_at = COALESCE(agent_status_updated_at, NOW())
WHERE ua.aff_count > 0 OR ua.aff_history_quota > 0
   OR EXISTS (SELECT 1 FROM user_affiliate_ledger l WHERE l.user_id = ua.user_id);
CREATE INDEX IF NOT EXISTS idx_user_affiliates_agent_status ON user_affiliates(agent_status);

CREATE TABLE IF NOT EXISTS user_affiliate_payment_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('alipay', 'bank_card', 'usdt')),
    details_encrypted TEXT NOT NULL,
    masked_summary VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_payment_accounts_user_id ON user_affiliate_payment_accounts(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_payment_accounts_one_default
    ON user_affiliate_payment_accounts(user_id) WHERE is_default;

CREATE TABLE IF NOT EXISTS user_affiliate_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount DECIMAL(20,8) NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted','approved','paid','rejected','canceled')),
    payment_account_id BIGINT NULL REFERENCES user_affiliate_payment_accounts(id) ON DELETE SET NULL,
    payment_account_type VARCHAR(20) NOT NULL,
    payment_details_encrypted TEXT NOT NULL,
    payment_account_summary VARCHAR(255) NOT NULL,
    idempotency_key VARCHAR(128) NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ NULL,
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    paid_at TIMESTAMPTZ NULL,
    paid_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    canceled_at TIMESTAMPTZ NULL,
    cancel_reason TEXT NULL,
    reject_reason TEXT NULL,
    actual_currency VARCHAR(20) NULL,
    actual_amount DECIMAL(20,8) NULL,
    exchange_rate DECIMAL(20,8) NULL,
    external_reference VARCHAR(255) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_withdrawals_user_created ON user_affiliate_withdrawals(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_withdrawals_status_created ON user_affiliate_withdrawals(status, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_withdrawals_user_idempotency
    ON user_affiliate_withdrawals(user_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_ledger_accrue_source_order
    ON user_affiliate_ledger(source_order_id) WHERE action='accrue' AND source_order_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_affiliate_admin_audits (
    id BIGSERIAL PRIMARY KEY,
    operator_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    withdrawal_id BIGINT NULL REFERENCES user_affiliate_withdrawals(id) ON DELETE SET NULL,
    action VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(128) NULL,
    detail JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_admin_audits_target_created ON user_affiliate_admin_audits(target_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_admin_audits_withdrawal_created ON user_affiliate_admin_audits(withdrawal_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_admin_audits_idempotency
    ON user_affiliate_admin_audits(action, idempotency_key) WHERE idempotency_key IS NOT NULL;

INSERT INTO settings (key, value, updated_at)
VALUES ('affiliate_minimum_withdrawal', '10', NOW()) ON CONFLICT (key) DO NOTHING;
INSERT INTO settings (key, value, updated_at)
VALUES ('affiliate_rebate_freeze_hours', '168', NOW()) ON CONFLICT (key) DO NOTHING;
UPDATE settings SET value = '168', updated_at = NOW()
WHERE key = 'affiliate_rebate_freeze_hours' AND value = '0';
