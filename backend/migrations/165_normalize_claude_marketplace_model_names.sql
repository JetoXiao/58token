-- The API-compatible Claude model ID uses hyphens between the version parts.
-- Keep marketplace canonical names identical to the values users should paste
-- into Claude clients. The existing pricing aliases remain compatible.
DO $$
DECLARE
    dotted_name TEXT;
    canonical_name TEXT;
BEGIN
    FOR dotted_name, canonical_name IN
        SELECT * FROM (VALUES
            ('claude-opus-4.8', 'claude-opus-4-8'),
            ('claude-opus-4.7', 'claude-opus-4-7'),
            ('claude-opus-4.6', 'claude-opus-4-6'),
            ('claude-sonnet-4.6', 'claude-sonnet-4-6'),
            ('claude-haiku-4.5', 'claude-haiku-4-5')
        ) AS model_names(dotted_name, canonical_name)
    LOOP
        IF EXISTS (SELECT 1 FROM model_marketplace_items WHERE model_name = dotted_name)
           AND NOT EXISTS (SELECT 1 FROM model_marketplace_items WHERE model_name = canonical_name) THEN
            UPDATE model_marketplace_items
            SET model_name = canonical_name, updated_at = NOW()
            WHERE model_name = dotted_name;
        END IF;
    END LOOP;
END $$;
