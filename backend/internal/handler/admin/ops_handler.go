package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type OpsHandler struct {
	opsService    *service.OpsService
	responseCache *service.ResponseCache
}

// GetErrorLogByID returns ops error log detail.
// GET /api/v1/admin/ops/errors/:id
func (h *OpsHandler) GetErrorLogByID(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid error id")
		return
	}

	detail, err := h.opsService.GetErrorLogByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, detail)
}

const (
	opsListViewErrors   = "errors"
	opsListViewExcluded = "excluded"
	opsListViewAll      = "all"
)

func parseOpsViewParam(c *gin.Context) string {
	if c == nil {
		return ""
	}
	v := strings.ToLower(strings.TrimSpace(c.Query("view")))
	switch v {
	case "", opsListViewErrors:
		return opsListViewErrors
	case opsListViewExcluded:
		return opsListViewExcluded
	case opsListViewAll:
		return opsListViewAll
	default:
		return opsListViewErrors
	}
}

func parsePositiveIntQuery(c *gin.Context, key string) int {
	if c == nil {
		return 0
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}

func parsePositiveInt64Query(c *gin.Context, key string) int64 {
	if c == nil {
		return 0
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}

func parseFloatQuery(c *gin.Context, key string) float64 {
	if c == nil {
		return 0
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return 0
	}
	return v
}

func parseRatioQuery(c *gin.Context, key string) float64 {
	v := parseFloatQuery(c, key)
	if v > 1 {
		v = v / 100
	}
	if v < 0 || v > 1 {
		return 0
	}
	return v
}

func NewOpsHandler(opsService *service.OpsService) *OpsHandler {
	return &OpsHandler{opsService: opsService}
}

func (h *OpsHandler) SetResponseCache(responseCache *service.ResponseCache) {
	if h == nil {
		return
	}
	h.responseCache = responseCache
}

// GetResponseCacheRecommendation returns the shadow-mode recommendation for exact cache rollout.
// GET /api/v1/admin/ops/response-cache/recommendation
func (h *OpsHandler) GetResponseCacheRecommendation(c *gin.Context) {
	if h == nil || h.responseCache == nil {
		response.Success(c, gin.H{
			"enabled":     false,
			"recommended": false,
			"decision":    "not_available",
			"reasons":     []string{"response_cache_not_configured"},
		})
		return
	}
	opts := service.ResponseCacheRecommendationOptions{
		WindowHours:      parsePositiveIntQuery(c, "window_hours"),
		MinCandidates:    parsePositiveInt64Query(c, "min_candidates"),
		HitRateThreshold: parseRatioQuery(c, "hit_rate_threshold"),
		MinObservedHours: parsePositiveIntQuery(c, "min_observed_hours"),
		MaxSpikeRatio:    parseFloatQuery(c, "max_spike_ratio"),
		MinUniqueKeys:    parsePositiveInt64Query(c, "min_unique_keys"),
		Top1MaxHitShare:  parseRatioQuery(c, "top1_max_hit_share"),
		Top5MaxHitShare:  parseRatioQuery(c, "top5_max_hit_share"),
	}
	rec, err := h.responseCache.GetRecommendation(c.Request.Context(), opts)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, rec)
}

// GetResponseCacheKeyStats returns anonymized top cache-key shadow statistics.
// GET /api/v1/admin/ops/response-cache/keys
func (h *OpsHandler) GetResponseCacheKeyStats(c *gin.Context) {
	if h == nil || h.responseCache == nil {
		response.Success(c, gin.H{
			"enabled":      false,
			"window_hours": parsePositiveIntQuery(c, "window_hours"),
			"items":        []any{},
			"summary": gin.H{
				"unique_keys":  0,
				"tracked_keys": 0,
			},
		})
		return
	}
	opts := service.ResponseCacheKeyStatsOptions{
		WindowHours: parsePositiveIntQuery(c, "window_hours"),
		Limit:       parsePositiveIntQuery(c, "limit"),
		Sort:        c.Query("sort"),
		Model:       c.Query("model"),
		APIKeyID:    parsePositiveInt64Query(c, "api_key_id"),
		Monitor:     c.Query("monitor"),
	}
	if strings.TrimSpace(c.Query("group_id")) != "" {
		opts.GroupIDSet = true
		opts.GroupID = parsePositiveInt64Query(c, "group_id")
	}
	result, err := h.responseCache.GetKeyStats(c.Request.Context(), opts)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// GetTTFTAnalysis returns first-token latency attribution and recommendations.
// GET /api/v1/admin/ops/ttft-analysis
func (h *OpsHandler) GetTTFTAnalysis(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsTTFTAnalysisFilter{
		StartTime: startTime,
		EndTime:   endTime,
		Platform:  strings.TrimSpace(c.Query("platform")),
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filter.GroupID = &id
	}
	if v := strings.TrimSpace(c.Query("slow_threshold_ms")); v != "" {
		threshold, err := strconv.Atoi(v)
		if err != nil || threshold <= 0 {
			response.BadRequest(c, "Invalid slow_threshold_ms")
			return
		}
		filter.SlowThresholdMs = threshold
	}
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit <= 0 {
			response.BadRequest(c, "Invalid limit")
			return
		}
		filter.TopLimit = limit
	}

	out, err := h.opsService.GetTTFTAnalysis(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

// GetErrorLogs lists ops error logs.
// GET /api/v1/admin/ops/errors
func (h *OpsHandler) GetErrorLogs(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	page, pageSize := response.ParsePagination(c)
	// Ops list can be larger than standard admin tables.
	if pageSize > 500 {
		pageSize = 500
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsErrorLogFilter{Page: page, PageSize: pageSize}

	if !startTime.IsZero() {
		filter.StartTime = &startTime
	}
	if !endTime.IsZero() {
		filter.EndTime = &endTime
	}
	filter.View = parseOpsViewParam(c)
	filter.Phase = strings.TrimSpace(c.Query("phase"))
	filter.Owner = strings.TrimSpace(c.Query("error_owner"))
	filter.Source = strings.TrimSpace(c.Query("error_source"))
	filter.Query = strings.TrimSpace(c.Query("q"))
	filter.UserQuery = strings.TrimSpace(c.Query("user_query"))

	// Force request errors: client-visible status >= 400.
	// buildOpsErrorLogsWhere already applies this for non-upstream phase.
	if strings.EqualFold(strings.TrimSpace(filter.Phase), "upstream") {
		filter.Phase = ""
	}

	if platform := strings.TrimSpace(c.Query("platform")); platform != "" {
		filter.Platform = platform
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filter.GroupID = &id
	}
	if v := strings.TrimSpace(c.Query("account_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid account_id")
			return
		}
		filter.AccountID = &id
	}

	if v := strings.TrimSpace(c.Query("resolved")); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes":
			b := true
			filter.Resolved = &b
		case "0", "false", "no":
			b := false
			filter.Resolved = &b
		default:
			response.BadRequest(c, "Invalid resolved")
			return
		}
	}
	if statusCodesStr := strings.TrimSpace(c.Query("status_codes")); statusCodesStr != "" {
		parts := strings.Split(statusCodesStr, ",")
		out := make([]int, 0, len(parts))
		for _, part := range parts {
			p := strings.TrimSpace(part)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 {
				response.BadRequest(c, "Invalid status_codes")
				return
			}
			out = append(out, n)
		}
		filter.StatusCodes = out
	}

	result, err := h.opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// ListRequestErrors lists client-visible request errors.
// GET /api/v1/admin/ops/request-errors
func (h *OpsHandler) ListRequestErrors(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 500 {
		pageSize = 500
	}
	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsErrorLogFilter{Page: page, PageSize: pageSize}
	if !startTime.IsZero() {
		filter.StartTime = &startTime
	}
	if !endTime.IsZero() {
		filter.EndTime = &endTime
	}
	filter.View = parseOpsViewParam(c)
	filter.Phase = strings.TrimSpace(c.Query("phase"))
	filter.Owner = strings.TrimSpace(c.Query("error_owner"))
	filter.Source = strings.TrimSpace(c.Query("error_source"))
	filter.Query = strings.TrimSpace(c.Query("q"))
	filter.UserQuery = strings.TrimSpace(c.Query("user_query"))

	// Force request errors: client-visible status >= 400.
	// buildOpsErrorLogsWhere already applies this for non-upstream phase.
	if strings.EqualFold(strings.TrimSpace(filter.Phase), "upstream") {
		filter.Phase = ""
	}

	if platform := strings.TrimSpace(c.Query("platform")); platform != "" {
		filter.Platform = platform
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filter.GroupID = &id
	}
	if v := strings.TrimSpace(c.Query("account_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid account_id")
			return
		}
		filter.AccountID = &id
	}

	if v := strings.TrimSpace(c.Query("resolved")); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes":
			b := true
			filter.Resolved = &b
		case "0", "false", "no":
			b := false
			filter.Resolved = &b
		default:
			response.BadRequest(c, "Invalid resolved")
			return
		}
	}
	if statusCodesStr := strings.TrimSpace(c.Query("status_codes")); statusCodesStr != "" {
		parts := strings.Split(statusCodesStr, ",")
		out := make([]int, 0, len(parts))
		for _, part := range parts {
			p := strings.TrimSpace(part)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 {
				response.BadRequest(c, "Invalid status_codes")
				return
			}
			out = append(out, n)
		}
		filter.StatusCodes = out
	}

	result, err := h.opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// GetRequestError returns request error detail.
