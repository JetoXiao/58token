CREATE TABLE IF NOT EXISTS model_marketplace_items (
    id BIGSERIAL PRIMARY KEY,
    model_name VARCHAR(128) NOT NULL UNIQUE,
    pricing_aliases JSONB NOT NULL DEFAULT '[]'::jsonb,
    vendor_name VARCHAR(64) NOT NULL,
    groups JSONB NOT NULL DEFAULT '[]'::jsonb,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    endpoints JSONB NOT NULL DEFAULT '[]'::jsonb,
    description TEXT NOT NULL DEFAULT '',
    official_prices JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_marketplace_items_enabled_sort
    ON model_marketplace_items (enabled, sort_order, model_name);

INSERT INTO model_marketplace_items (
    model_name,
    pricing_aliases,
    vendor_name,
    groups,
    tags,
    endpoints,
    official_prices,
    sort_order,
    enabled
) VALUES
(
    'claude-opus-4.7',
    '["claude-opus-4-7"]'::jsonb,
    'Anthropic',
    '["Claude Lite","Claude Plus","Claude Max"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","1M"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    '{"input":5,"output":25,"cacheWrite":6.25,"cacheRead":0.5}'::jsonb,
    10,
    TRUE
),
(
    'claude-opus-4.6',
    '["claude-opus-4-6"]'::jsonb,
    'Anthropic',
    '["Claude Lite","Claude Plus","Claude Max"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","1M"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    '{"input":5,"output":25,"cacheWrite":6.25,"cacheRead":0.5}'::jsonb,
    20,
    TRUE
),
(
    'claude-sonnet-4.6',
    '["claude-sonnet-4-6"]'::jsonb,
    'Anthropic',
    '["Claude Lite","Claude Plus","Claude Max"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","1M"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    '{"input":3,"output":15,"cacheWrite":3.75,"cacheRead":0.3}'::jsonb,
    30,
    TRUE
),
(
    'claude-haiku-4.5',
    '["claude-haiku-4-5"]'::jsonb,
    'Anthropic',
    '["Claude Lite","Claude Plus","Claude Max"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","200K"]'::jsonb,
    '["anthropic","openai"]'::jsonb,
    '{"input":1,"output":5,"cacheWrite":1.25,"cacheRead":0.1}'::jsonb,
    40,
    TRUE
),
(
    'gpt-5.5',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Lite","Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","1.1M"]'::jsonb,
    '["openai"]'::jsonb,
    '{"input":[{"label":"<=272K","value":5},{"label":">272K","value":10}],"output":[{"label":"<=272K","value":30},{"label":">272K","value":45}],"cacheWrite":null,"cacheRead":[{"label":"<=272K","value":0.5},{"label":">272K","value":1}]}'::jsonb,
    50,
    TRUE
),
(
    'gpt-5.4',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Lite","Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","400K"]'::jsonb,
    '["openai"]'::jsonb,
    '{"input":[{"label":"<=272K","value":2.5},{"label":">272K","value":5}],"output":[{"label":"<=272K","value":15},{"label":">272K","value":22.5}],"cacheWrite":null,"cacheRead":[{"label":"<=272K","value":0.25},{"label":">272K","value":0.5}]}'::jsonb,
    60,
    TRUE
),
(
    'gpt-5.4-mini',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Lite","Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","400K"]'::jsonb,
    '["openai"]'::jsonb,
    '{"input":0.75,"output":4.5,"cacheWrite":null,"cacheRead":0.075}'::jsonb,
    70,
    TRUE
),
(
    'gpt-5.2',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Lite","Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Files","Vision","400K"]'::jsonb,
    '["openai"]'::jsonb,
    '{"input":1.75,"output":14,"cacheWrite":null,"cacheRead":0.175}'::jsonb,
    80,
    TRUE
),
(
    'gpt-5.3-codex',
    '[]'::jsonb,
    'OpenAI',
    '["Codex Lite","Codex Pro"]'::jsonb,
    '["Reasoning","Tools","Vision","400K"]'::jsonb,
    '["openai"]'::jsonb,
    '{"input":1.75,"output":14,"cacheWrite":null,"cacheRead":0.175}'::jsonb,
    90,
    TRUE
)
ON CONFLICT (model_name) DO NOTHING;
