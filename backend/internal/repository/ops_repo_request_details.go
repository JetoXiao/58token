package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *opsRepository) ListRequestDetails(ctx context.Context, filter *service.OpsRequestDetailFilter) ([]*service.OpsRequestDetail, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil ops repository")
	}

	page, pageSize, startTime, endTime := filter.Normalize()
	offset := (page - 1) * pageSize

	conditions := make([]string, 0, 16)
	args := make([]any, 0, 24)

	// Placeholders $1/$2 reserved for time window inside the CTE.
	args = append(args, startTime.UTC(), endTime.UTC())

	addCondition := func(condition string, values ...any) {
		conditions = append(conditions, condition)
		args = append(args, values...)
	}

	if filter != nil {
		if kind := strings.TrimSpace(strings.ToLower(filter.Kind)); kind != "" && kind != "all" {
			if kind != string(service.OpsRequestKindSuccess) && kind != string(service.OpsRequestKindError) {
				return nil, 0, fmt.Errorf("invalid kind")
			}
			addCondition(fmt.Sprintf("kind = $%d", len(args)+1), kind)
		}

		if platform := strings.TrimSpace(strings.ToLower(filter.Platform)); platform != "" {
			addCondition(fmt.Sprintf("LOWER(platform) = $%d", len(args)+1), platform)
		}
		if filter.GroupID != nil && *filter.GroupID > 0 {
			addCondition(fmt.Sprintf("group_id = $%d", len(args)+1), *filter.GroupID)
		}

		if filter.UserID != nil && *filter.UserID > 0 {
			addCondition(fmt.Sprintf("user_id = $%d", len(args)+1), *filter.UserID)
		}
		if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
			addCondition(fmt.Sprintf("api_key_id = $%d", len(args)+1), *filter.APIKeyID)
		}
		if filter.AccountID != nil && *filter.AccountID > 0 {
			addCondition(fmt.Sprintf("account_id = $%d", len(args)+1), *filter.AccountID)
		}

		if model := strings.TrimSpace(filter.Model); model != "" {
			addCondition(fmt.Sprintf("model = $%d", len(args)+1), model)
		}
		if requestID := strings.TrimSpace(filter.RequestID); requestID != "" {
			addCondition(fmt.Sprintf("request_id = $%d", len(args)+1), requestID)
		}
		if q := strings.TrimSpace(filter.Query); q != "" {
			like := "%" + strings.ToLower(q) + "%"
			startIdx := len(args) + 1
			addCondition(
				fmt.Sprintf("(LOWER(COALESCE(request_id,'')) LIKE $%d OR LOWER(COALESCE(model,'')) LIKE $%d OR LOWER(COALESCE(message,'')) LIKE $%d OR LOWER(COALESCE(request_params::text,'')) LIKE $%d OR LOWER(COALESCE(user_email,'')) LIKE $%d OR LOWER(COALESCE(username,'')) LIKE $%d OR LOWER(COALESCE(api_key_name,'')) LIKE $%d OR LOWER(COALESCE(account_name,'')) LIKE $%d OR LOWER(COALESCE(group_name,'')) LIKE $%d)",
					startIdx, startIdx+1, startIdx+2, startIdx+3, startIdx+4, startIdx+5, startIdx+6, startIdx+7, startIdx+8,
				),
				like, like, like, like, like, like, like, like, like,
			)
		}

		if filter.MinDurationMs != nil {
			addCondition(fmt.Sprintf("duration_ms >= $%d", len(args)+1), *filter.MinDurationMs)
		}
		if filter.MaxDurationMs != nil {
			addCondition(fmt.Sprintf("duration_ms <= $%d", len(args)+1), *filter.MaxDurationMs)
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	cte := `
WITH combined AS (
  SELECT
    'success'::TEXT AS kind,
    ul.created_at AS created_at,
    ul.request_id AS request_id,
    COALESCE(NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    ul.model AS model,
    ul.duration_ms AS duration_ms,
    ul.first_token_ms AS first_token_ms,
    NULL::INT AS status_code,
    NULL::BIGINT AS error_id,
    NULL::TEXT AS phase,
    NULL::TEXT AS severity,
    NULL::TEXT AS message,
    ul.user_id AS user_id,
    ul.api_key_id AS api_key_id,
    ul.account_id AS account_id,
    ul.group_id AS group_id,
    COALESCE(u.email, '') AS user_email,
    COALESCE(u.username, '') AS username,
    COALESCE(k.name, '') AS api_key_name,
    COALESCE(a.name, '') AS account_name,
    COALESCE(g.name, '') AS group_name,
    ul.stream AS stream,
    ul.openai_ws_mode AS openai_ws_mode,
    ul.request_type AS request_type,
    COALESCE(ul.inbound_endpoint, '') AS inbound_endpoint,
    COALESCE(ul.upstream_endpoint, '') AS upstream_endpoint,
    COALESCE(ul.requested_model, ul.model, '') AS requested_model,
    COALESCE(ul.upstream_model, '') AS upstream_model,
    (
      jsonb_strip_nulls(jsonb_build_object(
        'model', COALESCE(ul.requested_model, ul.model),
        'stream', ul.stream,
        'request_type', CASE
          WHEN ul.request_type = 3 THEN 'ws_v2'
          WHEN ul.request_type = 2 THEN 'stream'
          WHEN ul.request_type = 1 THEN 'sync'
          WHEN ul.openai_ws_mode THEN 'ws_v2'
          WHEN ul.stream THEN 'stream'
          ELSE 'sync'
        END,
        'service_tier', ul.service_tier,
        'reasoning_effort', ul.reasoning_effort,
        'image_count', NULLIF(ul.image_count, 0),
        'image_size', ul.image_size,
        'image_input_size', ul.image_input_size,
        'image_output_size', ul.image_output_size,
        'image_size_source', ul.image_size_source,
        'billing_mode', ul.billing_mode
      )) || COALESCE(ul.request_params, '{}'::jsonb)
    ) AS request_params
  FROM usage_logs ul
  LEFT JOIN users u ON u.id = ul.user_id
  LEFT JOIN api_keys k ON k.id = ul.api_key_id
  LEFT JOIN groups g ON g.id = ul.group_id
  LEFT JOIN accounts a ON a.id = ul.account_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2

  UNION ALL

  SELECT
    'error'::TEXT AS kind,
    o.created_at AS created_at,
    COALESCE(NULLIF(o.request_id,''), NULLIF(o.client_request_id,''), '') AS request_id,
    COALESCE(NULLIF(o.platform, ''), NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    o.model AS model,
    o.duration_ms AS duration_ms,
    o.time_to_first_token_ms AS first_token_ms,
    o.status_code AS status_code,
    o.id AS error_id,
    o.error_phase AS phase,
    o.severity AS severity,
    o.error_message AS message,
    o.user_id AS user_id,
    o.api_key_id AS api_key_id,
    o.account_id AS account_id,
    o.group_id AS group_id,
    COALESCE(u.email, '') AS user_email,
    COALESCE(u.username, '') AS username,
    COALESCE(k.name, '') AS api_key_name,
    COALESCE(a.name, '') AS account_name,
    COALESCE(g.name, '') AS group_name,
    o.stream AS stream,
    FALSE AS openai_ws_mode,
    o.request_type AS request_type,
    COALESCE(o.inbound_endpoint, '') AS inbound_endpoint,
    COALESCE(o.upstream_endpoint, '') AS upstream_endpoint,
    COALESCE(o.requested_model, o.model, '') AS requested_model,
    COALESCE(o.upstream_model, '') AS upstream_model,
    (
      jsonb_strip_nulls(jsonb_build_object(
        'model', COALESCE(o.requested_model, o.model),
        'stream', o.stream,
        'request_type', CASE
          WHEN o.request_type = 3 THEN 'ws_v2'
          WHEN o.request_type = 2 THEN 'stream'
          WHEN o.request_type = 1 THEN 'sync'
          WHEN o.stream THEN 'stream'
          ELSE 'sync'
        END
      )) || COALESCE(o.request_params, '{}'::jsonb)
    ) AS request_params
  FROM ops_error_logs o
  LEFT JOIN users u ON u.id = o.user_id
  LEFT JOIN api_keys k ON k.id = o.api_key_id
  LEFT JOIN groups g ON g.id = o.group_id
  LEFT JOIN accounts a ON a.id = o.account_id
  WHERE o.created_at >= $1 AND o.created_at < $2
    AND COALESCE(o.status_code, 0) >= 400
)
`

	countQuery := fmt.Sprintf(`%s SELECT COUNT(1) FROM combined %s`, cte, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			total = 0
		} else {
			return nil, 0, err
		}
	}

	sort := "ORDER BY created_at DESC"
	if filter != nil {
		switch strings.TrimSpace(strings.ToLower(filter.Sort)) {
		case "", "created_at_desc":
			// default
		case "duration_desc":
			sort = "ORDER BY duration_ms DESC NULLS LAST, created_at DESC"
		case "first_token_desc":
			sort = "ORDER BY first_token_ms DESC NULLS LAST, duration_ms DESC NULLS LAST, created_at DESC"
		default:
			return nil, 0, fmt.Errorf("invalid sort")
		}
	}

	listQuery := fmt.Sprintf(`
%s
SELECT
  kind,
  created_at,
  request_id,
  platform,
  model,
  duration_ms,
  first_token_ms,
  status_code,
  error_id,
  phase,
  severity,
  message,
  user_id,
  api_key_id,
  account_id,
  group_id,
  user_email,
  username,
  api_key_name,
  account_name,
  group_name,
  stream,
  openai_ws_mode,
  request_type,
  inbound_endpoint,
  upstream_endpoint,
  requested_model,
  upstream_model,
  request_params
FROM combined
%s
%s
LIMIT $%d OFFSET $%d
`, cte, where, sort, len(args)+1, len(args)+2)

	listArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	toIntPtr := func(v sql.NullInt64) *int {
		if !v.Valid {
			return nil
		}
		i := int(v.Int64)
		return &i
	}
	toInt64Ptr := func(v sql.NullInt64) *int64 {
		if !v.Valid {
			return nil
		}
		i := v.Int64
		return &i
	}

	out := make([]*service.OpsRequestDetail, 0, pageSize)
	for rows.Next() {
		var (
			kind      string
			createdAt time.Time
			requestID sql.NullString
			platform  sql.NullString
			model     sql.NullString

			durationMs   sql.NullInt64
			firstTokenMs sql.NullInt64
			statusCode   sql.NullInt64
			errorID      sql.NullInt64

			phase    sql.NullString
			severity sql.NullString
			message  sql.NullString

			userID    sql.NullInt64
			apiKeyID  sql.NullInt64
			accountID sql.NullInt64
			groupID   sql.NullInt64

			userEmail   sql.NullString
			username    sql.NullString
			apiKeyName  sql.NullString
			accountName sql.NullString
			groupName   sql.NullString

			stream           bool
			openAIWSMode     bool
			requestTypeRaw   sql.NullInt64
			inboundEndpoint  sql.NullString
			upstreamEndpoint sql.NullString
			requestedModel   sql.NullString
			upstreamModel    sql.NullString
			requestParams    sql.NullString
		)

		if err := rows.Scan(
			&kind,
			&createdAt,
			&requestID,
			&platform,
			&model,
			&durationMs,
			&firstTokenMs,
			&statusCode,
			&errorID,
			&phase,
			&severity,
			&message,
			&userID,
			&apiKeyID,
			&accountID,
			&groupID,
			&userEmail,
			&username,
			&apiKeyName,
			&accountName,
			&groupName,
			&stream,
			&openAIWSMode,
			&requestTypeRaw,
			&inboundEndpoint,
			&upstreamEndpoint,
			&requestedModel,
			&upstreamModel,
			&requestParams,
		); err != nil {
			return nil, 0, err
		}

		requestType := service.RequestTypeUnknown
		if requestTypeRaw.Valid {
			requestType = service.RequestTypeFromInt16(int16(requestTypeRaw.Int64))
		}
		if requestType == service.RequestTypeUnknown {
			requestType = service.RequestTypeFromLegacy(stream, openAIWSMode)
		}

		item := &service.OpsRequestDetail{
			Kind:      service.OpsRequestKind(kind),
			CreatedAt: createdAt,
			RequestID: strings.TrimSpace(requestID.String),
			Platform:  strings.TrimSpace(platform.String),
			Model:     strings.TrimSpace(model.String),

			DurationMs:   toIntPtr(durationMs),
			FirstTokenMs: toIntPtr(firstTokenMs),
			StatusCode:   toIntPtr(statusCode),
			ErrorID:      toInt64Ptr(errorID),
			Phase:        phase.String,
			Severity:     severity.String,
			Message:      message.String,

			UserID:    toInt64Ptr(userID),
			APIKeyID:  toInt64Ptr(apiKeyID),
			AccountID: toInt64Ptr(accountID),
			GroupID:   toInt64Ptr(groupID),

			UserEmail:   strings.TrimSpace(userEmail.String),
			Username:    strings.TrimSpace(username.String),
			APIKeyName:  strings.TrimSpace(apiKeyName.String),
			AccountName: strings.TrimSpace(accountName.String),
			GroupName:   strings.TrimSpace(groupName.String),

			Stream:           stream,
			RequestType:      requestType.String(),
			InboundEndpoint:  strings.TrimSpace(inboundEndpoint.String),
			UpstreamEndpoint: strings.TrimSpace(upstreamEndpoint.String),
			RequestedModel:   strings.TrimSpace(requestedModel.String),
			UpstreamModel:    strings.TrimSpace(upstreamModel.String),
			RequestParams:    requestParamsFromNullJSON(requestParams),
		}

		if item.Platform == "" {
			item.Platform = "unknown"
		}

		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *opsRepository) ListRequestFilterOptions(ctx context.Context, filter *service.OpsRequestDetailFilter) (*service.OpsRequestFilterOptions, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}

	_, _, startTime, endTime := filter.Normalize()
	args := []any{startTime.UTC(), endTime.UTC()}
	kindWhere := ""
	if filter != nil {
		if kind := strings.TrimSpace(strings.ToLower(filter.Kind)); kind != "" && kind != "all" {
			if kind != string(service.OpsRequestKindSuccess) && kind != string(service.OpsRequestKindError) {
				return nil, fmt.Errorf("invalid kind")
			}
			args = append(args, kind)
			kindWhere = "WHERE kind = $3"
		}
	}

	queryWithKind := func(selectSQL string) string {
		return requestFilterOptionsCTE + fmt.Sprintf(selectSQL, kindWhere)
	}

	out := &service.OpsRequestFilterOptions{}
	var err error
	out.Platforms, err = r.scanRequestFilterOptions(ctx, queryWithKind(`
SELECT value, UPPER(value) AS label
FROM (
  SELECT DISTINCT LOWER(NULLIF(TRIM(platform), '')) AS value
  FROM combined
  %s
) v
WHERE value IS NOT NULL
ORDER BY value
LIMIT 200
`), args...)
	if err != nil {
		return nil, err
	}

	out.Models, err = r.scanRequestFilterOptions(ctx, queryWithKind(`
SELECT value, value AS label
FROM (
  SELECT DISTINCT NULLIF(TRIM(model), '') AS value
  FROM combined
  %s
) v
WHERE value IS NOT NULL
ORDER BY value
LIMIT 300
`), args...)
	if err != nil {
		return nil, err
	}

	out.Users, err = r.scanRequestFilterOptions(ctx, queryWithKind(`
SELECT c.user_id::TEXT AS value,
       CASE
         WHEN NULLIF(TRIM(COALESCE(u.email, '')), '') IS NULL THEN '#' || c.user_id::TEXT
         ELSE TRIM(u.email) || ' (#' || c.user_id::TEXT || ')'
       END AS label
FROM (
  SELECT DISTINCT user_id
  FROM combined
  %s
) c
LEFT JOIN users u ON u.id = c.user_id
WHERE c.user_id IS NOT NULL
ORDER BY label
LIMIT 200
`), args...)
	if err != nil {
		return nil, err
	}

	out.APIKeys, err = r.scanRequestFilterOptions(ctx, queryWithKind(`
SELECT c.api_key_id::TEXT AS value,
       CASE
         WHEN NULLIF(TRIM(COALESCE(k.name, '')), '') IS NULL THEN '#' || c.api_key_id::TEXT
         ELSE TRIM(k.name) || ' (#' || c.api_key_id::TEXT || ')'
       END AS label
FROM (
  SELECT DISTINCT api_key_id
  FROM combined
  %s
) c
LEFT JOIN api_keys k ON k.id = c.api_key_id
WHERE c.api_key_id IS NOT NULL
ORDER BY label
LIMIT 200
`), args...)
	if err != nil {
		return nil, err
	}

	out.Accounts, err = r.scanRequestFilterOptions(ctx, queryWithKind(`
SELECT c.account_id::TEXT AS value,
       CASE
         WHEN NULLIF(TRIM(COALESCE(a.name, '')), '') IS NULL THEN '#' || c.account_id::TEXT
         ELSE TRIM(a.name) || ' (#' || c.account_id::TEXT || ')'
       END AS label
FROM (
  SELECT DISTINCT account_id
  FROM combined
  %s
) c
LEFT JOIN accounts a ON a.id = c.account_id
WHERE c.account_id IS NOT NULL
ORDER BY label
LIMIT 200
`), args...)
	if err != nil {
		return nil, err
	}

	out.Groups, err = r.scanRequestFilterOptions(ctx, queryWithKind(`
SELECT c.group_id::TEXT AS value,
       CASE
         WHEN NULLIF(TRIM(COALESCE(g.name, '')), '') IS NULL THEN '#' || c.group_id::TEXT
         ELSE TRIM(g.name) || ' (#' || c.group_id::TEXT || ')'
       END AS label
FROM (
  SELECT DISTINCT group_id
  FROM combined
  %s
) c
LEFT JOIN groups g ON g.id = c.group_id
WHERE c.group_id IS NOT NULL
ORDER BY label
LIMIT 200
`), args...)
	if err != nil {
		return nil, err
	}

	return out, nil
}

