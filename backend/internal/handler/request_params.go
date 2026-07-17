package handler

import (
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	opsRequestParamsKey      = "ops_request_params"
	requestParamListMaxItems = 8
)

const (
	opsRequestParamTTFTSourceUnknown  = "unknown"
	opsRequestParamTTFTSourceOwnPool  = "own_pool"
	opsRequestParamTTFTSourceUpstream = "upstream"
	opsRequestParamTTFTSourceCache    = "cache"

	opsRequestParamTTFTSlowUnknown      = "unknown"
	opsRequestParamTTFTSlowCacheHit     = "cache_hit"
	opsRequestParamTTFTSlowAccountQueue = "account_queue_slow"
	opsRequestParamTTFTSlowConnPick     = "connection_pick_slow"
	opsRequestParamTTFTSlowRouting      = "routing_slow"
	opsRequestParamTTFTSlowUpstream     = "upstream_ttft_slow"
	opsRequestParamTTFTSlowPlatform     = "platform_overhead_slow"
	opsRequestParamTTFTSlowNormal       = "normal"
)

func setOpsRequestParamsFromBody(c *gin.Context, body []byte) {
	if c == nil || len(body) == 0 {
		return
	}
	params := buildSanitizedRequestParamsFromBody(body)
	if len(params) == 0 {
		return
	}
	c.Set(opsRequestParamsKey, params)
}

func getOpsRequestParams(c *gin.Context) map[string]any {
	if c == nil {
		return nil
	}
	raw, ok := c.Get(opsRequestParamsKey)
	if !ok {
		return nil
	}
	params, ok := raw.(map[string]any)
	if !ok || len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for k, v := range params {
		out[k] = v
	}
	return out
}

func getOpsRequestParamsWithTTFTObservation(c *gin.Context, account *service.Account, result interface{}) map[string]any {
	params := getOpsRequestParams(c)
	obs := buildTTFTObservationParams(c, account, result)
	if len(obs) == 0 {
		return params
	}
	if params == nil {
		params = make(map[string]any, len(obs))
	}
	for k, v := range obs {
		params[k] = v
	}
	return params
}

func mergeOpsRequestParams(c *gin.Context, params map[string]any) {
	if c == nil || len(params) == 0 {
		return
	}
	merged := getOpsRequestParams(c)
	if merged == nil {
		merged = make(map[string]any, len(params))
	}
	for k, v := range params {
		if strings.TrimSpace(k) == "" || v == nil {
			continue
		}
		merged[k] = v
	}
	if len(merged) > 0 {
		c.Set(opsRequestParamsKey, merged)
	}
}

func mergeOpsSchedulerDecision(c *gin.Context, decision service.OpenAIAccountScheduleDecision) {
	params := make(map[string]any, 10)
	if layer := strings.TrimSpace(decision.Layer); layer != "" {
		params["scheduler_layer"] = layer
	}
	if reason := schedulerDecisionReason(decision); reason != "" {
		params["scheduler_reason"] = reason
	}
	if decision.CandidateCount > 0 {
		params["scheduler_candidate_count"] = decision.CandidateCount
	}
	if decision.TopK > 0 {
		params["scheduler_top_k"] = decision.TopK
	}
	if decision.LatencyMs > 0 {
		params["scheduler_latency_ms"] = decision.LatencyMs
	}
	if decision.LoadSkew > 0 {
		params["scheduler_load_skew"] = decision.LoadSkew
	}
	if decision.SelectedAccountID > 0 {
		params["scheduler_selected_account_id"] = decision.SelectedAccountID
	}
	if selectedType := strings.TrimSpace(decision.SelectedAccountType); selectedType != "" {
		params["scheduler_selected_account_type"] = selectedType
	}
	if candidates := opsSchedulerDecisionCandidates(decision.Candidates); len(candidates) > 0 {
		params["scheduler_candidates"] = candidates
		if len(decision.Candidates) > len(candidates) {
			params["scheduler_candidates_truncated"] = true
		}
	}
	mergeOpsRequestParams(c, params)
}

