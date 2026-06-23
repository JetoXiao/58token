-- Remove stored user-content previews from request drilldown parameters.
-- Keep non-content diagnostics such as model, stream, sizes, cache status,
-- routing latency, and scheduler decisions.

UPDATE usage_logs
SET request_params = NULLIF(
  request_params
    - 'prompt_preview'
    - 'input_preview'
    - 'last_user_message_preview'
    - 'last_input_preview'
    - 'last_user_content_preview'
    - '_encrypted_sensitive_request_params',
  '{}'::jsonb
)
WHERE request_params ?| ARRAY[
  'prompt_preview',
  'input_preview',
  'last_user_message_preview',
  'last_input_preview',
  'last_user_content_preview',
  '_encrypted_sensitive_request_params'
];

UPDATE ops_error_logs
SET request_params = NULLIF(
  request_params
    - 'prompt_preview'
    - 'input_preview'
    - 'last_user_message_preview'
    - 'last_input_preview'
    - 'last_user_content_preview'
    - '_encrypted_sensitive_request_params',
  '{}'::jsonb
)
WHERE request_params ?| ARRAY[
  'prompt_preview',
  'input_preview',
  'last_user_message_preview',
  'last_input_preview',
  'last_user_content_preview',
  '_encrypted_sensitive_request_params'
];

COMMENT ON COLUMN usage_logs.request_params IS 'Non-content request parameter diagnostics for admin request drilldowns. User prompt/input text is not stored.';
COMMENT ON COLUMN ops_error_logs.request_params IS 'Non-content request parameter diagnostics for admin request drilldowns. User prompt/input text is not stored.';
