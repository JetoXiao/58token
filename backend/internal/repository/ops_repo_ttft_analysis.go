package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *opsRepository) GetTTFTAnalysis(ctx context.Context, filter *service.OpsTTFTAnalysisFilter) (*service.OpsTTFTAnalysisResponse, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil {
		return nil, fmt.Errorf("nil filter")
	}

	start := filter.StartTime.UTC()
	end := filter.EndTime.UTC()
	threshold := filter.SlowThresholdMs
	if threshold <= 0 {
		threshold = 1000
	}
	limit := filter.TopLimit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	summary, err := r.queryTTFTSummary(ctx, filter, start, end, threshold)
	if err != nil {
		return nil, err
	}
	reasons, err := r.queryTTFTReasons(ctx, filter, start, end, threshold, limit)
	if err != nil {
		return nil, err
	}
	topModels, err := r.queryTTFTTop(ctx, filter, start, end, threshold, limit, "model")
	if err != nil {
		return nil, err
	}
	topAccounts, err := r.queryTTFTTop(ctx, filter, start, end, threshold, limit, "account")
	if err != nil {
		return nil, err
	}
	topGroups, err := r.queryTTFTTop(ctx, filter, start, end, threshold, limit, "group")
	if err != nil {
		return nil, err
	}
	topAPIKeys, err := r.queryTTFTTop(ctx, filter, start, end, threshold, limit, "api_key")
	if err != nil {
		return nil, err
	}
	accountParticipation, err := r.queryTTFTAccountParticipation(ctx, filter, start, end, threshold, limit)
	if err != nil {
		return nil, err
	}
	slowRequests, err := r.queryTTFTSlowRequests(ctx, filter, start, end, threshold, limit)
	if err != nil {
		return nil, err
	}

	return &service.OpsTTFTAnalysisResponse{
		StartTime:            start,
		EndTime:              end,
		Platform:             strings.TrimSpace(filter.Platform),
		GroupID:              filter.GroupID,
		SlowThresholdMs:      threshold,
		Summary:              summary,
		Reasons:              reasons,
		TopModels:            topModels,
		TopAccounts:          topAccounts,
		TopGroups:            topGroups,
		TopAPIKeys:           topAPIKeys,
		AccountParticipation: accountParticipation,
		SlowRequests:         slowRequests,
		Recommendations:      buildTTFTRecommendations(summary, reasons, topModels, topAccounts, topGroups, topAPIKeys),
	}, nil
}

