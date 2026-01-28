-- Snippets table for storing text content
CREATE TABLE IF NOT EXISTS snippets (
    id VARCHAR(12) PRIMARY KEY,
    content BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient expired snippet cleanup
CREATE INDEX IF NOT EXISTS idx_snippets_expires_at ON snippets(expires_at);

-- Index for efficient lookups by ID with expiry check
CREATE INDEX IF NOT EXISTS idx_snippets_id_expires_at ON snippets(id, expires_at);
