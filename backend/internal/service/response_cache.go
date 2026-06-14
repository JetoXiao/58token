package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
)

const (
	ResponseCacheHeader = "x-uado-cache"

	ResponseCacheStatusHit    = "hit"
	ResponseCacheStatusMiss   = "miss"
	ResponseCacheStatusBypass = "bypass"
	ResponseCacheStatusShadow = "shadow"
)

type ResponseCache struct {
	cfg              config.ResponseCacheConfig
	rdb              *redis.Client
	redisBypassUntil atomic.Int64
}

type ResponseCacheRequest struct {
	Endpoint string
	Protocol string
	Model    string
	Stream   bool
	Body     []byte
	APIKeyID int64
	GroupID  *int64
	Headers  http.Header
}

type ResponseCacheDecision struct {
	Enabled       bool
	ExactEnabled  bool
	ShadowEnabled bool
	Monitor       bool
	Key           string
	Reason        string
}

type ResponseCacheEntry struct {
	StatusCode int         `json:"status_code"`
	Header     http.Header `json:"header"`
	Body       []byte      `json:"body"`
	StoredAt   time.Time   `json:"stored_at"`
}

type ResponseCacheRecommendationOptions struct {
	WindowHours      int
	MinCandidates    int64
	HitRateThreshold float64
	MinObservedHours int
	MaxSpikeRatio    float64
}

type ResponseCacheHourlyStat struct {
	Hour         string  `json:"hour"`
	Total        int64   `json:"total"`
	Hit          int64   `json:"hit"`
	MonitorTotal int64   `json:"monitor_total"`
	MonitorHit   int64   `json:"monitor_hit"`
	HitRate      float64 `json:"hit_rate"`
}

type ResponseCacheRecommendation struct {
	Enabled             bool                       `json:"enabled"`
	Recommended         bool                       `json:"recommended"`
	Decision            string                     `json:"decision"`
	Reasons             []string                   `json:"reasons"`
	WindowHours         int                        `json:"window_hours"`
	ObservedHours       int                        `json:"observed_hours"`
	MinObservedHours    int                        `json:"min_observed_hours"`
	TotalCandidates     int64                      `json:"total_candidates"`
	ShadowHits          int64                      `json:"shadow_hits"`
	MonitorCandidates   int64                      `json:"monitor_candidates"`
	MonitorHits         int64                      `json:"monitor_hits"`
	HitRate             float64                    `json:"hit_rate"`
	Threshold           float64                    `json:"threshold"`
	MinCandidates       int64                      `json:"min_candidates"`
	BelowThresholdHours int                        `json:"below_threshold_hours"`
	SpikeDetected       bool                       `json:"spike_detected"`
	MaxSpikeRatio       float64                    `json:"max_spike_ratio"`
	GeneratedAt         time.Time                  `json:"generated_at"`
	Hours               []*ResponseCacheHourlyStat `json:"hours"`
}

type responseCachePayload struct {
	StatusCode int                 `json:"status_code"`
	Header     map[string][]string `json:"header"`
	Body       string              `json:"body"`
	StoredAt   int64               `json:"stored_at"`
}

func NewResponseCache(cfg *config.Config, rdb *redis.Client) *ResponseCache {
	if cfg == nil {
		return nil
	}
	rcCfg := cfg.ResponseCache
	if !rcCfg.Enabled && !rcCfg.ShadowEnabled && !rcCfg.RecommendationEnabled {
		return nil
	}
	return &ResponseCache{cfg: rcCfg, rdb: rdb}
}