func ttftBaseUsageSQL(filter *service.OpsTTFTAnalysisFilter, start, end time.Time, startIndex int) (string, []any) {
	dashboardFilter := &service.OpsDashboardFilter{
		StartTime: start,
		EndTime:   end,
		Platform:  filter.Platform,
		GroupID:   filter.GroupID,
	}
	join, where, args, next := buildUsageWhere(dashboardFilter, start, end, startIndex)
	if !strings.Contains(join, "accounts a") {
		join += " LEFT JOIN accounts a ON a.id = ul.account_id"
	}
	if !strings.Contains(join, "groups g") {
		join += " LEFT JOIN groups g ON g.id = ul.group_id"
	}

	queueWaitExpr := ttftJSONIntExpr("openai_ws_queue_wait_ms")
	connPickExpr := ttftJSONIntExpr("openai_ws_conn_pick_ms")
	routingExpr := ttftJSONIntExpr("routing_latency_ms")
	authExpr := ttftJSONIntExpr("auth_latency_ms")
	upstreamExpr := ttftJSONIntExpr("upstream_latency_ms")
	preUpstreamExpr := ttftNullableIntSumExpr(authExpr, routingExpr, queueWaitExpr, connPickExpr)

	base := fmt.Sprintf(`
WITH base AS (
  SELECT
    ul.created_at,
    ul.request_id,
    COALESCE(NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    COALESCE(ul.model, '') AS model,
    ul.user_id,
    ul.api_key_id,
    ul.account_id,
    ul.group_id,
    COALESCE(u.email, '') AS user_email,
    COALESCE(k.name, '') AS api_key_name,
    COALESCE(a.name, '') AS account_name,
    COALESCE(g.name, '') AS group_name,
    ul.request_type,
    ul.stream,
    ul.openai_ws_mode,
    ul.duration_ms,
    ul.first_token_ms,
    COALESCE(ul.request_params, '{}'::jsonb) AS request_params,
    COALESCE(NULLIF(ul.request_params->>'route_source', ''), CASE
      WHEN lower(COALESCE(a.type, '')) IN ('api','apikey','api_key','upstream','bedrock','service_account','openai-api','openai_api') THEN 'upstream'
      ELSE 'own_pool'
    END) AS route_source,
    NULLIF(ul.request_params->>'response_cache_status', '') AS cache_status,
    CASE
      WHEN NULLIF(ul.request_params->>'ttft_slow_reason', '') = 'response_flush_slow' THEN 'upstream_ttft_slow'
      WHEN NULLIF(ul.request_params->>'ttft_slow_reason', '') IS NOT NULL THEN NULLIF(ul.request_params->>'ttft_slow_reason', '')
      WHEN ul.first_token_ms IS NULL THEN 'unknown'
      WHEN NULLIF(ul.request_params->>'response_cache_status', '') = 'hit' THEN 'cache_hit'
      WHEN ul.first_token_ms < $%d THEN 'normal'
      WHEN COALESCE(%s, 0) >= 300
        AND COALESCE(%s, 0) * 100.0 / GREATEST(ul.first_token_ms, 1) >= 35 THEN 'account_queue_slow'
      WHEN COALESCE(%s, 0) >= 200
        AND COALESCE(%s, 0) * 100.0 / GREATEST(ul.first_token_ms, 1) >= 25 THEN 'connection_pick_slow'
      WHEN COALESCE(%s, 0) >= 200
        AND COALESCE(%s, 0) * 100.0 / GREATEST(ul.first_token_ms, 1) >= 25 THEN 'routing_slow'
      WHEN COALESCE(%s, 0) >= 500
        AND COALESCE(%s, 0) * 100.0 / GREATEST(ul.first_token_ms, 1) >= 60 THEN 'upstream_ttft_slow'
      ELSE 'upstream_ttft_slow'
    END AS slow_reason,
    NULLIF(ul.request_params->>'ttft_slow_reason_detail', '') AS slow_detail,
    %s AS auth_latency_ms,
    %s AS routing_latency_ms,
    %s AS upstream_latency_ms,
    %s AS response_latency_ms,
    %s AS queue_wait_ms,
    %s AS conn_pick_ms,
    %s AS pre_upstream_latency_ms,
    %s AS scheduler_candidate_count,
    %s AS scheduler_top_k,
    %s AS scheduler_latency_ms,
    NULLIF(ul.request_params->>'scheduler_layer', '') AS scheduler_layer,
    NULLIF(ul.request_params->>'scheduler_reason', '') AS scheduler_reason,
    %s AS scheduler_load_skew
  FROM usage_logs ul
  %s
  LEFT JOIN users u ON u.id = ul.user_id
  LEFT JOIN api_keys k ON k.id = ul.api_key_id
  %s
)
`,
		next,
		queueWaitExpr, queueWaitExpr,
		connPickExpr, connPickExpr,
		routingExpr, routingExpr,
		upstreamExpr, upstreamExpr,
		authExpr,
		routingExpr,
		upstreamExpr,
		ttftJSONIntExpr("response_latency_ms"),
		queueWaitExpr,
		connPickExpr,
		preUpstreamExpr,
		ttftJSONIntExpr("scheduler_candidate_count"),
		ttftJSONIntExpr("scheduler_top_k"),
		ttftJSONIntExpr("scheduler_latency_ms"),
		ttftJSONFloatExpr("scheduler_load_skew"),
		join,
		where,
	)
	args = append(args, thresholdPlaceholder{})
	return base, args
}

type thresholdPlaceholder struct{}

func ttftJSONIntExpr(key string) string {
	key = strings.ReplaceAll(strings.TrimSpace(key), "'", "''")
	return fmt.Sprintf("CASE WHEN (ul.request_params->>'%s') ~ '^-?[0-9]+$' THEN (ul.request_params->>'%s')::int END", key, key)
}

func ttftJSONFloatExpr(key string) string {
	key = strings.ReplaceAll(strings.TrimSpace(key), "'", "''")
	return fmt.Sprintf("CASE WHEN (ul.request_params->>'%s') ~ '^-?[0-9]+(\\.[0-9]+)?$' THEN (ul.request_params->>'%s')::double precision END", key, key)
}

