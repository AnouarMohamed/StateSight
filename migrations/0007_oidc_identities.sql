CREATE TABLE IF NOT EXISTS user_identities (
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issuer, subject),
    UNIQUE (issuer, user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities(user_id);
