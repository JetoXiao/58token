package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type upstreamPricingSequence struct {
	responses []*http.Response
	requests  []*http.Request
}

func (u *upstreamPricingSequence) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.requests = append(u.requests, req)
	if len(u.responses) == 0 {
		return nil, io.EOF
	}
	response := u.responses[0]
	u.responses = u.responses[1:]
	return response, nil
}

func (u *upstreamPricingSequence) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}

func TestBuildUpstreamPricingCandidates(t *testing.T) {
	t.Parallel()

	candidates, err := buildUpstreamPricingCandidates("https://gateway.example.com/v1")
	require.NoError(t, err)
	require.Equal(t, []upstreamPricingCandidate{
		{Source: "newapi", URL: "https://gateway.example.com/api/pricing"},
		{Source: "sub2api", URL: "https://gateway.example.com/api/v1/public/model-pricing"},
	}, candidates)
}

func TestParseUpstreamPricingSnapshot(t *testing.T) {
	t.Parallel()

	snapshot, err := parseUpstreamPricingSnapshot([]byte(`{
		"success": true,
		"pricing_version": "version-1",
		"group_ratio": {"default": 0.15, "vip": "0.2"},
		"usable_group": {"default": "默认分组", "vip": "VIP"},
		"data": [{
			"model_name": "gpt-test",
			"tags": "Reasoning,Tools",
			"quota_type": 0,
			"model_ratio": 1.25,
			"completion_ratio": 6,
			"cache_ratio": 0.1,
			"create_cache_ratio": 1.25,
			"enable_groups": ["default", "vip"],
			"supported_endpoint_types": ["openai"]
		}]
	}`))

	require.NoError(t, err)
	require.Equal(t, "version-1", snapshot.PricingVersion)
	require.Equal(t, map[string]float64{"default": 0.15, "vip": 0.2}, snapshot.GroupRatios)
	require.Equal(t, "默认分组", snapshot.GroupNames["default"])
	require.Len(t, snapshot.Models, 1)
	require.Equal(t, "gpt-test", snapshot.Models[0].ModelName)
	require.Equal(t, 1.25, *snapshot.Models[0].ModelRatio)
	require.Equal(t, 6.0, *snapshot.Models[0].CompletionRatio)
	require.Equal(t, []string{"default", "vip"}, snapshot.Models[0].EnabledGroups)
}

func TestParseEffectiveUpstreamGroups(t *testing.T) {
	 t.Parallel()

	ratios, names, err := parseEffectiveUpstreamGroups([]byte(`{
		"success": true,
		"data": {"gpt-pro": {"base_ratio": 0.45, "desc": "纯pro池，极速可靠", "ratio": 0.15}}
	}`), true)
	require.NoError(t, err)
	require.Equal(t, 0.15, ratios["gpt-pro"])
	require.Equal(t, "纯pro池，极速可靠", names["gpt-pro"])
}

func TestParseSub2APIModelPlaza(t *testing.T) {
	t.Parallel()

	ratios, names, err := parseSub2APIModelPlaza([]byte(`{
		"data": {"groups": [
			{"id": 7, "name": "Claude Pro", "description": "pro", "rate_multiplier": 0.4, "user_rate_multiplier": 0.15}
		]}
	}`))
	require.NoError(t, err)
	require.Equal(t, 0.15, ratios["7"])
	require.Equal(t, "Claude Pro", names["7"])
}

func TestFetchUpstreamPricingFallsBackToSub2API(t *testing.T) {
	t.Parallel()

	upstream := &upstreamPricingSequence{responses: []*http.Response{
		{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		},
		{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"success":true,
				"group_ratio":{"default":0.3},
				"data":[{"model_name":"claude-test","model_ratio":1,"completion_ratio":5}]
			}`)),
			Header: http.Header{"Content-Type": []string{"application/json"}},
		},
	}}
	svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}

	snapshot, err := svc.FetchUpstreamPricing(context.Background(), &Account{
		ID:       12,
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "upstream-key", "base_url": "https://gateway.example.com/v1", "upstream_group_key": "default",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "sub2api", snapshot.Source)
	require.Equal(t, "https://gateway.example.com/api/v1/public/model-pricing", snapshot.Endpoint)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "Bearer upstream-key", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, 0.3, snapshot.GroupRatios["default"])
}

func TestFetchUpstreamPricingRejectsNonAPIKeyAccount(t *testing.T) {
	t.Parallel()

	svc := &AccountTestService{httpUpstream: &upstreamPricingSequence{}, cfg: upstreamModelSyncTestConfig()}
	_, err := svc.FetchUpstreamPricing(context.Background(), &Account{Type: AccountTypeOAuth})
	require.Error(t, err)

	var pricingErr *UpstreamPricingError
	require.ErrorAs(t, err, &pricingErr)
	require.Equal(t, UpstreamPricingErrorUnsupported, pricingErr.Kind)
}

func TestFetchUpstreamPricingCachesSuccessfulSnapshot(t *testing.T) {
	t.Parallel()

	upstream := &upstreamPricingSequence{responses: []*http.Response{{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"success":true,"group_ratio":{"default":0.2},"data":[{"model_name":"gpt-test","model_ratio":1}]}`)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}}}
	svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
	account := &Account{
		ID:       13,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "upstream-key", "base_url": "https://cached.example.com/v1", "upstream_group_key": "default",
		},
	}

	first, err := svc.FetchUpstreamPricing(context.Background(), account)
	require.NoError(t, err)
	second, err := svc.FetchUpstreamPricing(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Len(t, upstream.requests, 1)
}