func ttftNullableIntSumExpr(exprs ...string) string {
	if len(exprs) == 0 {
		return "NULL"
	}
	nullChecks := make([]string, 0, len(exprs))
	sums := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			continue
		}
		nullChecks = append(nullChecks, fmt.Sprintf("(%s) IS NULL", expr))
		sums = append(sums, fmt.Sprintf("COALESCE((%s), 0)", expr))
	}
	if len(sums) == 0 {
		return "NULL"
	}
	return fmt.Sprintf("CASE WHEN %s THEN NULL ELSE %s END", strings.Join(nullChecks, " AND "), strings.Join(sums, " + "))
}

func ttftBaseSQLAndArgs(filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int) (string, []any) {
	base, rawArgs := ttftBaseUsageSQL(filter, start, end, 1)
	args := make([]any, 0, len(rawArgs))
	for _, arg := range rawArgs {
		if _, ok := arg.(thresholdPlaceholder); ok {
			args = append(args, threshold)
			continue
		}
		args = append(args, arg)
	}
	return base, args
}

func (r *opsRepository) queryTTFTSummary(ctx context.Context, filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int) (service.OpsTTFTSummary, error) {
	base, args := ttftBaseSQLAndArgs(filter, start, end, threshold)
	q := base + `
SELECT
  COUNT(*)::bigint,
  COUNT(first_token_ms)::bigint,
  COUNT(*) FILTER (WHERE first_token_ms >= $` + fmt.Sprint(len(args)+1) + `)::bigint,
  COUNT(*) FILTER (WHERE cache_status = 'hit')::bigint,
  percentile_cont(0.50) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  percentile_cont(0.90) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  percentile_cont(0.99) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  AVG(first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  MAX(first_token_ms),
  percentile_cont(0.50) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL),
  percentile_cont(0.90) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL),
  percentile_cont(0.99) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL),
  AVG(duration_ms) FILTER (WHERE duration_ms IS NOT NULL),
  MAX(duration_ms),
  percentile_cont(0.50) WITHIN GROUP (ORDER BY pre_upstream_latency_ms) FILTER (WHERE pre_upstream_latency_ms IS NOT NULL),
  percentile_cont(0.90) WITHIN GROUP (ORDER BY pre_upstream_latency_ms) FILTER (WHERE pre_upstream_latency_ms IS NOT NULL),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY pre_upstream_latency_ms) FILTER (WHERE pre_upstream_latency_ms IS NOT NULL),
  percentile_cont(0.99) WITHIN GROUP (ORDER BY pre_upstream_latency_ms) FILTER (WHERE pre_upstream_latency_ms IS NOT NULL),
  AVG(pre_upstream_latency_ms) FILTER (WHERE pre_upstream_latency_ms IS NOT NULL),
  MAX(pre_upstream_latency_ms)
FROM base`
	args = append(args, threshold)

	var out service.OpsTTFTSummary
	var tP50, tP90, tP95, tP99, tAvg sql.NullFloat64
	var tMax sql.NullInt64
	var dP50, dP90, dP95, dP99, dAvg sql.NullFloat64
	var dMax sql.NullInt64
	var pP50, pP90, pP95, pP99, pAvg sql.NullFloat64
	var pMax sql.NullInt64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(
		&out.RequestCount,
		&out.FirstTokenSampleCount,
		&out.SlowRequestCount,
		&out.CacheHitCount,
		&tP50, &tP90, &tP95, &tP99, &tAvg, &tMax,
		&dP50, &dP90, &dP95, &dP99, &dAvg, &dMax,
		&pP50, &pP90, &pP95, &pP99, &pAvg, &pMax,
	); err != nil {
		return out, err
	}
	out.SlowRate = safeRatio(out.SlowRequestCount, out.FirstTokenSampleCount)
	out.CacheHitRate = safeRatio(out.CacheHitCount, out.RequestCount)
	out.TTFT = service.OpsPercentiles{
		P50: floatToIntPtr(tP50),
		P90: floatToIntPtr(tP90),
		P95: floatToIntPtr(tP95),
		P99: floatToIntPtr(tP99),
		Avg: floatToIntPtr(tAvg),
	}
	if tMax.Valid {
		v := int(tMax.Int64)
		out.TTFT.Max = &v
	}
	out.Duration = service.OpsPercentiles{
		P50: floatToIntPtr(dP50),
		P90: floatToIntPtr(dP90),
		P95: floatToIntPtr(dP95),
		P99: floatToIntPtr(dP99),
		Avg: floatToIntPtr(dAvg),
	}
	if dMax.Valid {
		v := int(dMax.Int64)
		out.Duration.Max = &v
	}
	out.PreUpstream = service.OpsPercentiles{
		P50: floatToIntPtr(pP50),
		P90: floatToIntPtr(pP90),
		P95: floatToIntPtr(pP95),
		P99: floatToIntPtr(pP99),
		Avg: floatToIntPtr(pAvg),
	}
	if pMax.Valid {
		v := int(pMax.Int64)
		out.PreUpstream.Max = &v
	}

	routeSources, err := r.queryTTFTRouteSources(ctx, filter, start, end, threshold)
	if err != nil {
		return out, err
	}
	routeCounts := make(map[string]int, len(routeSources))
	for _, item := range routeSources {
		if item == nil {
			continue
		}
		routeCounts[item.Source] = int(item.Count)
	}
	out.ByRouteSource = routeCounts
	out.RouteSources = routeSources
	return out, nil
}

