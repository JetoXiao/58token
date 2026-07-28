-- First-party visitor analytics and channel attribution.
-- IP addresses are retained only for administrator-facing operational analytics.

CREATE TABLE IF NOT EXISTS visitor_channels (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(64) NOT NULL UNIQUE,
    destination_path VARCHAR(512) NOT NULL DEFAULT '/home',
    description VARCHAR(500) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS visitor_events (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT REFERENCES visitor_channels(id) ON DELETE SET NULL,
    channel_code VARCHAR(64) NOT NULL DEFAULT 'direct',
    visitor_id VARCHAR(64) NOT NULL DEFAULT '',
    session_id VARCHAR(64) NOT NULL DEFAULT '',
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ip VARCHAR(45) NOT NULL,
    country_code VARCHAR(8) NOT NULL DEFAULT '',
    path VARCHAR(512) NOT NULL DEFAULT '/',
    referrer VARCHAR(1024) NOT NULL DEFAULT '',
    landing_url VARCHAR(1024) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    language VARCHAR(32) NOT NULL DEFAULT '',
    screen VARCHAR(32) NOT NULL DEFAULT '',
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visitor_events_occurred_at ON visitor_events (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_visitor_events_channel_time ON visitor_events (channel_code, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_visitor_events_ip_time ON visitor_events (ip, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_visitor_events_visitor_time ON visitor_events (visitor_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS visitor_ip_geolocation_cache (
    ip VARCHAR(45) PRIMARY KEY,
    country VARCHAR(120) NOT NULL DEFAULT '',
    country_code VARCHAR(8) NOT NULL DEFAULT '',
    region VARCHAR(160) NOT NULL DEFAULT '',
    city VARCHAR(160) NOT NULL DEFAULT '',
    timezone VARCHAR(80) NOT NULL DEFAULT '',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    provider VARCHAR(40) NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_visitor_ip_geo_expires_at ON visitor_ip_geolocation_cache (expires_at);

CREATE TABLE IF NOT EXISTS visitor_analytics_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    retention_days INTEGER NOT NULL DEFAULT 90 CHECK (retention_days BETWEEN 7 AND 730),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO visitor_analytics_settings (id, enabled, retention_days)
VALUES (1, TRUE, 90)
ON CONFLICT (id) DO NOTHING;
