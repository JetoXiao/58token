-- Public download catalog. Package binaries remain private in object storage;
-- this table only stores the metadata needed to issue a short-lived download.

CREATE TABLE IF NOT EXISTS download_resources (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL UNIQUE,
    name_zh VARCHAR(160) NOT NULL DEFAULT '',
    name_en VARCHAR(160) NOT NULL DEFAULT '',
    description_zh VARCHAR(1000) NOT NULL DEFAULT '',
    description_en VARCHAR(1000) NOT NULL DEFAULT '',
    version VARCHAR(64) NOT NULL DEFAULT '',
    platform VARCHAR(64) NOT NULL DEFAULT '',
    object_key VARCHAR(512) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(128) NOT NULL DEFAULT 'application/octet-stream',
    size_bytes BIGINT NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    checksum_sha256 VARCHAR(64) NOT NULL DEFAULT '',
    published BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    download_count BIGINT NOT NULL DEFAULT 0 CHECK (download_count >= 0),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_download_resources_public
    ON download_resources (published, sort_order DESC, uploaded_at DESC);

CREATE TABLE IF NOT EXISTS download_resource_downloads (
    id BIGSERIAL PRIMARY KEY,
    resource_id BIGINT NOT NULL REFERENCES download_resources(id) ON DELETE CASCADE,
    ip VARCHAR(45) NOT NULL,
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    referrer VARCHAR(1024) NOT NULL DEFAULT '',
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_download_resource_downloads_recent
    ON download_resource_downloads (requested_at DESC);
CREATE INDEX IF NOT EXISTS idx_download_resource_downloads_resource_time
    ON download_resource_downloads (resource_id, requested_at DESC);
CREATE INDEX IF NOT EXISTS idx_download_resource_downloads_ip_time
    ON download_resource_downloads (ip, requested_at DESC);

-- Make the new public navigation entry visible for installations that still use
-- the original default menu. Explicitly customized menus are left untouched.
UPDATE settings
SET value = '["models","docs","resources","partner"]', updated_at = NOW()
WHERE key = 'marketing_nav_items'
  AND value = '["models","docs","partner"]';