func (c *ResponseCache) Decide(req ResponseCacheRequest) ResponseCacheDecision {
	if c == nil {
		return ResponseCacheDecision{Reason: "disabled"}
	}
	if !c.cfg.Enabled && !c.cfg.ShadowEnabled {
		return ResponseCacheDecision{Reason: "disabled"}
	}
	if c.rdb == nil {
		return ResponseCacheDecision{Reason: "redis_unavailable"}
	}
	if c.temporarilyBypassed() {
		return ResponseCacheDecision{Reason: "redis_unavailable"}
	}
	if len(req.Body) == 0 {
		return ResponseCacheDecision{Reason: "empty_body"}
	}
	if c.cfg.MaxBodyBytes > 0 && len(req.Body) > c.cfg.MaxBodyBytes {
		return ResponseCacheDecision{Reason: "body_too_large"}
	}
	if shouldBypassCacheControl(req.Headers) {
		return ResponseCacheDecision{Reason: "request_bypass"}
	}
	if hasID(req.APIKeyID, c.cfg.BypassAPIKeyIDs) {
		return ResponseCacheDecision{Reason: "api_key_bypass"}
	}
	if req.GroupID != nil && hasID(*req.GroupID, c.cfg.BypassGroupIDs) {
		return ResponseCacheDecision{Reason: "group_bypass"}
	}
	if matchStringList(req.Model, c.cfg.BypassModels) {
		return ResponseCacheDecision{Reason: "model_bypass"}
	}
	if looksLikeImageOrToolRequest(req.Endpoint, req.Body) {
		return ResponseCacheDecision{Reason: "unsupported_request"}
	}
	promptChars := approxPromptChars(req.Body)
	if c.cfg.MinPromptChars > 0 && promptChars < c.cfg.MinPromptChars {
		return ResponseCacheDecision{Reason: "prompt_too_short"}
	}
	if c.cfg.MaxPromptChars > 0 && promptChars > c.cfg.MaxPromptChars {
		return ResponseCacheDecision{Reason: "prompt_too_long"}
	}
	key, ok := c.cacheKey(req)
	if !ok {
		return ResponseCacheDecision{Reason: "invalid_json"}
	}
	monitor := c.isMonitorRequest(req)
	if req.Stream {
		// Stream exact replay is intentionally deferred in v1 to avoid adding
		// capture overhead to the user's first-token path.
		if !c.cfg.ShadowEnabled {
			return ResponseCacheDecision{Reason: "stream_bypass"}
		}
		return ResponseCacheDecision{
			Enabled:       true,
			ShadowEnabled: true,
			Monitor:       monitor,
			Key:           key,
			Reason:        "stream_shadow_only",
		}
	}
	deterministic := isDeterministicRequest(req.Body)
	exactAllowed := c.cfg.Enabled && deterministic && !monitor && c.scopeAllowsExact(req)
	return ResponseCacheDecision{
		Enabled:       exactAllowed || c.cfg.ShadowEnabled,
		ExactEnabled:  exactAllowed,
		ShadowEnabled: c.cfg.ShadowEnabled,
		Monitor:       monitor,
		Key:           key,
		Reason:        nonDeterministicReason(deterministic, exactAllowed),
	}
}

func (c *ResponseCache) Lookup(ctx context.Context, decision ResponseCacheDecision) (*ResponseCacheEntry, bool) {
	if c == nil || !decision.ExactEnabled || decision.Key == "" || c.rdb == nil || c.temporarilyBypassed() {
		return nil, false
	}
	cacheCtx, cancel := c.shortContext(ctx)
	defer cancel()
	raw, err := c.rdb.Get(cacheCtx, c.valueKey(decision.Key)).Bytes()
	if err != nil {
		c.markRedisFailure(err)
		return nil, false
	}
	if len(raw) == 0 {
		return nil, false
	}
	entry, err := decodeResponseCacheEntry(raw)
	if err != nil {
		return nil, false
	}
	return entry, true
}

func (c *ResponseCache) StoreAsync(decision ResponseCacheDecision, entry *ResponseCacheEntry) bool {
	if c == nil || !decision.ExactEnabled || decision.Key == "" || entry == nil || c.rdb == nil {
		return false
	}
	if entry.StatusCode < 200 || entry.StatusCode >= 300 {
		return false
	}
	if len(entry.Body) == 0 || (c.cfg.MaxValueBytes > 0 && len(entry.Body) > c.cfg.MaxValueBytes) {
		return false
	}
	payload, err := encodeResponseCacheEntry(entry)
	if err != nil || len(payload) == 0 {
		return false
	}
	if c.cfg.MaxValueBytes > 0 && len(payload) > c.cfg.MaxValueBytes {
		return false
	}
	ttl := time.Duration(c.cfg.TTLSeconds) * time.Second
	key := c.valueKey(decision.Key)
	go func() {
		ctx, cancel := c.shortContext(context.Background())
		defer cancel()
		if err := c.rdb.Set(ctx, key, payload, ttl).Err(); err != nil {
			c.markRedisFailure(err)
		}
		if err := c.rdb.Del(ctx, c.flightKey(decision.Key)).Err(); err != nil {
			c.markRedisFailure(err)
		}
	}()
	return true
}

