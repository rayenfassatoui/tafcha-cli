CREATE TABLE IF NOT EXISTS snippets (
    id VARCHAR(20) PRIMARY KEY,
    content BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snippets_expires_at ON snippets(expires_at);
