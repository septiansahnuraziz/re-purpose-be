-- +migrate Up
CREATE TABLE IF NOT EXISTS auth_sessions (
  id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  refresh_token_hash VARCHAR(64) NOT NULL,
  user_agent VARCHAR(512) NOT NULL DEFAULT '',
  ip_address VARCHAR(64) NOT NULL DEFAULT '',
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  CONSTRAINT uq_auth_sessions_refresh UNIQUE (refresh_token_hash)
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user ON auth_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires ON auth_sessions (expires_at);

-- +migrate Down
DROP TABLE IF EXISTS auth_sessions;
