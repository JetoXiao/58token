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
    'claude-fable-5',
    '[]'::jsonb,
    'Anthropic',
    '["Claude Max"]'::jsonb,
    '["Reasoning","Tools","Vision","Adaptive Thinking","1M","128K"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    'Claude Fable 5 is Anthropic''s most capable widely released model for demanding reasoning and long-horizon agentic work. It supports a 1M context window, up to 128K output tokens, adaptive thinking, tools, and vision.',
    '{
        "input": 10,
        "output": 50,
        "cacheWrite": [
            {"label":"5m","value":12.5},
            {"label":"1h","value":20}
        ],
        "cacheRead": 1
    }'::jsonb,
    1,
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