// GET /api/v1/admin/ops/request-errors/:id
func (h *OpsHandler) GetRequestError(c *gin.Context) {
	// same storage; just proxy to existing detail
	h.GetErrorLogByID(c)
}

// ListRequestErrorUpstreamErrors lists upstream error logs correlated to a request error.
// GET /api/v1/admin/ops/request-errors/:id/upstream-errors
func (h *OpsHandler) ListRequestErrorUpstreamErrors(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid error id")
		return
	}

	// Load request error to get correlation keys.
	detail, err := h.opsService.GetErrorLogByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Correlate by request_id/client_request_id.
	requestID := strings.TrimSpace(detail.RequestID)
	clientRequestID := strings.TrimSpace(detail.ClientRequestID)
	if requestID == "" && clientRequestID == "" {
		response.Paginated(c, []*service.OpsErrorLog{}, 0, 1, 10)
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 500 {
		pageSize = 500
	}

	// Keep correlation window wide enough so linked upstream errors
	// are discoverable even when UI defaults to 1h elsewhere.
	startTime, endTime, err := parseOpsTimeRange(c, "30d")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsErrorLogFilter{Page: page, PageSize: pageSize}
	if !startTime.IsZero() {
		filter.StartTime = &startTime
	}
	if !endTime.IsZero() {
		filter.EndTime = &endTime
	}
	filter.View = "all"
	filter.Phase = "upstream"
	filter.Owner = "provider"
	filter.Source = strings.TrimSpace(c.Query("error_source"))
	filter.Query = strings.TrimSpace(c.Query("q"))

	if platform := strings.TrimSpace(c.Query("platform")); platform != "" {
		filter.Platform = platform
	}

	// Prefer exact match on request_id; if missing, fall back to client_request_id.
	if requestID != "" {
		filter.RequestID = requestID
	} else {
		filter.ClientRequestID = clientRequestID
	}

	result, err := h.opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// If client asks for details, expand each upstream error log to include upstream response fields.
	includeDetail := strings.TrimSpace(c.Query("include_detail"))
	if includeDetail == "1" || strings.EqualFold(includeDetail, "true") || strings.EqualFold(includeDetail, "yes") {
		details := make([]*service.OpsErrorLogDetail, 0, len(result.Errors))
		for _, item := range result.Errors {
			if item == nil {
				continue
			}
			d, err := h.opsService.GetErrorLogByID(c.Request.Context(), item.ID)
			if err != nil || d == nil {
				continue
			}
			details = append(details, d)
		}
		response.Paginated(c, details, int64(result.Total), result.Page, result.PageSize)
		return
	}

	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// ResolveRequestError toggles resolved status.
// PUT /api/v1/admin/ops/request-errors/:id/resolve
func (h *OpsHandler) ResolveRequestError(c *gin.Context) {
	h.UpdateErrorResolution(c)
}

// ListUpstreamErrors lists independent upstream errors.
// GET /api/v1/admin/ops/upstream-errors
func (h *OpsHandler) ListUpstreamErrors(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 500 {
		pageSize = 500
	}
	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsErrorLogFilter{Page: page, PageSize: pageSize}
	if !startTime.IsZero() {
		filter.StartTime = &startTime
	}
	if !endTime.IsZero() {
		filter.EndTime = &endTime
	}

	filter.View = parseOpsViewParam(c)
	filter.Phase = "upstream"
	filter.Owner = "provider"
	filter.Source = strings.TrimSpace(c.Query("error_source"))
	filter.Query = strings.TrimSpace(c.Query("q"))

	if platform := strings.TrimSpace(c.Query("platform")); platform != "" {
		filter.Platform = platform
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filter.GroupID = &id
	}
	if v := strings.TrimSpace(c.Query("account_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid account_id")
			return
		}
		filter.AccountID = &id
	}

	if v := strings.TrimSpace(c.Query("resolved")); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes":
			b := true
			filter.Resolved = &b
		case "0", "false", "no":
			b := false
			filter.Resolved = &b
		default:
			response.BadRequest(c, "Invalid resolved")
			return
		}
	}
	if statusCodesStr := strings.TrimSpace(c.Query("status_codes")); statusCodesStr != "" {
		parts := strings.Split(statusCodesStr, ",")
		out := make([]int, 0, len(parts))
		for _, part := range parts {
			p := strings.TrimSpace(part)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 {
				response.BadRequest(c, "Invalid status_codes")
				return
			}
			out = append(out, n)
		}
		filter.StatusCodes = out
	}

	result, err := h.opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// GetUpstreamError returns upstream error detail.
