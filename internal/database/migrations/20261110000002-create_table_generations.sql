-- +migrate Up
CREATE TABLE IF NOT EXISTS generations (
  id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  source_url VARCHAR(1024) NOT NULL DEFAULT '',
  tone VARCHAR(32) NOT NULL,
  audience VARCHAR(255) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
  raw_text TEXT NOT NULL,
  result_json TEXT NULL,
  latency_ms BIGINT NOT NULL DEFAULT 0,
  tokens INTEGER NOT NULL DEFAULT 0,
  confidence VARCHAR(16) NOT NULL DEFAULT '',
  semantic_deviation VARCHAR(16) NOT NULL DEFAULT '',
  model VARCHAR(64) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_generations_user_created ON generations (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_generations_status ON generations (status);

-- +migrate Down
DROP TABLE IF EXISTS generations;