func schedulerDecisionReason(decision service.OpenAIAccountScheduleDecision) string {
	if decision.StickyTTFTBypass {
		return "sticky_ttft_bypass"
	}
	switch strings.TrimSpace(decision.Layer) {
	case "previous_response_id":
		return "previous_response_sticky"
	case "session_hash":
		if decision.StickySessionHit {
			return "session_sticky_hit"
		}
		return "session_sticky"
	case "load_balance":
		if decision.StickySessionHit {
			return "legacy_session_sticky_hit"
		}
		if decision.TopK > 0 && decision.CandidateCount > decision.TopK {
			return "load_balance_top_k"
		}
		return "load_balance_priority_load_queue_ttft"
	default:
		return ""
	}
}

func opsSchedulerDecisionCandidates(candidates []service.OpenAIAccountScheduleCandidateDiagnostic) []service.OpenAIAccountScheduleCandidateDiagnostic {
	const maxCandidates = 12
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) <= maxCandidates {
		return candidates
	}
	cloned := make([]service.OpenAIAccountScheduleCandidateDiagnostic, maxCandidates)
	copy(cloned, candidates[:maxCandidates])
	return cloned
}

func buildTTFTObservationParams(c *gin.Context, account *service.Account, result interface{}) map[string]any {
	out := make(map[string]any, 16)
	if c == nil {
		return out
	}

	if account != nil {
		out["route_source"] = classifyTTFTRouteSource(account)
		out["account_type"] = strings.TrimSpace(account.Type)
	} else {
		out["route_source"] = opsRequestParamTTFTSourceUnknown
	}

	if v, ok := getContextInt64Local(c, service.OpsAuthLatencyMsKey); ok {
		out["auth_latency_ms"] = v
	}
	if v, ok := getContextInt64Local(c, service.OpsRoutingLatencyMsKey); ok {
		out["routing_latency_ms"] = v
	}
	if v, ok := getContextInt64Local(c, service.OpsUpstreamLatencyMsKey); ok {
		out["upstream_latency_ms"] = v
	}
	if v, ok := getContextInt64Local(c, service.OpsResponseLatencyMsKey); ok {
		out["response_latency_ms"] = v
	}
	if v, ok := getContextInt64Local(c, service.OpsTimeToFirstTokenMsKey); ok {
		out["time_to_first_token_ms"] = v
	}
	if v, ok := getContextInt64Local(c, service.OpsOpenAIWSQueueWaitMsKey); ok {
		out["openai_ws_queue_wait_ms"] = v
	}
	if v, ok := getContextInt64Local(c, service.OpsOpenAIWSConnPickMsKey); ok {
		out["openai_ws_conn_pick_ms"] = v
	}
	if status := responseCacheStatusFromContext(c); status != "" {
		out["response_cache_status"] = status
		if status == service.ResponseCacheStatusHit {
			out["route_source"] = opsRequestParamTTFTSourceCache
		}
	}
	if reason := responseCacheBypassReasonFromContext(c); reason != "" {
		out["response_cache_bypass_reason"] = reason
	}

	if firstTokenMs, ok := ttftFirstTokenMsFromResult(result); ok {
		out["first_token_ms"] = firstTokenMs
	}
	if durationMs, ok := ttftDurationMsFromResult(result); ok {
		out["duration_ms"] = durationMs
	}

	reason, detail := classifyTTFTSlowReason(out)
	out["ttft_slow_reason"] = reason
	if detail != "" {
		out["ttft_slow_reason_detail"] = detail
	}

	return out
}

func classifyTTFTRouteSource(account *service.Account) string {
	if account == nil {
		return opsRequestParamTTFTSourceUnknown
	}
	accountType := strings.ToLower(strings.TrimSpace(account.Type))
	switch accountType {
	case service.AccountTypeAPIKey, service.AccountTypeUpstream, service.AccountTypeBedrock, service.AccountTypeServiceAccount,
		"api", "api_key", "openai-api", "openai_api":
		return opsRequestParamTTFTSourceUpstream
	default:
		return opsRequestParamTTFTSourceOwnPool
	}
}

