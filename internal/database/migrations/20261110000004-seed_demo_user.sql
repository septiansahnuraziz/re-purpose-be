-- +migrate Up
INSERT INTO users (id, email, name, avatar_url, tier, daily_credits, monthly_credits)
VALUES ('usr_demo', 'demo@repurpose.ai', 'Demo Creator', '', 'PRO', 25, 250)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  tier = EXCLUDED.tier;

-- +migrate Down
DELETE FROM users WHERE id = 'usr_demo';