// GET /api/v1/admin/ops/upstream-errors/:id
func (h *OpsHandler) GetUpstreamError(c *gin.Context) {
	h.GetErrorLogByID(c)
}

// ResolveUpstreamError toggles resolved status.
// PUT /api/v1/admin/ops/upstream-errors/:id/resolve
func (h *OpsHandler) ResolveUpstreamError(c *gin.Context) {
	h.UpdateErrorResolution(c)
}

// ==================== Existing endpoints ====================

// ListRequestDetails returns a request-level list (success + error) for drill-down.
// GET /api/v1/admin/ops/requests
func (h *OpsHandler) ListRequestDetails(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 100 {
		pageSize = 100
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsRequestDetailFilter{
		Page:      page,
		PageSize:  pageSize,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	filter.Kind = strings.TrimSpace(c.Query("kind"))
	filter.Platform = strings.TrimSpace(c.Query("platform"))
	filter.Model = strings.TrimSpace(c.Query("model"))
	filter.RequestID = strings.TrimSpace(c.Query("request_id"))
	filter.Query = strings.TrimSpace(c.Query("q"))
	filter.Sort = strings.TrimSpace(c.Query("sort"))

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		filter.APIKeyID = &id
	}
	if v := strings.TrimSpace(c.Query("account_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid account_id")
			return
		}
		filter.AccountID = &id
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filter.GroupID = &id
	}

	if v := strings.TrimSpace(c.Query("min_duration_ms")); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			response.BadRequest(c, "Invalid min_duration_ms")
			return
		}
		filter.MinDurationMs = &parsed
	}
	if v := strings.TrimSpace(c.Query("max_duration_ms")); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			response.BadRequest(c, "Invalid max_duration_ms")
			return
		}
		filter.MaxDurationMs = &parsed
	}

	out, err := h.opsService.ListRequestDetails(c.Request.Context(), filter)
	if err != nil {
		// Invalid sort/kind/platform etc should be a bad request; keep it simple.
		if strings.Contains(strings.ToLower(err.Error()), "invalid") {
			response.BadRequest(c, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to list request details")
		return
	}

	response.Paginated(c, out.Items, out.Total, out.Page, out.PageSize)
}

// ListRequestFilterOptions returns dropdown options for request drilldown filters.
// GET /api/v1/admin/ops/requests/filters
func (h *OpsHandler) ListRequestFilterOptions(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.OpsRequestDetailFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Kind:      strings.TrimSpace(c.Query("kind")),
	}

	out, err := h.opsService.ListRequestFilterOptions(c.Request.Context(), filter)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalid") {
			response.BadRequest(c, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to list request filter options")
		return
	}

	response.Success(c, out)
}

