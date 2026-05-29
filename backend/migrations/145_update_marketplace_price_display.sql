UPDATE model_marketplace_items
SET
    official_prices = '{"input":5,"output":30,"cacheWrite":null,"cacheRead":0.5}'::jsonb,
    updated_at = NOW()
WHERE model_name = 'gpt-5.5';

UPDATE model_marketplace_items
SET
    official_prices = '{"input":2.5,"output":15,"cacheWrite":null,"cacheRead":0.25}'::jsonb,
    updated_at = NOW()
WHERE model_name = 'gpt-5.4';

UPDATE model_marketplace_items
SET
    tags = tags - 'Fast Mode',
    official_prices = '{"input":5,"output":25,"cacheWrite":6.25,"cacheRead":0.5}'::jsonb,
    updated_at = NOW()
WHERE model_name = 'claude-opus-4.8';
