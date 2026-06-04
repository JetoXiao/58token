INSERT INTO settings (key, value, updated_at)
VALUES ('marketing_nav_items', '["models","docs","partner"]', NOW())
ON CONFLICT (key) DO NOTHING;
