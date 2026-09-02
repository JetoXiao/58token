UPDATE settings
SET value = REPLACE(value, 'https://useaifor.me', 'https://58token.vip'),
    updated_at = NOW()
WHERE key IN ('help_center_draft_config', 'help_center_published_config')
  AND value LIKE '%https://useaifor.me%';
