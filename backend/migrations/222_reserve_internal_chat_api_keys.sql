-- Reserve built-in chat API key names and make lazy creation safe across
-- multiple application instances. These credentials are service-managed and
-- are never part of the user's public API-key collection.

-- The marker is database-enforced, not merely a naming convention. Adding the
-- NOT VALID constraint before retiring historical rows closes the rolling-
-- upgrade window: an old application instance can no longer recreate a
-- user-visible reserved key after the cleanup commits. The ALTER lock is held
-- until this migration transaction commits, so an insert is either included in
-- the UPDATE below or checked by the constraint.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE api_keys
    ADD CONSTRAINT api_keys_reserved_chat_name_requires_internal
    CHECK (
        deleted_at IS NOT NULL
        OR LEFT(name, 19) <> '__chat_playground__'
        OR is_internal
    ) NOT VALID;

-- No historical reserved key can be proven service-generated: before this
-- release a user could pre-create either reserved name and retain its secret.
-- Retire all of them. The service recreates a fresh per-group credential on
-- the next Chat request; usage-history foreign keys continue to reference the
-- soft-deleted row.
UPDATE api_keys
SET status = 'disabled',
    deleted_at = NOW(),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND LEFT(name, 19) = '__chat_playground__';

ALTER TABLE api_keys
    VALIDATE CONSTRAINT api_keys_reserved_chat_name_requires_internal;
