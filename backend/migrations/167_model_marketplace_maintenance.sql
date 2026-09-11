CREATE TABLE IF NOT EXISTS model_marketplace_group_configs (
    id BIGSERIAL PRIMARY KEY,
    source_group_id BIGINT NOT NULL UNIQUE,
    source_group_name VARCHAR(255) NOT NULL DEFAULT '',
    source_platform VARCHAR(64) NOT NULL DEFAULT '',
    source_rate_multiplier NUMERIC(20, 8) NOT NULL DEFAULT 1,
    source_sort_order INTEGER NOT NULL DEFAULT 0,
    source_models JSONB NOT NULL DEFAULT '[]'::jsonb,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    display_multiplier NUMERIC(20, 8) NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    source_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_marketplace_group_configs_active_sort
    ON model_marketplace_group_configs (source_active, enabled, source_sort_order, source_group_id);

CREATE TABLE IF NOT EXISTS model_marketplace_model_configs (
    id BIGSERIAL PRIMARY KEY,
    group_config_id BIGINT NOT NULL REFERENCES model_marketplace_group_configs(id) ON DELETE CASCADE,
    source_model_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_config_id, source_model_name)
);

CREATE INDEX IF NOT EXISTS idx_model_marketplace_model_configs_enabled_sort
    ON model_marketplace_model_configs (group_config_id, enabled, sort_order, source_model_name);

CREATE TABLE IF NOT EXISTS model_marketplace_sync_state (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE),
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
