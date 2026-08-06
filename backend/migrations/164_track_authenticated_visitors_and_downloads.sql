-- Associate visit and download records with an authenticated user when the
-- public request includes a valid access token. Anonymous traffic remains
-- represented only by its IP address.

ALTER TABLE download_resource_downloads
    ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_download_resource_downloads_user_time
    ON download_resource_downloads (user_id, requested_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_events_user_time
    ON visitor_events (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;
