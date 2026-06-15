package service

import (
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/redis/go-redis/v9"
)

func TestResponseCacheDecideExactSafetyRules(t *testing.T) {
	rc := newTestResponseCache()

	tests := []struct {
		name        string
		body        []byte
		wantEnabled bool
		wantExact   bool
		wantShadow  bool
		wantReason  string
	}{
		{
			name:       "short prompt bypasses probes",
			body:       []byte(`{"model":"gpt-5","temperature":0,"messages":[{"role":"user","content":"hi"}]}`),
			wantReason: "prompt_too_short",
		},
		{
			name:       "short text block prompt bypasses probes",
			body:       []byte(`{"model":"claude-sonnet","temperature":0,"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}`),
			wantReason: "prompt_too_short",
		},
		{
			name:       "short openai text block prompt bypasses probes",
			body:       []byte(`{"model":"gpt-5","temperature":0,"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}`),
			wantReason: "prompt_too_short",
		},
		{
			name:        "implicit temperature can be counted in shadow but not exact cache",
			body:        []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"please summarize this fairly long paragraph"}]}`),
			wantEnabled: true,
			wantShadow:  true,
			wantReason:  "non_deterministic",
		},
		{
			name:        "long text block prompt is eligible",
			body:        []byte(`{"model":"claude-sonnet","temperature":0,"messages":[{"role":"user","content":[{"type":"text","text":"please answer this deterministic request"}]}]}`),
			wantEnabled: true,
			wantExact:   true,
			wantShadow:  true,
		},
		{
			name:        "long openai text block prompt is eligible",
			body:        []byte(`{"model":"gpt-5","temperature":0,"messages":[{"role":"user","content":[{"type":"text","text":"please answer this deterministic request"}]}]}`),
			wantEnabled: true,
			wantExact:   true,
			wantShadow:  true,
		},
		{
			name:       "tools are not cached",
			body:       []byte(`{"model":"gpt-5","temperature":0,"tools":[{"type":"function","function":{"name":"x"}}],"messages":[{"role":"user","content":"please call the tool with this long enough prompt"}]}`),
			wantReason: "unsupported_request",
		},
		{
			name:        "deterministic medium prompt can use exact cache",
			body:        []byte(`{"model":"gpt-5","temperature":0,"messages":[{"role":"user","content":"please answer this deterministic request"}]}`),
			wantEnabled: true,
			wantExact:   true,
			wantShadow:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rc.Decide(ResponseCacheRequest{
				Endpoint: "/v1/chat/completions",
				Protocol: "test",
				Model:    "gpt-5",
				Body:     tt.body,
				Headers:  http.Header{},
			})
			if got.Enabled != tt.wantEnabled || got.ExactEnabled != tt.wantExact || got.ShadowEnabled != tt.wantShadow || got.Reason != tt.wantReason {
				t.Fatalf("Decide() = %+v, want enabled=%v exact=%v shadow=%v reason=%q", got, tt.wantEnabled, tt.wantExact, tt.wantShadow, tt.wantReason)
			}
		})
	}
}

func TestResponseCacheDecideStreamShadowOnly(t *testing.T) {
	rc := newTestResponseCache()
	got := rc.Decide(ResponseCacheRequest{
		Endpoint: "/v1/messages",
		Protocol: "test",
		Model:    "claude-sonnet",
		Stream:   true,
		Body:     []byte(`{"model":"claude-sonnet","temperature":0,"messages":[{"role":"user","content":"please answer this deterministic request"}]}`),
		Headers:  http.Header{},
	})
	if !got.Enabled || got.ExactEnabled || !got.ShadowEnabled || got.Reason != "stream_shadow_only" {
		t.Fatalf("Decide() = %+v, want shadow-only stream decision", got)
	}
}

func TestResponseCacheDecisionCarriesAnonymizedStatsMetadata(t *testing.T) {
	rc := newTestResponseCache()
	groupID := int64(42)
	got := rc.Decide(ResponseCacheRequest{
		Endpoint: " /v1/chat/completions ",
		Protocol: " openai ",
		Model:    " gpt-5 ",
		Body:     []byte(`{"model":"gpt-5","temperature":0,"messages":[{"role":"user","content":"please answer this deterministic request"}]}`),
		APIKeyID: 7,
		GroupID:  &groupID,
		Headers:  http.Header{},
	})
	groupID = 99

	if got.Endpoint != "/v1/chat/completions" || got.Protocol != "openai" || got.Model != "gpt-5" || got.APIKeyID != 7 {
		t.Fatalf("metadata not normalized: %+v", got)
	}
	if got.GroupID == nil || *got.GroupID != 42 {
		t.Fatalf("group metadata not cloned, got %+v", got.GroupID)
	}
}

func TestResponseCacheKeyIsStableForJSONFieldOrder(t *testing.T) {
	rc := newTestResponseCache()
	reqA := ResponseCacheRequest{
		Endpoint: "/v1/chat/completions",
		Protocol: "test",
		Model:    "gpt-5",
		Body:     []byte(`{"model":"gpt-5","temperature":0,"messages":[{"role":"user","content":"please answer this deterministic request"}]}`),
		Headers:  http.Header{},
	}
	reqB := ResponseCacheRequest{
		Endpoint: "/v1/chat/completions",
		Protocol: "test",
		Model:    "gpt-5",
		Body:     []byte(`{"messages":[{"content":"please answer this deterministic request","role":"user"}],"temperature":0,"model":"gpt-5"}`),
		Headers:  http.Header{},
	}
	a := rc.Decide(reqA)
	b := rc.Decide(reqB)
	if a.Key == "" || a.Key != b.Key {
		t.Fatalf("keys not stable: %q vs %q", a.Key, b.Key)
	}
}

func TestResponseCacheSummarizeKeyStatsDetectsConcentration(t *testing.T) {
	items := []*ResponseCacheKeyStatsItem{
		{CacheKeyHash: "a", TotalCount: 80, HitCount: 60},
		{CacheKeyHash: "b", TotalCount: 30, HitCount: 20},
		{CacheKeyHash: "c", TotalCount: 20, HitCount: 10},
		{CacheKeyHash: "d", TotalCount: 10, HitCount: 3},
		{CacheKeyHash: "e", TotalCount: 10, HitCount: 2},
		{CacheKeyHash: "f", TotalCount: 10, HitCount: 1},
	}

	summary := summarizeKeyStats(items, 160, 96, 0.50, 0.80)

	if summary.UniqueKeys != 6 || summary.TrackedKeys != 6 {
		t.Fatalf("unique keys = %d/%d, want 6/6", summary.UniqueKeys, summary.TrackedKeys)
	}
	if summary.Top1HitShare < 0.62 || summary.Top1HitShare > 0.63 {
		t.Fatalf("top1 share = %v, want about 0.625", summary.Top1HitShare)
	}
	if summary.Top5HitShare < 0.98 || summary.Top5HitShare > 0.99 {
		t.Fatalf("top5 share = %v, want about 0.989", summary.Top5HitShare)
	}
	if !summary.ConcentrationDetected {
		t.Fatalf("expected concentration to be detected")
	}
}

func TestResponseCacheKeyStatsMatchesFilters(t *testing.T) {
	groupID := int64(12)
	item := &ResponseCacheKeyStatsItem{
		Model:        "gpt-5",
		APIKeyID:     34,
		GroupID:      &groupID,
		Monitor:      false,
		LastSeenAt:   time.Now().Add(-time.Hour),
		TotalCount:   10,
		HitCount:     2,
		CacheKeyHash: "abcdef123456",
	}

	if !responseCacheKeyStatsMatches(item, ResponseCacheKeyStatsOptions{
		WindowHours: 24,
		Model:       "GPT-5",
		APIKeyID:    34,
		GroupID:     12,
		GroupIDSet:  true,
		Monitor:     "no",
	}) {
		t.Fatalf("expected item to match exact filters")
	}
	if responseCacheKeyStatsMatches(item, ResponseCacheKeyStatsOptions{WindowHours: 1, Model: "gpt-5", Monitor: "yes"}) {
		t.Fatalf("expected item to be excluded by monitor filter")
	}
	if responseCacheKeyStatsMatches(item, ResponseCacheKeyStatsOptions{WindowHours: 1, Model: "claude", Monitor: "no"}) {
		t.Fatalf("expected item to be excluded by model filter")
	}
}

func TestResponseCacheKeyStatsCountsUseRequestedWindow(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)
	values := map[string]string{
		"total:" + responseCacheStatsHour(now):                   "5",
		"hit:" + responseCacheStatsHour(now):                     "3",
		"total:" + responseCacheStatsHour(now.Add(-time.Hour)):   "4",
		"hit:" + responseCacheStatsHour(now.Add(-time.Hour)):     "2",
		"total:" + responseCacheStatsHour(now.Add(-2*time.Hour)): "100",
		"hit:" + responseCacheStatsHour(now.Add(-2*time.Hour)):   "90",
	}

	total, hit := keyStatsCountsFromHash(values, 2, now)

	if total != 9 || hit != 5 {
		t.Fatalf("counts = %d/%d, want 9/5 for last two hours", total, hit)
	}
}

func TestResponseCacheKeyStatsCountsFallbackToLegacyTotals(t *testing.T) {
	total, hit := keyStatsCountsFromHash(map[string]string{
		"total_count": "12",
		"hit_count":   "4",
	}, 24, time.Now())

	if total != 12 || hit != 4 {
		t.Fatalf("legacy counts = %d/%d, want 12/4", total, hit)
	}
}

func newTestResponseCache() *ResponseCache {
	cfg := &config.Config{}
	cfg.ResponseCache = config.ResponseCacheConfig{
		Enabled:                       true,
		ShadowEnabled:                 true,
		KeyPrefix:                     "test:rc:",
		TTLSeconds:                    300,
		ShadowTTLSeconds:              3600,
		RedisTimeoutMs:                10,
		MaxBodyBytes:                  64 * 1024,
		MaxValueBytes:                 1024 * 1024,
		MinPromptChars:                16,
		MaxPromptChars:                12000,
		SingleflightEnabled:           true,
		SingleflightWaitTimeoutMs:     150,
		PrefixCacheEnabled:            true,
		RecommendationEnabled:         true,
		RecommendationWindowHours:     72,
		RecommendationMinUniqueKeys:   20,
		RecommendationTop1MaxHitShare: 0.50,
		RecommendationTop5MaxHitShare: 0.80,
	}
	return &ResponseCache{
		cfg: cfg.ResponseCache,
		rdb: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}),
	}
}
