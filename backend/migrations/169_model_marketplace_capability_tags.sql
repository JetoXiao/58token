ALTER TABLE model_marketplace_model_configs
    ADD COLUMN IF NOT EXISTS capability_tags JSONB;
