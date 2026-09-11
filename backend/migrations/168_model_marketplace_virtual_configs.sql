-- Allow marketplace-only groups and models that are independent from real gateway data.
ALTER TABLE model_marketplace_group_configs
    ALTER COLUMN source_group_id DROP NOT NULL;

ALTER TABLE model_marketplace_group_configs
    ADD COLUMN IF NOT EXISTS is_virtual BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE model_marketplace_model_configs
    ADD COLUMN IF NOT EXISTS is_virtual BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_model_marketplace_group_configs_virtual_sort
    ON model_marketplace_group_configs (is_virtual, source_sort_order, id);

CREATE INDEX IF NOT EXISTS idx_model_marketplace_model_configs_virtual_sort
    ON model_marketplace_model_configs (group_config_id, is_virtual, sort_order, source_model_name);