func (c *ResponseCache) ObserveShadowAsync(decision ResponseCacheDecision) {
	if c == nil || !decision.ShadowEnabled || decision.Key == "" || c.rdb == nil {
		return
	}
	ttl := time.Duration(c.cfg.ShadowTTLSeconds) * time.Second
	key := c.seenKey(decision.Key)
	if decision.Monitor {
		key = c.monitorSeenKey(decision.Key)
	}
	go func() {
		ctx, cancel := c.shortContext(context.Background())
		defer cancel()
		created, err := c.rdb.SetNX(ctx, key, "1", ttl).Result()
		if err != nil {
			c.markRedisFailure(err)
			return
		}
		statsTTL := c.statsTTL()
		pipe := c.rdb.Pipeline()
		now := time.Now()
		totalKind := "shadow_total"
		hitKind := "shadow_hit"
		if decision.Monitor {
			totalKind = "monitor_total"
			hitKind = "monitor_hit"
		}
		totalKey := c.statsKey(totalKind)
		pipe.Incr(ctx, totalKey)
		pipe.Expire(ctx, totalKey, statsTTL)
		hourlyTotalKey := c.hourlyStatsKey(totalKind, now)
		pipe.Incr(ctx, hourlyTotalKey)
		pipe.Expire(ctx, hourlyTotalKey, statsTTL)
		if !created {
			hitKey := c.statsKey(hitKind)
			pipe.Incr(ctx, hitKey)
			pipe.Expire(ctx, hitKey, statsTTL)
			hourlyHitKey := c.hourlyStatsKey(hitKind, now)
			pipe.Incr(ctx, hourlyHitKey)
			pipe.Expire(ctx, hourlyHitKey, statsTTL)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			c.markRedisFailure(err)
		}
	}()
}