func responseCacheStatusFromContext(c *gin.Context) string {
	if c == nil || c.Writer == nil {
		return ""
	}
	raw := strings.TrimSpace(c.Writer.Header().Get(service.ResponseCacheHeader))
	if raw == "" {
		return ""
	}
	status := strings.TrimSpace(strings.Split(raw, ";")[0])
	switch status {
	case service.ResponseCacheStatusHit, service.ResponseCacheStatusMiss, service.ResponseCacheStatusBypass, service.ResponseCacheStatusShadow:
		return status
	default:
		return status
	}
}

func responseCacheBypassReasonFromContext(c *gin.Context) string {
	if c == nil || c.Writer == nil {
		return ""
	}
	raw := strings.TrimSpace(c.Writer.Header().Get(service.ResponseCacheHeader))
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ";")
	if len(parts) < 2 || strings.TrimSpace(parts[0]) != service.ResponseCacheStatusBypass {
		return ""
	}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || strings.TrimSpace(key) != "reason" {
			continue
		}
		return strings.TrimSpace(value)
	}
	return ""
}

func ttftFirstTokenMsFromResult(result interface{}) (int64, bool) {
	switch v := result.(type) {
	case *service.ForwardResult:
		if v != nil && v.FirstTokenMs != nil {
			return int64(*v.FirstTokenMs), true
		}
	case *service.OpenAIForwardResult:
		if v != nil && v.FirstTokenMs != nil {
			return int64(*v.FirstTokenMs), true
		}
	}
	return 0, false
}

func ttftDurationMsFromResult(result interface{}) (int64, bool) {
	switch v := result.(type) {
	case *service.ForwardResult:
		if v != nil {
			return v.Duration.Milliseconds(), true
		}
	case *service.OpenAIForwardResult:
		if v != nil {
			return v.Duration.Milliseconds(), true
		}
	}
	return 0, false
}

func classifyTTFTSlowReason(params map[string]any) (string, string) {
	if len(params) == 0 {
		return opsRequestParamTTFTSlowUnknown, ""
	}
	cacheStatus, _ := params["response_cache_status"].(string)
	if cacheStatus == service.ResponseCacheStatusHit {
		return opsRequestParamTTFTSlowCacheHit, "response cache hit returned without upstream wait"
	}

	ttft, ok := numericParamInt64(params, "time_to_first_token_ms")
	if !ok {
		ttft, ok = numericParamInt64(params, "first_token_ms")
	}
	if !ok {
		return opsRequestParamTTFTSlowUnknown, "missing first token sample"
	}
	if ttft < 800 {
		return opsRequestParamTTFTSlowNormal, "first token below observation threshold"
	}

	queue, _ := numericParamInt64(params, "openai_ws_queue_wait_ms")
	if queue >= 300 && queue*100/maxInt64(ttft, 1) >= 35 {
		return opsRequestParamTTFTSlowAccountQueue, "OpenAI WS queue wait dominates TTFT"
	}
	connPick, _ := numericParamInt64(params, "openai_ws_conn_pick_ms")
	if connPick >= 200 && connPick*100/maxInt64(ttft, 1) >= 25 {
		return opsRequestParamTTFTSlowConnPick, "OpenAI WS connection pick is high"
	}
	routing, _ := numericParamInt64(params, "routing_latency_ms")
	if routing >= 200 && routing*100/maxInt64(ttft, 1) >= 25 {
		return opsRequestParamTTFTSlowRouting, "routing or account selection is high"
	}
	auth, _ := numericParamInt64(params, "auth_latency_ms")
	if auth+routing >= 300 && (auth+routing)*100/maxInt64(ttft, 1) >= 25 {
		return opsRequestParamTTFTSlowPlatform, "platform pre-upstream processing is high"
	}
	upstream, _ := numericParamInt64(params, "upstream_latency_ms")
	if upstream >= 500 && upstream*100/maxInt64(ttft, 1) >= 60 {
		return opsRequestParamTTFTSlowUpstream, "upstream wait dominates TTFT"
	}
	return opsRequestParamTTFTSlowUpstream, "TTFT is high and no platform stage dominates"
}

