-- +migrate Up
CREATE TABLE IF NOT EXISTS sponsor_clicks (
  id VARCHAR(64) NOT NULL,
  sponsor_id VARCHAR(64) NOT NULL,
  event VARCHAR(32) NOT NULL,
  user_id VARCHAR(64) NOT NULL DEFAULT '',
  ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(512) NOT NULL DEFAULT '',
  occurred_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_sponsor_clicks_sponsor_event ON sponsor_clicks (sponsor_id, event);
CREATE INDEX IF NOT EXISTS idx_sponsor_clicks_created ON sponsor_clicks (created_at);

-- +migrate Down
DROP TABLE IF EXISTS sponsor_clicks;