func (c *ResponseCache) GetRecommendation(ctx context.Context, opts ResponseCacheRecommendationOptions) (*ResponseCacheRecommendation, error) {
	opts = c.normalizeRecommendationOptions(opts)
	rec := &ResponseCacheRecommendation{
		Enabled:          c != nil && c.cfg.RecommendationEnabled,
		Decision:         "not_recommended",
		WindowHours:      opts.WindowHours,
		MinObservedHours: opts.MinObservedHours,
		Threshold:        opts.HitRateThreshold,
		MinCandidates:    opts.MinCandidates,
		MaxSpikeRatio:    opts.MaxSpikeRatio,
		GeneratedAt:      time.Now(),
	}
	if c == nil || !c.cfg.RecommendationEnabled {
		rec.Reasons = append(rec.Reasons, "recommendation_disabled")
		return rec, nil
	}
	if c.rdb == nil {
		rec.Reasons = append(rec.Reasons, "redis_unavailable")
		return rec, nil
	}
	cacheCtx, cancel := c.shortContext(ctx)
	defer cancel()

	endHour := time.Now().Truncate(time.Hour)
	hours := make([]time.Time, 0, opts.WindowHours)
	keys := make([]string, 0, opts.WindowHours*4)
	for i := opts.WindowHours - 1; i >= 0; i-- {
		hour := endHour.Add(-time.Duration(i) * time.Hour)
		hours = append(hours, hour)
		keys = append(keys,
			c.hourlyStatsKey("shadow_total", hour),
			c.hourlyStatsKey("shadow_hit", hour),
			c.hourlyStatsKey("monitor_total", hour),
			c.hourlyStatsKey("monitor_hit", hour),
		)
	}

	values, err := c.rdb.MGet(cacheCtx, keys...).Result()
	if err != nil {
		c.markRedisFailure(err)
		rec.Reasons = append(rec.Reasons, "redis_error")
		return rec, nil
	}

	var maxHourly int64
	for i, hour := range hours {
		offset := i * 4
		total := redisInt64(values[offset])
		hit := redisInt64(values[offset+1])
		monitorTotal := redisInt64(values[offset+2])
		monitorHit := redisInt64(values[offset+3])
		stat := &ResponseCacheHourlyStat{
			Hour:         hour.Format(time.RFC3339),
			Total:        total,
			Hit:          hit,
			MonitorTotal: monitorTotal,
			MonitorHit:   monitorHit,
		}
		if total > 0 {
			stat.HitRate = float64(hit) / float64(total)
			rec.ObservedHours++
			if stat.HitRate < opts.HitRateThreshold {
				rec.BelowThresholdHours++
			}
			if total > maxHourly {
				maxHourly = total
			}
		}
		rec.TotalCandidates += total
		rec.ShadowHits += hit
		rec.MonitorCandidates += monitorTotal
		rec.MonitorHits += monitorHit
		rec.Hours = append(rec.Hours, stat)
	}
	if rec.TotalCandidates > 0 {
		rec.HitRate = float64(rec.ShadowHits) / float64(rec.TotalCandidates)
	}
	if opts.MaxSpikeRatio > 0 && rec.ObservedHours > 0 {
		avgHourly := float64(rec.TotalCandidates) / float64(rec.ObservedHours)
		rec.SpikeDetected = avgHourly > 0 && float64(maxHourly) > avgHourly*opts.MaxSpikeRatio
	}

	if c.cfg.Enabled {
		rec.Decision = "already_enabled"
		rec.Reasons = append(rec.Reasons, "exact_cache_already_enabled")
		return rec, nil
	}
	if rec.ObservedHours < opts.MinObservedHours {
		rec.Reasons = append(rec.Reasons, "insufficient_observed_hours")
	}
	if rec.TotalCandidates < opts.MinCandidates {
		rec.Reasons = append(rec.Reasons, "insufficient_candidates")
	}
	if rec.HitRate < opts.HitRateThreshold {
		rec.Reasons = append(rec.Reasons, "low_shadow_hit_rate")
	}
	if rec.BelowThresholdHours > 0 {
		rec.Reasons = append(rec.Reasons, "hourly_hit_rate_below_threshold")
	}
	if rec.SpikeDetected {
		rec.Reasons = append(rec.Reasons, "traffic_spike_detected")
	}
	if len(rec.Reasons) == 0 {
		rec.Recommended = true
		rec.Decision = "recommend_enable_exact_cache"
	}
	return rec, nil
}

func (c *ResponseCache) WaitOrClaimInflight(ctx context.Context, decision ResponseCacheDecision) (*ResponseCacheEntry, bool, bool) {
	if c == nil || !c.cfg.SingleflightEnabled || !decision.ExactEnabled || decision.Key == "" || c.rdb == nil {
		return nil, false, false
	}
	if c.temporarilyBypassed() {
		return nil, false, false
	}
	waitTimeout := time.Duration(c.cfg.SingleflightWaitTimeoutMs) * time.Millisecond
	if waitTimeout <= 0 {
		return nil, false, false
	}
	claimCtx, cancel := c.shortContext(ctx)
	claimed, err := c.rdb.SetNX(claimCtx, c.flightKey(decision.Key), "1", c.flightTTL(waitTimeout)).Result()
	cancel()
	if err != nil {
		c.markRedisFailure(err)
		return nil, false, false
	}
	if claimed {
		return nil, false, claimed
	}
	deadline := time.NewTimer(waitTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false, false
		case <-deadline.C:
			return nil, false, false
		case <-ticker.C:
			if c.temporarilyBypassed() {
				return nil, false, false
			}
			if entry, ok := c.Lookup(ctx, decision); ok {
				return entry, true, false
			}
		}
	}
}

