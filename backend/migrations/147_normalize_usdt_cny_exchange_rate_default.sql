-- Normalize the previous shipped payment USDT/CNY exchange-rate default to 7.
-- Only values that still match the old shipped default are changed; later manual settings are preserved.
UPDATE settings
SET value = '7',
    updated_at = NOW()
WHERE key = 'USDT_CNY_EXCHANGE_RATE'
  AND trim(value) ~ '^[0-9]+(\.[0-9]+)?$'
  AND trim(value)::numeric = (72::numeric / 10);