type opsResolveRequest struct {
	Resolved bool `json:"resolved"`
}

// UpdateErrorResolution allows manual resolve/unresolve.
// PUT /api/v1/admin/ops/errors/:id/resolve
func (h *OpsHandler) UpdateErrorResolution(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid error id")
		return
	}

	var req opsResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	uid := subject.UserID
	if err := h.opsService.UpdateErrorResolution(c.Request.Context(), id, req.Resolved, &uid); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func parseOpsTimeRange(c *gin.Context, defaultRange string) (time.Time, time.Time, error) {
	startStr := strings.TrimSpace(c.Query("start_time"))
	endStr := strings.TrimSpace(c.Query("end_time"))

	parseTS := func(s string) (time.Time, error) {
		if s == "" {
			return time.Time{}, nil
		}
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t, nil
		}
		return time.Parse(time.RFC3339, s)
	}

	start, err := parseTS(startStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseTS(endStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// start/end explicitly provided (even partially)
	if startStr != "" || endStr != "" {
		if end.IsZero() {
			end = time.Now()
		}
		if start.IsZero() {
			dur, _ := parseOpsDuration(defaultRange)
			start = end.Add(-dur)
		}
		if start.After(end) {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time range: start_time must be <= end_time")
		}
		if end.Sub(start) > 30*24*time.Hour {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time range: max window is 30 days")
		}
		return start, end, nil
	}

	// time_range fallback
	tr := strings.TrimSpace(c.Query("time_range"))
	if tr == "" {
		tr = defaultRange
	}
	dur, ok := parseOpsDuration(tr)
	if !ok {
		dur, _ = parseOpsDuration(defaultRange)
	}

	end = time.Now()
	start = end.Add(-dur)
	if end.Sub(start) > 30*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range: max window is 30 days")
	}
	return start, end, nil
}

func parseOpsDuration(v string) (time.Duration, bool) {
	switch strings.TrimSpace(v) {
	case "5m":
		return 5 * time.Minute, true
	case "30m":
		return 30 * time.Minute, true
	case "1h":
		return time.Hour, true
	case "6h":
		return 6 * time.Hour, true
	case "24h":
		return 24 * time.Hour, true
	case "7d":
		return 7 * 24 * time.Hour, true
	case "30d":
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}