func (c *ResponseCache) ReleaseInflightAsync(decision ResponseCacheDecision) {
	if c == nil || !decision.ExactEnabled || decision.Key == "" || c.rdb == nil {
		return
	}
	go func() {
		ctx, cancel := c.shortContext(context.Background())
		defer cancel()
		if err := c.rdb.Del(ctx, c.flightKey(decision.Key)).Err(); err != nil {
			c.markRedisFailure(err)
		}
	}()
}

func (c *ResponseCache) MaxCaptureBytes() int {
	if c == nil || c.cfg.MaxValueBytes <= 0 {
		return 0
	}
	return c.cfg.MaxValueBytes
}

func (c *ResponseCache) shortContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := time.Duration(c.cfg.RedisTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Millisecond
	}
	return context.WithTimeout(ctx, timeout)
}

func (c *ResponseCache) cacheKey(req ResponseCacheRequest) (string, bool) {
	if !gjson.ValidBytes(req.Body) {
		return "", false
	}
	normalized := canonicalRawJSON(string(req.Body))
	if normalized == "" {
		return "", false
	}
	scope := "global"
	if req.GroupID != nil {
		scope = "group:" + strconv.FormatInt(*req.GroupID, 10)
	}
	seed := strings.Join([]string{
		strings.TrimSpace(req.Protocol),
		strings.TrimSpace(req.Endpoint),
		strings.TrimSpace(req.Model),
		scope,
		normalized,
	}, "\n")
	sum := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%x", sum[:]), true
}

func (c *ResponseCache) scopeAllowsExact(req ResponseCacheRequest) bool {
	if hasID(req.APIKeyID, c.cfg.AllowedAPIKeyIDs) {
		return true
	}
	if req.GroupID != nil && hasID(*req.GroupID, c.cfg.AllowedGroupIDs) {
		return true
	}
	if matchStringList(req.Model, c.cfg.AllowedModels) {
		return true
	}
	return len(c.cfg.AllowedAPIKeyIDs) == 0 &&
		len(c.cfg.AllowedGroupIDs) == 0 &&
		len(c.cfg.AllowedModels) == 0
}

func (c *ResponseCache) keyPrefix() string {
	prefix := strings.TrimSpace(c.cfg.KeyPrefix)
	if prefix == "" {
		return "uado:rc:"
	}
	return prefix
}

func (c *ResponseCache) valueKey(hash string) string {
	return c.keyPrefix() + "value:" + hash
}

func (c *ResponseCache) seenKey(hash string) string {
	return c.keyPrefix() + "seen:" + hash
}

func (c *ResponseCache) monitorSeenKey(hash string) string {
	return c.keyPrefix() + "monitor_seen:" + hash
}

func (c *ResponseCache) flightKey(hash string) string {
	return c.keyPrefix() + "flight:" + hash
}

func (c *ResponseCache) flightTTL(waitTimeout time.Duration) time.Duration {
	if waitTimeout <= 0 {
		return 2 * time.Second
	}
	ttl := waitTimeout * 20
	if ttl < 2*time.Second {
		ttl = 2 * time.Second
	}
	if ttl > 30*time.Second {
		ttl = 30 * time.Second
	}
	return ttl
}

func (c *ResponseCache) statsKey(kind string) string {
	day := time.Now().Format("20060102")
	return c.keyPrefix() + "stats:" + day + ":" + kind
}

func (c *ResponseCache) hourlyStatsKey(kind string, hour time.Time) string {
	return c.keyPrefix() + "stats:hour:" + hour.UTC().Format("2006010215") + ":" + kind
}

func (c *ResponseCache) statsTTL() time.Duration {
	ttl := time.Duration(c.cfg.ShadowTTLSeconds) * time.Second
	if ttl < 24*time.Hour {
		ttl = 24 * time.Hour
	}
	if c.cfg.RecommendationWindowHours > 0 {
		windowTTL := time.Duration(c.cfg.RecommendationWindowHours) * time.Hour
		if ttl < windowTTL {
			ttl = windowTTL
		}
	}
	return ttl + 24*time.Hour
}

