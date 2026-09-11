-- Sync the model marketplace's displayed OpenAI prices with the current
-- OpenAI API pricing page (Standard processing, USD per 1M tokens).
UPDATE model_marketplace_items
SET
    official_prices = '{"input":4,"output":20,"cacheWrite":5,"cacheRead":0.4}'::jsonb,
    updated_at = NOW()
WHERE model_name = 'gpt-5.6-sol';

UPDATE model_marketplace_items
SET
    official_prices = '{"input":2,"output":12,"cacheWrite":2.5,"cacheRead":0.2}'::jsonb,
    updated_at = NOW()
WHERE model_name = 'gpt-5.6-terra';

UPDATE model_marketplace_items
SET
    official_prices = '{"input":0.2,"output":1.2,"cacheWrite":0.25,"cacheRead":0.02}'::jsonb,
    updated_at = NOW()
WHERE model_name = 'gpt-5.6-luna';

INSERT INTO model_marketplace_items (
    model_name,
    pricing_aliases,
    vendor_name,
    groups,
    tags,
    endpoints,
    description,
    official_prices,
    sort_order,
    enabled
) VALUES (
    'gpt-6-astra',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Pro"]'::jsonb,
    '[]'::jsonb,
    '["openai"]'::jsonb,
    'GPT-6 Astra',
    '{"input":10,"output":50,"cacheWrite":12.5,"cacheRead":1}'::jsonb,
    44,
    TRUE
)
ON CONFLICT (model_name) DO UPDATE SET
    vendor_name = EXCLUDED.vendor_name,
    official_prices = EXCLUDED.official_prices,
    updated_at = NOW();
