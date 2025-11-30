CREATE TABLE IF NOT EXISTS user_tokens (
    user_token_id  varchar(255) PRIMARY KEY NOT NULL,
    user_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    refresh_token  TEXT NOT NULL, 
    device_info    VARCHAR(255),          
    expires_at     TIMESTAMPTZ NOT NULL,
    is_revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_refresh_token ON user_tokens(refresh_token);