func (c *ResponseCache) temporarilyBypassed() bool {
	if c == nil {
		return true
	}
	until := c.redisBypassUntil.Load()
	return until > 0 && time.Now().UnixNano() < until
}

func (c *ResponseCache) markRedisFailure(err error) {
	if c == nil || err == nil || errors.Is(err, redis.Nil) {
		return
	}
	c.redisBypassUntil.Store(time.Now().Add(time.Second).UnixNano())
}

func (c *ResponseCache) normalizeRecommendationOptions(opts ResponseCacheRecommendationOptions) ResponseCacheRecommendationOptions {
	if c != nil {
		if opts.WindowHours <= 0 {
			opts.WindowHours = c.cfg.RecommendationWindowHours
		}
		if opts.MinCandidates <= 0 {
			opts.MinCandidates = c.cfg.RecommendationMinCandidates
		}
		if opts.HitRateThreshold <= 0 {
			opts.HitRateThreshold = c.cfg.RecommendationHitRateThreshold
		}
		if opts.MinObservedHours <= 0 {
			opts.MinObservedHours = c.cfg.RecommendationMinObservedHours
		}
		if opts.MaxSpikeRatio <= 0 {
			opts.MaxSpikeRatio = c.cfg.RecommendationMaxSpikeRatio
		}
	}
	if opts.WindowHours <= 0 {
		opts.WindowHours = 24
	}
	if opts.WindowHours > 168 {
		opts.WindowHours = 168
	}
	if opts.HitRateThreshold <= 0 {
		opts.HitRateThreshold = 0.20
	}
	if opts.HitRateThreshold > 1 {
		opts.HitRateThreshold = 1
	}
	if opts.MinCandidates < 0 {
		opts.MinCandidates = 0
	}
	if opts.MinObservedHours <= 0 {
		opts.MinObservedHours = opts.WindowHours
	}
	if opts.MinObservedHours > opts.WindowHours {
		opts.MinObservedHours = opts.WindowHours
	}
	if opts.MaxSpikeRatio < 0 {
		opts.MaxSpikeRatio = 0
	}
	return opts
}

func (c *ResponseCache) isMonitorRequest(req ResponseCacheRequest) bool {
	if c == nil {
		return false
	}
	if hasID(req.APIKeyID, c.cfg.MonitorAPIKeyIDs) {
		return true
	}
	return req.GroupID != nil && hasID(*req.GroupID, c.cfg.MonitorGroupIDs)
}

func redisInt64(v any) int64 {
	switch x := v.(type) {
	case nil:
		return 0
	case int64:
		return x
	case int:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return n
	case []byte:
		n, _ := strconv.ParseInt(strings.TrimSpace(string(x)), 10, 64)
		return n
	default:
		return 0
	}
}

func shouldBypassCacheControl(h http.Header) bool {
	if h == nil {
		return false
	}
	for _, v := range h.Values("Cache-Control") {
		lower := strings.ToLower(v)
		if strings.Contains(lower, "no-cache") || strings.Contains(lower, "no-store") {
			return true
		}
	}
	for _, k := range []string{"x-uado-cache-control", "x-response-cache-control"} {
		for _, v := range h.Values(k) {
			lower := strings.ToLower(strings.TrimSpace(v))
			if lower == "bypass" || lower == "no-cache" || lower == "no-store" {
				return true
			}
		}
	}
	return false
}

func isDeterministicRequest(body []byte) bool {
	temperature := gjson.GetBytes(body, "temperature")
	if !temperature.Exists() || temperature.Type != gjson.Number || temperature.Num != 0 {
		return false
	}
	if topP := gjson.GetBytes(body, "top_p"); topP.Exists() && (topP.Type != gjson.Number || (topP.Num != 0 && topP.Num != 1)) {
		return false
	}
	return true
}

func nonDeterministicReason(deterministic, exactAllowed bool) string {
	if deterministic || exactAllowed {
		return ""
	}
	return "non_deterministic"
}

