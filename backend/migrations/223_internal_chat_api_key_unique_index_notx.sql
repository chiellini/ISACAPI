-- Create the cross-instance uniqueness guard without blocking writes to the
-- api_keys table for the duration of an index build on large installations.
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS api_keys_internal_chat_name_unique_active
    ON api_keys (user_id, name)
    WHERE deleted_at IS NULL
      AND LEFT(name, 19) = '__chat_playground__';
