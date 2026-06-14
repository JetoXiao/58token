package handler

import (
	"bytes"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func newResponseCacheForHandler(cfg *config.Config) *service.ResponseCache {
	if cfg == nil || (!cfg.ResponseCache.Enabled && !cfg.ResponseCache.ShadowEnabled) {
		return nil
	}
	rdb := repository.InitResponseCacheRedis(cfg)
	return service.NewResponseCache(cfg, rdb)
}

func NewResponseCacheForAdmin(cfg *config.Config) *service.ResponseCache {
	if cfg == nil || !cfg.ResponseCache.RecommendationEnabled {
		return nil
	}
	rdb := repository.InitResponseCacheRedis(cfg)
	return service.NewResponseCache(cfg, rdb)
}

func responseCacheGroupID(apiKey *service.APIKey) *int64 {
	if apiKey == nil {
		return nil
	}
	return apiKey.GroupID
}

func responseCacheRequest(c *gin.Context, endpoint, protocol, model string, stream bool, body []byte, apiKey *service.APIKey) service.ResponseCacheRequest {
	var apiKeyID int64
	if apiKey != nil {
		apiKeyID = apiKey.ID
	}
	var headers http.Header
	if c != nil && c.Request != nil {
		headers = c.Request.Header
	}
	return service.ResponseCacheRequest{
		Endpoint: endpoint,
		Protocol: protocol,
		Model:    model,
		Stream:   stream,
		Body:     body,
		APIKeyID: apiKeyID,
		GroupID:  responseCacheGroupID(apiKey),
		Headers:  headers,
	}
}

func tryResponseCacheBeforeForward(
	c *gin.Context,
	rc *service.ResponseCache,
	req service.ResponseCacheRequest,
) (service.ResponseCacheDecision, bool, bool) {
	decision := rc.Decide(req)
	if decision.Enabled {
		rc.ObserveShadowAsync(decision)
		if decision.ExactEnabled {
			if entry, ok := rc.Lookup(c.Request.Context(), decision); ok {
				service.SetOpsLatencyMs(c, service.OpsTimeToFirstTokenMsKey, 0)
				replayResponseCacheEntry(c, entry)
				return decision, false, true
			}
			if entry, shared, claimed := rc.WaitOrClaimInflight(c.Request.Context(), decision); shared {
				service.SetOpsLatencyMs(c, service.OpsTimeToFirstTokenMsKey, 0)
				replayResponseCacheEntry(c, entry)
				return decision, false, true
			} else if claimed {
				markResponseCacheStatus(c, service.ResponseCacheStatusMiss)
				return decision, true, false
			}
			markResponseCacheStatus(c, service.ResponseCacheStatusMiss)
			return decision, false, false
		}
		markResponseCacheStatus(c, service.ResponseCacheStatusShadow)
		return decision, false, false
	}
	markResponseCacheBypass(c, decision.Reason)
	return decision, false, false
}

func finishResponseCacheAfterForward(
	rc *service.ResponseCache,
	decision service.ResponseCacheDecision,
	entry *service.ResponseCacheEntry,
	inflightOwner bool,
) {
	stored := false
	if entry != nil {
		stored = rc.StoreAsync(decision, entry)
	}
	if inflightOwner && !stored {
		rc.ReleaseInflightAsync(decision)
	}
}

func markResponseCacheBypass(c *gin.Context, reason string) {
	if c == nil {
		return
	}
	status := service.ResponseCacheStatusBypass
	if strings.TrimSpace(reason) != "" {
		status += "; reason=" + sanitizeCacheHeaderToken(reason)
	}
	c.Header(service.ResponseCacheHeader, status)
}

func markResponseCacheStatus(c *gin.Context, status string) {
	if c == nil || strings.TrimSpace(status) == "" {
		return
	}
	c.Header(service.ResponseCacheHeader, status)
}

func replayResponseCacheEntry(c *gin.Context, entry *service.ResponseCacheEntry) bool {
	if c == nil || entry == nil {
		return false
	}
	for key, vals := range entry.Header {
		if shouldSkipReplayHeader(key) {
			continue
		}
		for _, v := range vals {
			c.Writer.Header().Add(key, v)
		}
	}
	c.Header(service.ResponseCacheHeader, service.ResponseCacheStatusHit)
	status := entry.StatusCode
	if status <= 0 {
		status = http.StatusOK
	}
	c.Status(status)
	_, _ = c.Writer.Write(entry.Body)
	return true
}

func captureResponseForCache(c *gin.Context, maxBytes int) (*responseCacheCaptureWriter, func() *service.ResponseCacheEntry) {
	if c == nil || c.Writer == nil || maxBytes <= 0 {
		return nil, func() *service.ResponseCacheEntry { return nil }
	}
	original := c.Writer
	cw := &responseCacheCaptureWriter{
		ResponseWriter: original,
		maxBytes:       maxBytes,
	}
	c.Writer = cw
	return cw, func() *service.ResponseCacheEntry {
		c.Writer = original
		if cw.overflow.Load() {
			return nil
		}
		body := cw.body.Bytes()
		if len(body) == 0 {
			return nil
		}
		status := cw.Status()
		if status <= 0 {
			status = http.StatusOK
		}
		return &service.ResponseCacheEntry{
			StatusCode: status,
			Header:     cloneHeader(cw.Header()),
			Body:       append([]byte(nil), body...),
			StoredAt:   time.Now(),
		}
	}
}

type responseCacheCaptureWriter struct {
	gin.ResponseWriter
	mu       sync.Mutex
	body     bytes.Buffer
	maxBytes int
	overflow atomicBool
}

func (w *responseCacheCaptureWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseCacheCaptureWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *responseCacheCaptureWriter) capture(data []byte) {
	if w == nil || len(data) == 0 || w.maxBytes <= 0 || w.overflow.Load() {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.body.Len()+len(data) > w.maxBytes {
		w.overflow.Store(true)
		w.body.Reset()
		return
	}
	_, _ = w.body.Write(data)
}

type atomicBool struct {
	mu sync.Mutex
	v  bool
}

func (b *atomicBool) Load() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.v
}

func (b *atomicBool) Store(v bool) {
	b.mu.Lock()
	b.v = v
	b.mu.Unlock()
}

func cloneHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	out := make(http.Header, len(h))
	for k, vals := range h {
		out[k] = append([]string(nil), vals...)
	}
	return out
}

func shouldSkipReplayHeader(key string) bool {
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

func sanitizeCacheHeaderToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
