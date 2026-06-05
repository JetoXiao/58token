-- Store a compact, sanitized request-parameter summary for admin request drilldowns.
-- This intentionally avoids saving full prompts, images, uploaded files, or base64 payloads.

ALTER TABLE usage_logs
  ADD COLUMN IF NOT EXISTS request_params JSONB;

ALTER TABLE ops_error_logs
  ADD COLUMN IF NOT EXISTS request_params JSONB;

COMMENT ON COLUMN usage_logs.request_params IS 'Sanitized request parameter summary for admin request drilldowns.';
COMMENT ON COLUMN ops_error_logs.request_params IS 'Sanitized request parameter summary for admin request drilldowns.';
