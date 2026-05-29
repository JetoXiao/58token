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
    '["Reasoning","Tools","Files","Vision","Computer Use","Adaptive Thinking","Fast Mode","1M","128K"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    'Claude Opus 4.8 is Anthropic''s latest Opus model for complex reasoning, agentic coding, tool use, computer use, vision/PDF input, prompt caching, and long-context workflows. It supports a 1M input context window and up to 128K output tokens.',
    '{
        "input": [
            {"label":"standard","value":5},
            {"label":"fast","value":10},
            {"label":"batch","value":2.5}
        ],
        "output": [
            {"label":"standard","value":25},
            {"label":"fast","value":50},
            {"label":"batch","value":12.5}
        ],
        "cacheWrite": [
            {"label":"standard 5m","value":6.25},
            {"label":"standard 1h","value":10},
            {"label":"fast 5m","value":12.5},
            {"label":"fast 1h","value":20}
        ],
        "cacheRead": [
            {"label":"standard","value":0.5},
            {"label":"fast","value":1}
        ]
    }'::jsonb,
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
