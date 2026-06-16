CREATE TABLE IF NOT EXISTS user_free_quota_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(32) NOT NULL,
    source_id TEXT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    remaining_amount DECIMAL(20, 8) NOT NULL,
    allowed_group_ids BIGINT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notes TEXT NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (amount >= 0),
    CHECK (remaining_amount >= 0),
    CHECK (remaining_amount <= amount)
);

CREATE INDEX IF NOT EXISTS idx_user_free_quota_ledger_user_status
    ON user_free_quota_ledger(user_id, status)
    WHERE remaining_amount > 0;

CREATE INDEX IF NOT EXISTS idx_user_free_quota_ledger_allowed_groups
    ON user_free_quota_ledger USING GIN (allowed_group_ids);

CREATE TABLE IF NOT EXISTS trial_cards (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL DEFAULT '',
    amount DECIMAL(20, 8) NOT NULL,
    max_redemptions INTEGER NOT NULL DEFAULT 1,
    redeemed_count INTEGER NOT NULL DEFAULT 0,
    per_user_limit INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notes TEXT NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (amount >= 0),
    CHECK (max_redemptions >= 0),
    CHECK (redeemed_count >= 0),
    CHECK (redeemed_count <= max_redemptions),
    CHECK (per_user_limit > 0)
);

CREATE INDEX IF NOT EXISTS idx_trial_cards_status
    ON trial_cards(status);

CREATE TABLE IF NOT EXISTS trial_card_redemptions (
    id BIGSERIAL PRIMARY KEY,
    trial_card_id BIGINT NOT NULL REFERENCES trial_cards(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trial_card_redemptions_card_user
    ON trial_card_redemptions(trial_card_id, user_id);

CREATE INDEX IF NOT EXISTS idx_trial_card_redemptions_user
    ON trial_card_redemptions(user_id);

INSERT INTO settings (key, value)
VALUES
    ('free_quota_invitation_enabled', 'false'),
    ('free_quota_invitation_amount', '0'),
    ('free_quota_group_ids', '[]'),
    ('free_quota_show_locked_groups', 'false'),
    ('free_quota_transfer_on_payment', 'false')
ON CONFLICT (key) DO NOTHING;
