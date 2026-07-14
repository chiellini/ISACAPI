-- Idempotency-Key is scoped to a user.  Before adding the unique guard, keep
-- one canonical key owner for historical duplicates without deleting any job
-- or changing any billing/hold record.  Prefer a job whose hold was actually
-- committed; among otherwise equivalent rows the earliest job is the winner.
WITH ranked_jobs AS (
    SELECT
        j.id,
        ROW_NUMBER() OVER (
            PARTITION BY j.user_id, j.idempotency_key
            ORDER BY
                CASE WHEN EXISTS (
                    SELECT 1
                    FROM usage_billing_dedup d
                    WHERE d.request_id = 'batch_image_hold:' || j.batch_id
                      AND d.api_key_id = j.api_key_id
                ) OR EXISTS (
                    SELECT 1
                    FROM usage_billing_dedup_archive a
                    WHERE a.request_id = 'batch_image_hold:' || j.batch_id
                      AND a.api_key_id = j.api_key_id
                ) THEN 0 ELSE 1 END,
                j.id ASC
        ) AS duplicate_rank
    FROM batch_image_jobs j
    WHERE j.idempotency_key IS NOT NULL
      AND j.idempotency_key <> ''
), duplicate_jobs AS (
    SELECT id
    FROM ranked_jobs
    WHERE duplicate_rank > 1
)
UPDATE batch_image_jobs j
SET idempotency_key = NULL,
    updated_at = NOW()
FROM duplicate_jobs d
WHERE j.id = d.id;

DROP INDEX IF EXISTS batch_image_jobs_idempotency_key_idx;

CREATE UNIQUE INDEX IF NOT EXISTS batch_image_jobs_user_idempotency_key_uq
    ON batch_image_jobs (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';

COMMENT ON INDEX batch_image_jobs_user_idempotency_key_uq IS
    'Prevents concurrent duplicate batch jobs and balance holds for a user-scoped Idempotency-Key';