func numericParamInt64(params map[string]any, key string) (int64, bool) {
	if len(params) == 0 || strings.TrimSpace(key) == "" {
		return 0, false
	}
	v, ok := params[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case int:
		return int64(t), true
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case float32:
		return int64(t), true
	case float64:
		return int64(t), true
	default:
		return 0, false
	}
}

func getContextInt64Local(c *gin.Context, key string) (int64, bool) {
	if c == nil || strings.TrimSpace(key) == "" {
		return 0, false
	}
	v, ok := c.Get(key)
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case int:
		return int64(t), true
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case float32:
		return int64(t), true
	case float64:
		return int64(t), true
	default:
		return 0, false
	}
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func setOpsRequestParamsFromOpenAIImagesRequest(c *gin.Context, req *service.OpenAIImagesRequest) {
	params := buildSanitizedOpenAIImagesRequestParams(req)
	if len(params) == 0 {
		return
	}
	mergeOpsRequestParams(c, params)
}

func buildSanitizedRequestParamsFromBody(body []byte) map[string]any {
	if !gjson.ValidBytes(body) {
		return nil
	}

	out := make(map[string]any, 16)
	putJSONString(out, body, "model", "model")
	putJSONBool(out, body, "stream", "stream")
	putJSONNumber(out, body, "n", "n")
	putJSONNumber(out, body, "max_tokens", "max_tokens")
	putJSONNumber(out, body, "max_output_tokens", "max_output_tokens")
	putJSONNumber(out, body, "temperature", "temperature")
	putJSONNumber(out, body, "top_p", "top_p")
	putJSONString(out, body, "size", "image_size")
	putJSONString(out, body, "quality", "quality")
	putJSONString(out, body, "output_format", "output_format")
	putJSONString(out, body, "background", "background")
	putJSONString(out, body, "moderation", "moderation")
	putJSONString(out, body, "input_fidelity", "input_fidelity")
	putJSONString(out, body, "style", "style")
	putJSONString(out, body, "response_format", "response_format")
	putJSONString(out, body, "service_tier", "service_tier")
	putJSONString(out, body, "reasoning.effort", "reasoning_effort")
	putJSONString(out, body, "reasoning_effort", "reasoning_effort")
	putJSONString(out, body, "generationConfig.imageConfig.imageSize", "image_size")
	putJSONNumber(out, body, "output_compression", "output_compression")
	putJSONNumber(out, body, "partial_images", "partial_images")

	if prompt := strings.TrimSpace(gjson.GetBytes(body, "prompt").String()); prompt != "" {
		out["prompt_chars"] = utf8.RuneCountInString(prompt)
	}
	if input := strings.TrimSpace(gjson.GetBytes(body, "input").String()); input != "" && gjson.GetBytes(body, "input").Type == gjson.String {
		out["input_chars"] = utf8.RuneCountInString(input)
	}

	captureArraySummary(out, body, "messages", "messages_count")
	captureArraySummary(out, body, "input", "input_items_count")
	captureArraySummary(out, body, "contents", "contents_count")
	captureToolSummary(out, body)

	if len(out) == 0 {
		return nil
	}
	return out
}

