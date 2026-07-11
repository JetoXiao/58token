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
) VALUES
(
    'claude-sonnet-5',
    '[]'::jsonb,
    'Anthropic',
    '["Claude Lite","Claude Max"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","1M","128K"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    'Claude Sonnet 5 supports long-context reasoning, coding agents, tools, vision, files, and prompt caching. Prices use Anthropic''s introductory rates through August 31, 2026.',
    '{
        "input": 2,
        "output": 10,
        "cacheWrite": [
            {"label":"5m","value":2.5},
            {"label":"1h","value":4}
        ],
        "cacheRead": 0.2
    }'::jsonb,
    2,
    TRUE
),
(
    'gpt-5.6-sol',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","Web Search","1.05M","128K"]'::jsonb,
    '["openai"]'::jsonb,
    'GPT-5.6 Sol supports reasoning, tools, files, vision, web search, and prompt caching.',
    '{"input":5,"output":30,"cacheWrite":6.25,"cacheRead":0.5}'::jsonb,
    45,
    TRUE
),
(
    'gpt-5.6-terra',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","Web Search","1.05M","128K"]'::jsonb,
    '["openai"]'::jsonb,
    'GPT-5.6 Terra supports reasoning, tools, files, vision, web search, and prompt caching.',
    '{"input":2.5,"output":15,"cacheWrite":3.125,"cacheRead":0.25}'::jsonb,
    46,
    TRUE
),
(
    'gpt-5.6-luna',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","Web Search","1.05M","128K"]'::jsonb,
    '["openai"]'::jsonb,
    'GPT-5.6 Luna supports reasoning, tools, files, vision, web search, and prompt caching.',
    '{"input":1,"output":6,"cacheWrite":1.25,"cacheRead":0.1}'::jsonb,
    47,
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
