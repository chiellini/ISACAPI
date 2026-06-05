-- Rebrand persisted site/admin settings to ISACAPI.
-- This keeps already-initialized deployments aligned with the source defaults.

UPDATE settings
SET value = REPLACE(REPLACE(value, 'Sub' || '2API', 'ISACAPI'), 'ISAC' || 'AI', 'ISACAPI'),
    updated_at = NOW()
WHERE value LIKE '%' || 'Sub' || '2API' || '%'
   OR value LIKE '%' || 'ISAC' || 'AI' || '%';
