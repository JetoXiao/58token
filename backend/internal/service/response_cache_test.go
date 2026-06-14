package service

import (
	"net/http"
	"testing"

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
			name:        "implicit temperature can be counted in shadow but not exact cache",
			body:        []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"please summarize this fairly long paragraph"}]}`),
			wantEnabled: true,
			wantShadow:  true,
			wantReason:  "non_deterministic",
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

func newTestResponseCache() *ResponseCache {
	cfg := &config.Config{}
	cfg.ResponseCache = config.ResponseCacheConfig{
		Enabled:                   true,
		ShadowEnabled:             true,
		KeyPrefix:                 "test:rc:",
		TTLSeconds:                300,
		ShadowTTLSeconds:          3600,
		RedisTimeoutMs:            10,
		MaxBodyBytes:              64 * 1024,
		MaxValueBytes:             1024 * 1024,
		MinPromptChars:            16,
		MaxPromptChars:            12000,
		SingleflightEnabled:       true,
		SingleflightWaitTimeoutMs: 150,
		PrefixCacheEnabled:        true,
	}
	return &ResponseCache{
		cfg: cfg.ResponseCache,
		rdb: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}),
	}
}