func buildSanitizedOpenAIImagesRequestParams(req *service.OpenAIImagesRequest) map[string]any {
	if req == nil {
		return nil
	}
	out := make(map[string]any, 16)
	if model := strings.TrimSpace(req.Model); model != "" {
		out["model"] = model
	}
	out["stream"] = req.Stream
	if req.N > 0 {
		out["n"] = req.N
	}
	if size := strings.TrimSpace(req.Size); size != "" {
		out["image_size"] = size
	}
	if tier := strings.TrimSpace(req.SizeTier); tier != "" {
		out["image_size_tier"] = tier
	}
	putTrimmedString(out, "response_format", req.ResponseFormat)
	putTrimmedString(out, "quality", req.Quality)
	putTrimmedString(out, "background", req.Background)
	putTrimmedString(out, "output_format", req.OutputFormat)
	putTrimmedString(out, "moderation", req.Moderation)
	putTrimmedString(out, "input_fidelity", req.InputFidelity)
	putTrimmedString(out, "style", req.Style)
	if req.OutputCompression != nil {
		out["output_compression"] = *req.OutputCompression
	}
	if req.PartialImages != nil {
		out["partial_images"] = *req.PartialImages
	}
	if prompt := strings.TrimSpace(req.Prompt); prompt != "" {
		out["prompt_chars"] = utf8.RuneCountInString(prompt)
	}
	if req.Multipart {
		out["multipart"] = true
	}
	if uploadCount := len(req.Uploads); uploadCount > 0 {
		out["input_image_count"] = uploadCount
	}
	if req.HasMask || req.MaskUpload != nil || strings.TrimSpace(req.MaskImageURL) != "" {
		out["has_mask"] = true
	}
	if len(req.InputImageURLs) > 0 {
		out["input_image_url_count"] = len(req.InputImageURLs)
	}
	if req.ExplicitModel {
		out["explicit_model"] = true
	}
	if req.ExplicitSize {
		out["explicit_size"] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func putTrimmedString(out map[string]any, key string, value string) {
	if s := strings.TrimSpace(value); s != "" {
		out[key] = s
	}
}

func putJSONString(out map[string]any, body []byte, path string, key string) {
	if _, exists := out[key]; exists {
		return
	}
	v := gjson.GetBytes(body, path)
	if !v.Exists() || v.Type != gjson.String {
		return
	}
	if s := strings.TrimSpace(v.String()); s != "" {
		out[key] = s
	}
}

func putJSONBool(out map[string]any, body []byte, path string, key string) {
	v := gjson.GetBytes(body, path)
	if !v.Exists() || (v.Type != gjson.True && v.Type != gjson.False) {
		return
	}
	out[key] = v.Bool()
}

func putJSONNumber(out map[string]any, body []byte, path string, key string) {
	v := gjson.GetBytes(body, path)
	if !v.Exists() || v.Type != gjson.Number {
		return
	}
	if strings.Contains(v.Raw, ".") {
		out[key] = v.Float()
		return
	}
	out[key] = v.Int()
}

func captureArraySummary(out map[string]any, body []byte, path string, countKey string) {
	arr := gjson.GetBytes(body, path)
	if !arr.IsArray() {
		return
	}

	count := 0
	arr.ForEach(func(_, item gjson.Result) bool {
		count++
		return true
	})
	if count > 0 {
		out[countKey] = count
	}
}

func captureToolSummary(out map[string]any, body []byte) {
	tools := gjson.GetBytes(body, "tools")
	if !tools.IsArray() {
		return
	}
	types := make([]string, 0, requestParamListMaxItems)
	tools.ForEach(func(_, item gjson.Result) bool {
		toolType := strings.TrimSpace(item.Get("type").String())
		if toolType != "" && len(types) < requestParamListMaxItems {
			types = append(types, toolType)
		}
		if toolType == "image_generation" {
			putJSONItemString(out, item, "size", "image_size")
			putJSONItemString(out, item, "quality", "quality")
			putJSONItemString(out, item, "output_format", "output_format")
			putJSONItemString(out, item, "background", "background")
			putJSONItemString(out, item, "moderation", "moderation")
			putJSONItemNumber(out, item, "partial_images", "partial_images")
		}
		return true
	})
	if len(types) > 0 {
		out["tool_types"] = types
	}
}

func putJSONItemString(out map[string]any, item gjson.Result, path string, key string) {
	if _, exists := out[key]; exists {
		return
	}
	v := item.Get(path)
	if !v.Exists() || v.Type != gjson.String {
		return
	}
	if s := strings.TrimSpace(v.String()); s != "" {
		out[key] = s
	}
}

func putJSONItemNumber(out map[string]any, item gjson.Result, path string, key string) {
	if _, exists := out[key]; exists {
		return
	}
	v := item.Get(path)
	if !v.Exists() || v.Type != gjson.Number {
		return
	}
	out[key] = v.Int()
}
