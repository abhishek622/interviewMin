CREATE TABLE IF NOT EXISTS user_tokens (
    token_id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash     TEXT NOT NULL, 
    device_info    TEXT,          
    expires_at     TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_revoked     BOOLEAN NOT NULL DEFAULT FALSE
);


CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_token_hash ON user_tokens(token_hash);