func TestFetchUpstreamPricingDoesNotShareCacheAcrossAccounts(t *testing.T) {
	t.Parallel()

	upstream := &upstreamPricingSequence{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"group_ratio":{"group-a":0.2},"data":[{"model_name":"gpt-test","model_ratio":1}]}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		},
		{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"success":true,"group_ratio":{"group-b":0.5},"data":[{"model_name":"gpt-test","model_ratio":1}]}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		},
	}}
	svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
	accountA := &Account{ID: 21, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"api_key": "group-a-key", "base_url": "https://same-upstream.example.com/v1", "upstream_group_key": "group-a",
	}}
	accountB := &Account{ID: 22, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"api_key": "group-b-key", "base_url": "https://same-upstream.example.com/v1", "upstream_group_key": "group-b",
	}}

	first, err := svc.FetchUpstreamPricing(context.Background(), accountA)
	require.NoError(t, err)
	second, err := svc.FetchUpstreamPricing(context.Background(), accountB)
	require.NoError(t, err)

	require.Equal(t, 0.2, first.GroupRatios["group-a"])
	require.Equal(t, 0.5, second.GroupRatios["group-b"])
	require.Len(t, upstream.requests, 2)
}

func TestFetchUpstreamPricingFiltersToConfiguredGroupKey(t *testing.T) {
	t.Parallel()

	upstream := &upstreamPricingSequence{responses: []*http.Response{{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
				"success":true,
				"group_ratio":{"default":1,"gpt-pro":0.45,"vip":1.5},
				"usable_group":{"default":"默认分组","gpt-pro":"Codex Pro","vip":"VIP"},
				"data":[{"model_name":"codex-auto-review","model_ratio":1.25}]
			}`)),
		Header: http.Header{"Content-Type": []string{"application/json"}},
	}}}
	svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
	account := &Account{ID: 31, Name: "Codex Pro", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"api_key": "codex-pro-key", "base_url": "https://same-upstream.example.com/v1", "upstream_group_key": "gpt-pro",
	}}

	snapshot, err := svc.FetchUpstreamPricing(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, map[string]float64{"gpt-pro": 0.45}, snapshot.GroupRatios)
	require.Equal(t, "Codex Pro", snapshot.GroupNames["gpt-pro"])
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://same-upstream.example.com/api/pricing", upstream.requests[0].URL.String())
}

func TestFetchUpstreamPricingUsesDashboardEffectiveRatio(t *testing.T) {
	t.Parallel()

	upstream := &upstreamPricingSequence{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{"success":true,"group_ratio":{"gpt-pro":0.45},"usable_group":{"gpt-pro":"纯pro池，极速可靠"}}`)),
		},
		{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{"success":true,"data":{"gpt-pro":{"desc":"纯pro池，极速可靠","ratio":0.15}}}`)),
		},
	}}
	svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
	account := &Account{ID: 34, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"api_key": "model-key", "base_url": "https://same-upstream.example.com/v1", "upstream_group_key": "gpt-pro",
		"upstream_dashboard_access_token": "dashboard-token",
	}}

	snapshot, err := svc.FetchUpstreamPricing(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, upstreamPricingRatioScopeEffective, snapshot.RatioScope)
	require.Equal(t, 0.15, snapshot.GroupRatios["gpt-pro"])
	require.Equal(t, "Bearer dashboard-token", upstream.requests[1].Header.Get("Authorization"))
}

func TestFetchUpstreamPricingRequiresConfiguredGroupKey(t *testing.T) {
	t.Parallel()

	upstream := &upstreamPricingSequence{responses: []*http.Response{{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
			"success":true,
			"group_ratio":{"default":1,"codex-pro":0.45,"vip":1.5},
			"usable_group":{"default":"默认分组","codex-pro":"Codex Pro","vip":"VIP"},
			"data":[{"model_name":"codex-auto-review","model_ratio":1.25}]
		}`)),
		Header: http.Header{"Content-Type": []string{"application/json"}},
	}}}
	svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
	account := &Account{ID: 32, Name: "Codex account", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"api_key": "codex-pro-key", "base_url": "https://same-upstream.example.com/v1",
	}}

	_, err := svc.FetchUpstreamPricing(context.Background(), account)
	require.Error(t, err)
	require.Empty(t, upstream.requests)
}
