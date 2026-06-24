ALTER TABLE users
  ADD COLUMN IF NOT EXISTS allow_balance_subscription_purchase BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN users.allow_balance_subscription_purchase IS 'Whether the user may purchase subscription plans using account balance.';
