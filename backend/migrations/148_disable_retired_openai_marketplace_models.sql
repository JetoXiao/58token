UPDATE model_marketplace_items
SET
    enabled = FALSE,
    updated_at = NOW()
WHERE model_name IN ('gpt-5.2', 'gpt-5.3-codex');
