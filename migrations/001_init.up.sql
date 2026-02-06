CREATE TABLE users (
    id UUID PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE secrets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    encrypted_payload BYTEA NOT NULL,
    meta JSONB,
    version BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_secrets_user_id ON secrets(user_id);
CREATE INDEX idx_secrets_updated_at ON secrets(updated_at);
