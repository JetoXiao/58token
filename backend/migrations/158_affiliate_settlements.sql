CREATE TABLE IF NOT EXISTS user_affiliate_settlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL CHECK (amount > 0),
    settled_on DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_settlements_user_id
    ON user_affiliate_settlements(user_id, settled_on DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_settlements_settled_on
    ON user_affiliate_settlements(settled_on DESC, id DESC);

COMMENT ON TABLE user_affiliate_settlements IS 'Affiliate partner offline cashback settlement records';
COMMENT ON COLUMN user_affiliate_settlements.user_id IS 'Affiliate partner user receiving the offline cashback';
COMMENT ON COLUMN user_affiliate_settlements.amount IS 'Offline cashback amount recorded by admin';
COMMENT ON COLUMN user_affiliate_settlements.settled_on IS 'Actual cashback settlement date';
COMMENT ON COLUMN user_affiliate_settlements.note IS 'Settlement note for reconciliation';
COMMENT ON COLUMN user_affiliate_settlements.created_by IS 'Admin user who recorded the settlement';