const requestFilterOptionsCTE = `
WITH combined AS (
  SELECT
    'success'::TEXT AS kind,
    COALESCE(NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    COALESCE(ul.requested_model, ul.model, '') AS model,
    ul.user_id AS user_id,
    ul.api_key_id AS api_key_id,
    ul.account_id AS account_id,
    ul.group_id AS group_id
  FROM usage_logs ul
  LEFT JOIN groups g ON g.id = ul.group_id
  LEFT JOIN accounts a ON a.id = ul.account_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2

  UNION ALL

  SELECT
    'error'::TEXT AS kind,
    COALESCE(NULLIF(o.platform, ''), NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    COALESCE(o.requested_model, o.model, '') AS model,
    o.user_id AS user_id,
    o.api_key_id AS api_key_id,
    o.account_id AS account_id,
    o.group_id AS group_id
  FROM ops_error_logs o
  LEFT JOIN groups g ON g.id = o.group_id
  LEFT JOIN accounts a ON a.id = o.account_id
  WHERE o.created_at >= $1 AND o.created_at < $2
    AND COALESCE(o.status_code, 0) >= 400
)
`

func (r *opsRepository) scanRequestFilterOptions(ctx context.Context, query string, args ...any) ([]service.OpsRequestFilterOption, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]service.OpsRequestFilterOption, 0, 32)
	for rows.Next() {
		var value, label sql.NullString
		if err := rows.Scan(&value, &label); err != nil {
			return nil, err
		}
		if !value.Valid {
			continue
		}
		v := strings.TrimSpace(value.String)
		if v == "" {
			continue
		}
		l := strings.TrimSpace(label.String)
		if l == "" {
			l = v
		}
		out = append(out, service.OpsRequestFilterOption{
			Value: v,
			Label: l,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func requestParamsFromNullJSON(v sql.NullString) map[string]any {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(v.String), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
