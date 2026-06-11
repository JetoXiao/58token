ALTER TABLE subscription_plans
  ADD COLUMN IF NOT EXISTS limited_offer_price DECIMAL(20,2),
  ADD COLUMN IF NOT EXISTS limited_offer_expires_at TIMESTAMPTZ;