func (r *opsRepository) queryTTFTRouteSources(ctx context.Context, filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int) ([]*service.OpsTTFTRouteSourceItem, error) {
	base, args := ttftBaseSQLAndArgs(filter, start, end, threshold)
	q := base + `
SELECT
  COALESCE(NULLIF(route_source, ''), 'unknown'),
  COUNT(*)::bigint,
  COUNT(*) FILTER (WHERE first_token_ms >= $` + fmt.Sprint(len(args)+1) + `)::bigint
FROM base
GROUP BY 1
ORDER BY 2 DESC`
	args = append(args, threshold)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []*service.OpsTTFTRouteSourceItem{}
	var total int64
	for rows.Next() {
		item := &service.OpsTTFTRouteSourceItem{}
		if err := rows.Scan(&item.Source, &item.Count, &item.SlowCount); err != nil {
			return nil, err
		}
		item.SlowRate = safeRatio(item.SlowCount, item.Count)
		total += item.Count
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, item := range out {
		item.Share = safeRatio(item.Count, total)
	}
	return out, nil
}

func (r *opsRepository) queryTTFTReasons(ctx context.Context, filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int, limit int) ([]*service.OpsTTFTReasonItem, error) {
	base, args := ttftBaseSQLAndArgs(filter, start, end, threshold)
	q := base + `
SELECT
  slow_reason,
  COUNT(*)::bigint,
  AVG(first_token_ms),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL)
FROM base
GROUP BY slow_reason
ORDER BY COUNT(*) DESC, slow_reason ASC
LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var total int64
	items := make([]*service.OpsTTFTReasonItem, 0, limit)
	for rows.Next() {
		item := &service.OpsTTFTReasonItem{}
		var avg sql.NullFloat64
		var p95 sql.NullFloat64
		if err := rows.Scan(&item.Reason, &item.Count, &avg, &p95); err != nil {
			return nil, err
		}
		item.AvgTTFTMs = floatToIntPtr(avg)
		item.P95TTFTMs = floatToIntPtr(p95)
		item.Suggestion = ttftReasonSuggestion(item.Reason)
		total += item.Count
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, item := range items {
		item.Share = safeRatio(item.Count, total)
	}
	return items, nil
}

func (r *opsRepository) queryTTFTTop(ctx context.Context, filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int, limit int, dimension string) ([]*service.OpsTTFTTopItem, error) {
	base, args := ttftBaseSQLAndArgs(filter, start, end, threshold)
	keyExpr, labelExpr, routeExpr := ttftTopDimensionExpr(dimension)
	if keyExpr == "" {
		return []*service.OpsTTFTTopItem{}, nil
	}
	q := base + fmt.Sprintf(`
SELECT
  %s AS key,
  %s AS label,
  COUNT(*)::bigint AS total,
  COUNT(*) FILTER (WHERE first_token_ms >= $%d)::bigint AS slow_count,
  AVG(first_token_ms),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  percentile_cont(0.99) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL),
  MAX(first_token_ms),
  %s AS route_hint
FROM base
WHERE first_token_ms IS NOT NULL
GROUP BY 1, 2
ORDER BY slow_count DESC, MAX(first_token_ms) DESC NULLS LAST, total DESC
LIMIT $%d`, keyExpr, labelExpr, len(args)+1, routeExpr, len(args)+2)
	args = append(args, threshold, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*service.OpsTTFTTopItem, 0, limit)
	for rows.Next() {
		item := &service.OpsTTFTTopItem{}
		var avg sql.NullFloat64
		var p95 sql.NullFloat64
		var p99 sql.NullFloat64
		var max sql.NullInt64
		if err := rows.Scan(&item.Key, &item.Label, &item.Count, &item.SlowCount, &avg, &p95, &p99, &max, &item.RouteHint); err != nil {
			return nil, err
		}
		item.SlowRate = safeRatio(item.SlowCount, item.Count)
		item.AvgTTFTMs = floatToIntPtr(avg)
		item.P95TTFTMs = floatToIntPtr(p95)
		item.P99TTFTMs = floatToIntPtr(p99)
		if max.Valid {
			v := int(max.Int64)
			item.MaxTTFTMs = &v
		}
		item.Suggestion = ttftTopSuggestion(dimension, item)
		out = append(out, item)
	}
	return out, rows.Err()
}

func ttftTopDimensionExpr(dimension string) (keyExpr string, labelExpr string, routeExpr string) {
	switch dimension {
	case "model":
		return "COALESCE(NULLIF(model, ''), 'unknown')", "COALESCE(NULLIF(model, ''), 'unknown')", "MAX(route_source)"
	case "account":
		return "COALESCE(account_id::text, 'unknown')", "COALESCE(NULLIF(account_name, ''), account_id::text, 'unknown')", "MAX(route_source)"
	case "group":
		return "COALESCE(group_id::text, 'unknown')", "COALESCE(NULLIF(group_name, ''), group_id::text, 'unknown')", "MAX(route_source)"
	case "api_key":
		return "COALESCE(api_key_id::text, 'unknown')", "COALESCE(NULLIF(api_key_name, ''), api_key_id::text, 'unknown')", "MAX(route_source)"
	default:
		return "", "", ""
	}
}

func (r *opsRepository) queryTTFTAccountParticipation(ctx context.Context, filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int, limit int) ([]*service.OpsTTFTAccountParticipationItem, error) {
	base, args := ttftBaseSQLAndArgs(filter, start, end, threshold)
	accountLimit := limit * 3
	if accountLimit < 10 {
		accountLimit = 10
	}
	if accountLimit > 50 {
		accountLimit = 50
	}
	q := base + `
, observed_groups AS (
  SELECT DISTINCT group_id FROM base WHERE group_id IS NOT NULL
),
observed_platforms AS (
  SELECT DISTINCT lower(platform) AS platform FROM base WHERE platform <> ''
),
usage_stats AS (
  SELECT
    account_id,
    COUNT(*)::bigint AS request_count,
    COUNT(*) FILTER (WHERE first_token_ms >= $` + fmt.Sprint(len(args)+1) + `)::bigint AS slow_count,
    AVG(first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL) AS avg_ttft_ms,
    percentile_cont(0.95) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE first_token_ms IS NOT NULL) AS p95_ttft_ms,
    MAX(first_token_ms) AS max_ttft_ms
  FROM base
  WHERE account_id IS NOT NULL
  GROUP BY account_id
)
SELECT
  a.id,
  COALESCE(a.name, ''),
  COALESCE(a.type, ''),
  COALESCE(a.platform, ''),
  COALESCE(a.status, ''),
  COALESCE(a.schedulable, false),
  COALESCE(scope.in_observed_group, false),
  COALESCE(scope.observed_group_ids_json, '[]'),
  COALESCE(us.request_count, 0)::bigint,
  COALESCE(us.slow_count, 0)::bigint,
  us.avg_ttft_ms,
  us.p95_ttft_ms,
  us.max_ttft_ms
