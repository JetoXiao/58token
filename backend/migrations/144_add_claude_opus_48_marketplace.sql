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
    'claude-opus-4.8',
    '["claude-opus-4-8"]'::jsonb,
    'Anthropic',
    '["Claude Lite","Claude Plus","Claude Max"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","Computer Use","Adaptive Thinking","1M","128K"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    'Claude Opus 4.8 is Anthropic''s latest Opus model for complex reasoning, agentic coding, tool use, computer use, vision/PDF input, prompt caching, and long-context workflows. It supports a 1M input context window and up to 128K output tokens.',
    '{"input":5,"output":25,"cacheWrite":6.25,"cacheRead":0.5}'::jsonb,
    5,
    TRUE
)
ON CONFLICT (model_name) DO UPDATE SET
    pricing_aliases = EXCLUDED.pricing_aliases,
    vendor_name = EXCLUDED.vendor_name,
    groups = EXCLUDED.groups,
    tags = EXCLUDED.tags,
    endpoints = EXCLUDED.endpoints,
    description = EXCLUDED.description,
    official_prices = EXCLUDED.official_prices,
    sort_order = EXCLUDED.sort_order,
    enabled = EXCLUDED.enabled,
    updated_at = NOW();

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
