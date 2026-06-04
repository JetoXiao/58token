-- Partner program: user-linked partner levels and self-service applications.

ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS partner_level VARCHAR(32) NOT NULL DEFAULT 'none';

CREATE INDEX IF NOT EXISTS idx_user_affiliates_partner_level
    ON user_affiliates(partner_level)
    WHERE partner_level <> 'none';

COMMENT ON COLUMN user_affiliates.partner_level IS 'Partner level: none|spark|voyage|summit|cocreate';

CREATE TABLE IF NOT EXISTS user_affiliate_partner_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requested_level VARCHAR(32) NOT NULL,
    source VARCHAR(128) NOT NULL,
    strengths TEXT NOT NULL,
    portal_url VARCHAR(512) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    granted_level VARCHAR(32) NULL,
    review_note TEXT NULL,
    reviewer_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE user_affiliate_partner_applications
    ADD COLUMN IF NOT EXISTS granted_level VARCHAR(32) NULL;

CREATE INDEX IF NOT EXISTS idx_affiliate_partner_applications_user_id
    ON user_affiliate_partner_applications(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_affiliate_partner_applications_status
    ON user_affiliate_partner_applications(status, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliate_partner_applications_pending_user
    ON user_affiliate_partner_applications(user_id)
    WHERE status = 'pending';

COMMENT ON TABLE user_affiliate_partner_applications IS 'User applications for partner levels';
COMMENT ON COLUMN user_affiliate_partner_applications.requested_level IS 'Requested partner level';
COMMENT ON COLUMN user_affiliate_partner_applications.source IS 'Where the user learned about the partner program';
COMMENT ON COLUMN user_affiliate_partner_applications.strengths IS 'Promotion strengths or community advantages';
COMMENT ON COLUMN user_affiliate_partner_applications.portal_url IS 'Public profile/community portal URL';
COMMENT ON COLUMN user_affiliate_partner_applications.granted_level IS 'Partner level granted during review';

UPDATE settings
SET value = '5',
    updated_at = NOW()
WHERE key = 'affiliate_rebate_rate'
  AND value ~ '^[0-9]+(\.[0-9]+)?$'
  AND value::numeric = 20;