func looksLikeImageOrToolRequest(endpoint string, body []byte) bool {
	path := strings.ToLower(endpoint)
	if strings.Contains(path, "image") {
		return true
	}
	if gjson.GetBytes(body, "tools").Exists() || gjson.GetBytes(body, "tool_choice").Exists() ||
		gjson.GetBytes(body, "functions").Exists() || gjson.GetBytes(body, "function_call").Exists() {
		return true
	}
	raw := strings.ToLower(string(body))
	return strings.Contains(raw, `"image_url"`) ||
		strings.Contains(raw, `"input_image"`) ||
		strings.Contains(raw, `"image_generation"`) ||
		strings.Contains(raw, `"file_id"`)
}

func approxPromptChars(body []byte) int {
	var total int
	paths := []string{
		"input",
		"instructions",
		"system",
		"messages.#.content",
		"messages.#.content.#.text",
		"messages.#.content.#.input_text",
	}
	for _, path := range paths {
		total += promptCharsFromResult(gjson.GetBytes(body, path))
	}
	if total == 0 {
		total = len(bytes.TrimSpace(body))
	}
	return total
}

func promptCharsFromResult(r gjson.Result) int {
	if !r.Exists() {
		return 0
	}
	switch {
	case r.IsArray():
		total := 0
		for _, item := range r.Array() {
			total += promptCharsFromResult(item)
		}
		return total
	case r.IsObject():
		return len(r.Raw)
	case r.Type == gjson.String:
		return len([]rune(strings.TrimSpace(r.String())))
	default:
		return len(r.Raw)
	}
}

func canonicalRawJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	normalized := normalizeJSONValue(v)
	out, err := json.Marshal(normalized)
	if err != nil {
		return raw
	}
	return string(out)
}

func normalizeJSONValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(x))
		for _, k := range keys {
			out[k] = normalizeJSONValue(x[k])
		}
		return out
	case []any:
		for i := range x {
			x[i] = normalizeJSONValue(x[i])
		}
		return x
	default:
		return v
	}
}

func encodeResponseCacheEntry(entry *ResponseCacheEntry) ([]byte, error) {
	if entry == nil {
		return nil, nil
	}
	header := make(map[string][]string, len(entry.Header))
	for k, vals := range entry.Header {
		if shouldSkipCachedHeader(k) {
			continue
		}
		header[k] = append([]string(nil), vals...)
	}
	payload := responseCachePayload{
		StatusCode: entry.StatusCode,
		Header:     header,
		Body:       base64.StdEncoding.EncodeToString(entry.Body),
		StoredAt:   entry.StoredAt.Unix(),
	}
	return json.Marshal(payload)
}

func decodeResponseCacheEntry(raw []byte) (*ResponseCacheEntry, error) {
	var payload responseCachePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	body, err := base64.StdEncoding.DecodeString(payload.Body)
	if err != nil {
		return nil, err
	}
	header := make(http.Header, len(payload.Header))
	for k, vals := range payload.Header {
		if shouldSkipCachedHeader(k) {
			continue
		}
		header[k] = append([]string(nil), vals...)
	}
	return &ResponseCacheEntry{
		StatusCode: payload.StatusCode,
		Header:     header,
		Body:       body,
		StoredAt:   time.Unix(payload.StoredAt, 0),
	}, nil
}

func shouldSkipCachedHeader(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	switch lower {
	case "content-length", "connection", "transfer-encoding", "date", "server",
		"x-uado-cache", "x-request-id", "x-openai-request-id", "request-id",
		"cf-ray", "openai-processing-ms":
		return true
	default:
		return strings.HasPrefix(lower, "x-ratelimit-")
	}
}

func hasID(id int64, ids []int64) bool {
	if id <= 0 {
		return false
	}
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

func matchStringList(value string, patterns []string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		p := strings.ToLower(strings.TrimSpace(pattern))
		if p == "" {
			continue
		}
		if p == "*" || p == value {
			return true
		}
		if strings.HasSuffix(p, "*") && strings.HasPrefix(value, strings.TrimSuffix(p, "*")) {
			return true
		}
	}
	return false
}
