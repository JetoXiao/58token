INSERT INTO settings (key, value, updated_at)
VALUES (
  'support_contact_config',
  '{"enabled":true,"title":"售后联系","description":"如需售后支持，请添加下方客服微信联系。","contacts":[]}',
  NOW()
)
ON CONFLICT (key) DO NOTHING;

INSERT INTO settings (key, value, updated_at)
VALUES (
  'user_menu_items',
  '["dashboard","api_keys","image_generation","usage","channel_status","subscriptions","purchase","orders","redeem","affiliate","support_contact","profile"]',
  NOW()
)
ON CONFLICT (key) DO UPDATE
SET
  value = CASE
    WHEN settings.value IS NULL OR btrim(settings.value) = '' THEN EXCLUDED.value
    WHEN btrim(settings.value) NOT LIKE '[%' THEN EXCLUDED.value
    WHEN settings.value::jsonb ? 'support_contact' THEN settings.value
    ELSE (settings.value::jsonb || '["support_contact"]'::jsonb)::text
  END,
  updated_at = NOW();