FROM accounts a
LEFT JOIN usage_stats us ON us.account_id = a.id
LEFT JOIN LATERAL (
  SELECT
    COALESCE(bool_or(og.group_id IS NOT NULL), false) AS in_observed_group,
    COALESCE(jsonb_agg(DISTINCT ag.group_id) FILTER (WHERE og.group_id IS NOT NULL), '[]'::jsonb)::text AS observed_group_ids_json
  FROM account_groups ag
  LEFT JOIN observed_groups og ON og.group_id = ag.group_id
  WHERE ag.account_id = a.id
) scope ON true
WHERE a.deleted_at IS NULL
  AND EXISTS (SELECT 1 FROM observed_platforms)
  AND lower(COALESCE(a.platform, '')) IN (SELECT platform FROM observed_platforms)
  AND lower(COALESCE(a.type, '')) IN ('api', 'apikey', 'api_key', 'upstream', 'openai-api', 'openai_api')
ORDER BY
  CASE WHEN COALESCE(scope.in_observed_group, false) THEN 0 ELSE 1 END,
  CASE WHEN COALESCE(us.request_count, 0) > 0 THEN 0 ELSE 1 END,
  COALESCE(us.slow_count, 0) DESC,
  COALESCE(us.request_count, 0) DESC,
  a.priority ASC,
  a.id ASC
