INSERT INTO settings (key, value, updated_at)
VALUES (
    'affiliate_partner_tiers',
    '[{"level":"spark","name":"Spark","rebate_rate_percent":40,"required_invitees":10},{"level":"voyage","name":"Voyage","rebate_rate_percent":50,"required_invitees":30},{"level":"summit","name":"Summit","rebate_rate_percent":60,"required_invitees":50},{"level":"cocreate","name":"Co-create","rebate_rate_percent":70,"required_invitees":100}]',
    NOW()
)
ON CONFLICT (key) DO NOTHING;
