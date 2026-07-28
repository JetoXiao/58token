-- Default marketing attribution channels for visitor analytics.
-- Keep codes stable because they are embedded in campaign URLs.

INSERT INTO visitor_channels (name, code, destination_path, description, active)
VALUES
    ('Reddit', 'reddit', '/home', 'Reddit campaign traffic', TRUE),
    ('Twitter', 'twitter', '/home', 'Twitter campaign traffic', TRUE),
    ('Discord', 'discord', '/home', 'Discord campaign traffic', TRUE),
    ('Tencent', 'tencent', '/home', 'Tencent campaign traffic', TRUE),
    ('WeChat', 'wechat', '/home', 'WeChat campaign traffic', TRUE),
    ('Warpcast', 'warpcast', '/home', 'Warpcast campaign traffic', TRUE),
    ('Telegram', 'telegram', '/home', 'Telegram campaign traffic', TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    active = TRUE,
    updated_at = NOW();