LIMIT $` + fmt.Sprint(len(args)+2)
	args = append(args, threshold, accountLimit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*service.OpsTTFTAccountParticipationItem, 0, accountLimit)
	for rows.Next() {
		item := &service.OpsTTFTAccountParticipationItem{}
		var groupIDsRaw string
		var avg, p95 sql.NullFloat64
		var max sql.NullInt64
		if err := rows.Scan(
			&item.AccountID,
			&item.AccountName,
			&item.AccountType,
			&item.Platform,
			&item.Status,
			&item.Schedulable,
			&item.InObservedGroup,
			&groupIDsRaw,
			&item.RequestCount,
			&item.SlowCount,
			&avg,
			&p95,
			&max,
		); err != nil {
			return nil, err
		}
		item.ObservedGroupIDs = decodeJSONInt64Slice(groupIDsRaw)
		item.SlowRate = safeRatio(item.SlowCount, item.RequestCount)
		item.AvgTTFTMs = floatToIntPtr(avg)
		item.P95TTFTMs = floatToIntPtr(p95)
		if max.Valid {
			v := int(max.Int64)
			item.MaxTTFTMs = &v
		}
		item.ParticipationReason = ttftAccountParticipationReason(item)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *opsRepository) queryTTFTSlowRequests(ctx context.Context, filter *service.OpsTTFTAnalysisFilter, start, end time.Time, threshold int, limit int) ([]*service.OpsTTFTSlowRequest, error) {
	base, args := ttftBaseSQLAndArgs(filter, start, end, threshold)
	q := base + `
SELECT
  created_at,
  COALESCE(request_id, ''),
  COALESCE(platform, ''),
  COALESCE(model, ''),
  user_id,
  api_key_id,
  account_id,
  group_id,
  COALESCE(user_email, ''),
  COALESCE(api_key_name, ''),
  COALESCE(account_name, ''),
  COALESCE(group_name, ''),
  request_type,
  stream,
  openai_ws_mode,
  COALESCE(route_source, 'unknown'),
  COALESCE(cache_status, ''),
  COALESCE(slow_reason, 'unknown'),
  COALESCE(slow_detail, ''),
  first_token_ms,
  duration_ms,
  auth_latency_ms,
  routing_latency_ms,
  upstream_latency_ms,
  response_latency_ms,
  queue_wait_ms,
  conn_pick_ms,
  pre_upstream_latency_ms,
  scheduler_candidate_count,
  scheduler_top_k,
  scheduler_latency_ms,
  COALESCE(scheduler_layer, ''),
  COALESCE(scheduler_reason, ''),
  scheduler_load_skew,
  request_params::text
