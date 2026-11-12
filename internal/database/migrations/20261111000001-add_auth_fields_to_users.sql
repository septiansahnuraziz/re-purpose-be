-- +migrate Up
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255),
  ADD COLUMN IF NOT EXISTS google_id VARCHAR(255),
  ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(32) NOT NULL DEFAULT 'password',
  ADD COLUMN IF NOT EXISTS primary_role VARCHAR(32),
  ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;

ALTER TABLE users ADD CONSTRAINT uq_users_google_id UNIQUE (google_id);

-- +migrate Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS uq_users_google_id;
ALTER TABLE users
  DROP COLUMN IF EXISTS last_login_at,
  DROP COLUMN IF EXISTS email_verified_at,
  DROP COLUMN IF EXISTS primary_role,
  DROP COLUMN IF EXISTS auth_provider,
  DROP COLUMN IF EXISTS google_id,
  DROP COLUMN IF EXISTS password_hash;