FROM base
WHERE first_token_ms >= $` + fmt.Sprint(len(args)+1) + `
ORDER BY first_token_ms DESC NULLS LAST, duration_ms DESC NULLS LAST, created_at DESC
LIMIT $` + fmt.Sprint(len(args)+2)
	args = append(args, threshold, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*service.OpsTTFTSlowRequest, 0, limit)
	for rows.Next() {
		item := &service.OpsTTFTSlowRequest{}
		var userID, apiKeyID, accountID, groupID sql.NullInt64
		var requestType sql.NullInt64
		var stream, openAIWSMode bool
		var firstToken, duration, auth, routing, upstream, response, queue, conn, preUpstream sql.NullInt64
		var schedulerCandidateCount, schedulerTopK, schedulerLatency sql.NullInt64
		var schedulerLayer, schedulerReason string
		var schedulerLoadSkew sql.NullFloat64
		var requestParamsRaw string
		if err := rows.Scan(
			&item.CreatedAt,
			&item.RequestID,
			&item.Platform,
			&item.Model,
			&userID,
			&apiKeyID,
			&accountID,
			&groupID,
			&item.UserEmail,
			&item.APIKeyName,
			&item.AccountName,
			&item.GroupName,
			&requestType,
			&stream,
			&openAIWSMode,
			&item.RouteSource,
			&item.CacheStatus,
			&item.SlowReason,
			&item.SlowDetail,
			&firstToken,
			&duration,
			&auth,
			&routing,
			&upstream,
			&response,
			&queue,
			&conn,
			&preUpstream,
			&schedulerCandidateCount,
			&schedulerTopK,
			&schedulerLatency,
			&schedulerLayer,
			&schedulerReason,
			&schedulerLoadSkew,
			&requestParamsRaw,
		); err != nil {
			return nil, err
		}
		item.UserID = nullInt64Ptr(userID)
		item.APIKeyID = nullInt64Ptr(apiKeyID)
		item.AccountID = nullInt64Ptr(accountID)
		item.GroupID = nullInt64Ptr(groupID)
		reqType := service.RequestTypeUnknown
		if requestType.Valid {
			reqType = service.RequestTypeFromInt16(int16(requestType.Int64))
		}
		if reqType == service.RequestTypeUnknown {
			reqType = service.RequestTypeFromLegacy(stream, openAIWSMode)
		}
		item.RequestType = reqType.String()
		item.FirstTokenMs = nullIntPtr(firstToken)
		item.DurationMs = nullIntPtr(duration)
		item.AuthLatencyMs = nullIntPtr(auth)
		item.RoutingMs = nullIntPtr(routing)
		item.UpstreamMs = nullIntPtr(upstream)
		item.ResponseMs = nullIntPtr(response)
		item.QueueWaitMs = nullIntPtr(queue)
		item.ConnPickMs = nullIntPtr(conn)
		item.PreUpstreamMs = nullIntPtr(preUpstream)
		item.SchedulerCandidateCount = nullIntPtr(schedulerCandidateCount)
		item.SchedulerTopK = nullIntPtr(schedulerTopK)
		item.SchedulerLatencyMs = nullIntPtr(schedulerLatency)
		item.SchedulerLayer = schedulerLayer
		item.SchedulerReason = schedulerReason
		if schedulerLoadSkew.Valid {
			v := schedulerLoadSkew.Float64
			item.SchedulerLoadSkew = &v
		}
		item.RequestParams = service.SanitizeRequestParamsForResponse(decodeJSONMap(requestParamsRaw))
		out = append(out, item)
	}
	return out, rows.Err()
}

func safeRatio(n, d int64) float64 {
	if d <= 0 {
		return 0
	}
	return math.Round((float64(n)/float64(d))*10000) / 10000
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	out := int(v.Int64)
	return &out
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
}

func decodeJSONMap(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) == 0 {
		return nil
	}
	return out
}

func decodeJSONInt64Slice(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil
	}
	var values []int64
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values
	}
	return nil
}

func ttftAccountParticipationReason(item *service.OpsTTFTAccountParticipationItem) string {
	if item == nil {
		return ""
	}
	if !item.InObservedGroup {
		return "not_in_observed_group"
	}
	if strings.TrimSpace(item.Status) != service.StatusActive {
		return "not_active"
	}
	if !item.Schedulable {
		return "unschedulable"
	}
	if item.RequestCount == 0 {
		return "candidate_not_selected"
	}
	return "selected"
}

func ttftReasonSuggestion(reason string) string {
	switch reason {
	case "account_queue_slow":
		return "检查自有号池并发容量、账号连接数和排队情况。"
	case "connection_pick_slow":
		return "检查 OpenAI WS 连接池预热、连接复用和代理质量。"
	case "routing_slow":
		return "检查账号选择、分组配置和路由前置逻辑是否过重。"
	case "upstream_ttft_slow":
		return "优先排查上游首字速度、模型本身负载和供应商质量。"
	case "response_flush_slow":
		return "这是整体响应尾部慢，不作为首字前归因；首字分析中会归入上游首字等待。"
	case "platform_overhead_slow":
		return "检查鉴权、计费、日志和请求转换的前置耗时。"
	case "cache_hit":
		return "缓存命中首字接近 0，可单独观察命中率和缓存范围。"
	case "unknown":
		return "样本缺少首字或阶段耗时，建议先补齐埋点再判断。"
	default:
		return "继续观察该慢因的 Top 模型、账号和请求明细。"
	}
}

func ttftTopSuggestion(dimension string, item *service.OpsTTFTTopItem) string {
	if item == nil {
		return ""
	}
	if item.SlowRate >= 0.5 && item.Count >= 5 {
		switch dimension {
		case "model":
			return "该模型慢请求占比较高，建议按模型单独检查上游质量或缓存策略。"
		case "account":
			return "该账号慢请求占比较高，建议检查账号容量、代理和上游状态。"
		case "group":
			return "该分组慢请求占比较高，建议检查分组优先级、模型分布和上游配置。"
		case "api_key":
			return "该 API Key 慢请求占比较高，建议确认是否有固定脚本、长上下文或高并发。"
		}
	}
	return ""
}

func buildTTFTRecommendations(summary service.OpsTTFTSummary, reasons []*service.OpsTTFTReasonItem, topModels, topAccounts, topGroups, topAPIKeys []*service.OpsTTFTTopItem) []*service.OpsTTFTRecommendation {
	out := make([]*service.OpsTTFTRecommendation, 0, 6)
	if summary.FirstTokenSampleCount == 0 {
		return []*service.OpsTTFTRecommendation{{
			Severity: "info",
			Title:    "首字样本不足",
			Message:  "当前时间窗口没有可用的 first_token_ms 样本。",
			Action:   "确认流式请求是否进入 usage_logs，并等待更多真实请求产生。",
		}}
	}
	if summary.TTFT.P95 != nil && *summary.TTFT.P95 >= 1500 {
		out = append(out, &service.OpsTTFTRecommendation{
			Severity: "warning",
			Title:    "TTFT P95 偏高",
			Message:  fmt.Sprintf("当前 P95 首字为 %dms，用户体感等待可能明显增加。", *summary.TTFT.P95),
			Action:   "优先查看慢因占比和 Top 账号/模型，定位是上游慢还是号池排队慢。",
		})
	}
	for _, item := range summary.RouteSources {
		if item == nil || item.SlowCount < 5 || item.SlowRate < 0.3 {
			continue
		}
		switch item.Source {
		case "own_pool":
			out = append(out, &service.OpsTTFTRecommendation{
				Severity: "warning",
				Title:    "自有号池首字偏慢",
				Message:  fmt.Sprintf("自有号池慢请求占比 %.1f%%，共 %d 次慢请求。", item.SlowRate*100, item.SlowCount),
				Action:   "优先检查自有账号并发容量、排队等待、连接复用和代理质量，再看是否需要按模型拆分账号池。",
			})
		case "upstream":
			out = append(out, &service.OpsTTFTRecommendation{
				Severity: "warning",
				Title:    "上游/API Key 首字偏慢",
				Message:  fmt.Sprintf("上游/API Key 慢请求占比 %.1f%%，共 %d 次慢请求。", item.SlowRate*100, item.SlowCount),
				Action:   "优先按上游账号、模型和供应商看 P95，确认是否需要调整兜底顺序、超时或供应商质量分层。",
			})
		}
	}
	if len(reasons) > 0 && reasons[0].Share >= 0.4 && reasons[0].Reason != "normal" {
		out = append(out, &service.OpsTTFTRecommendation{
			Severity: "warning",
			Title:    "慢因集中",
			Message:  fmt.Sprintf("慢因 %s 占比 %.1f%%。", reasons[0].Reason, reasons[0].Share*100),
			Action:   reasons[0].Suggestion,
		})
	}
	if rec := topItemRecommendation("模型", topModels); rec != nil {
		out = append(out, rec)
	}
	if rec := topItemRecommendation("账号", topAccounts); rec != nil {
		out = append(out, rec)
	}
	if rec := topItemRecommendation("分组", topGroups); rec != nil {
		out = append(out, rec)
	}
	if len(out) == 0 {
		out = append(out, &service.OpsTTFTRecommendation{
			Severity: "info",
			Title:    "未发现集中慢因",
			Message:  "当前首字样本没有明显集中慢因。",
			Action:   "继续观察更长窗口，或按模型/账号查看是否存在局部异常。",
		})
	}
	_ = topAPIKeys
	return out
}

func topItemRecommendation(label string, items []*service.OpsTTFTTopItem) *service.OpsTTFTRecommendation {
	if len(items) == 0 || items[0] == nil || items[0].SlowCount < 5 || items[0].SlowRate < 0.5 {
		return nil
	}
	return &service.OpsTTFTRecommendation{
		Severity: "warning",
		Title:    label + "慢请求集中",
		Message:  fmt.Sprintf("%s %s 慢请求占比 %.1f%%。", label, items[0].Label, items[0].SlowRate*100),
		Action:   items[0].Suggestion,
	}
}
