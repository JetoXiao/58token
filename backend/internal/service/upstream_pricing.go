package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	upstreamPricingBodyLimit int64 = 8 << 20
	upstreamPricingCacheTTL        = 30 * time.Second
)

type upstreamPricingCacheEntry struct {
	Snapshot  *UpstreamPricingSnapshot
	ExpiresAt time.Time
}

type UpstreamPricingErrorKind string

const (
	UpstreamPricingErrorConfiguration UpstreamPricingErrorKind = "configuration"
	UpstreamPricingErrorUnsupported   UpstreamPricingErrorKind = "unsupported"
	UpstreamPricingErrorUpstream      UpstreamPricingErrorKind = "upstream"
)

// UpstreamPricingError keeps upstream response details out of the admin API response.
type UpstreamPricingError struct {
	Kind    UpstreamPricingErrorKind
	Message string
	Err     error
}

func (e *UpstreamPricingError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

func (e *UpstreamPricingError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *UpstreamPricingError) SafeMessage() string {
	if e == nil || strings.TrimSpace(e.Message) == "" {
		return "Failed to fetch upstream pricing"
	}
	return e.Message
}

type UpstreamPricingModel struct {
	ModelName              string   `json:"model_name"`
	Tags                   string   `json:"tags,omitempty"`
	QuotaType              int      `json:"quota_type"`
	ModelRatio             *float64 `json:"model_ratio,omitempty"`
	CompletionRatio        *float64 `json:"completion_ratio,omitempty"`
	CacheRatio             *float64 `json:"cache_ratio,omitempty"`
	CreateCacheRatio       *float64 `json:"create_cache_ratio,omitempty"`
	ModelPrice             *float64 `json:"model_price,omitempty"`
	EnabledGroups          []string `json:"enabled_groups"`
	SupportedEndpointTypes []string `json:"supported_endpoint_types"`
}

type UpstreamPricingSnapshot struct {
	Source         string                 `json:"source"`
	Endpoint       string                 `json:"endpoint"`
	CheckedAt      time.Time              `json:"checked_at"`
	RatioScope     string                 `json:"ratio_scope"`
	PricingVersion string                 `json:"pricing_version,omitempty"`
	GroupRatios    map[string]float64     `json:"group_ratios"`
	GroupNames     map[string]string      `json:"group_names"`
	Models         []UpstreamPricingModel `json:"models"`
	Balance        *UpstreamBalance       `json:"balance,omitempty"`
}

// UpstreamBalance is an optional account balance returned by an upstream
// control-panel API. New API exposes quota units while Sub2API exposes a
// currency balance, so the unit is explicit instead of assuming dollars.
type UpstreamBalance struct {
	Amount     float64   `json:"amount"`
	RawAmount  float64   `json:"raw_amount,omitempty"`
	UsedAmount *float64  `json:"used_amount,omitempty"`
	Unit       string    `json:"unit"`
	Currency   string    `json:"currency,omitempty"`
	Source     string    `json:"source"`
	Endpoint   string    `json:"endpoint"`
	CheckedAt  time.Time `json:"checked_at"`
}

type UpstreamPricingGroup struct {
	Key   string  `json:"key"`
	Name  string  `json:"name"`
	Ratio float64 `json:"ratio"`
}

type upstreamPricingCandidate struct {
	Source string
	URL    string
}

const (
	upstreamPricingRatioScopeBase      = "base"
	upstreamPricingRatioScopeEffective = "effective"
)

// FetchUpstreamPricing probes the public pricing endpoints used by New API and Sub2API forks.
// The request is sent server-side so account credentials never need to be exposed to the browser.
func (s *AccountTestService) FetchUpstreamPricing(ctx context.Context, account *Account) (*UpstreamPricingSnapshot, error) {
	if account == nil {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Account is required"}
	}
	if account.Type != AccountTypeAPIKey {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorUnsupported, Message: "Only API-key upstream accounts support pricing lookup"}
	}
	groupKey := strings.TrimSpace(account.GetCredential("upstream_group_key"))
	if groupKey == "" {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Configure the account's upstream group first"}
	}
	snapshot, err := s.fetchUpstreamPricingCatalog(ctx, account)
	if err != nil {
		return nil, err
	}
	if _, ok := snapshot.GroupRatios[groupKey]; !ok {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorUpstream, Message: "The configured upstream group no longer exists"}
	}
	snapshot = cloneUpstreamPricingSnapshot(snapshot)
	filterUpstreamPricingToGroups(snapshot, []string{groupKey})
	return snapshot, nil
}

func cloneUpstreamPricingSnapshot(snapshot *UpstreamPricingSnapshot) *UpstreamPricingSnapshot {
	if snapshot == nil {
		return nil
	}
	clone := *snapshot
	clone.GroupRatios = make(map[string]float64, len(snapshot.GroupRatios))
	for key, ratio := range snapshot.GroupRatios {
		clone.GroupRatios[key] = ratio
	}
	clone.GroupNames = make(map[string]string, len(snapshot.GroupNames))
	for key, name := range snapshot.GroupNames {
		clone.GroupNames[key] = name
	}
	clone.Models = append([]UpstreamPricingModel(nil), snapshot.Models...)
	if snapshot.Balance != nil {
		balance := *snapshot.Balance
		if snapshot.Balance.UsedAmount != nil {
			used := *snapshot.Balance.UsedAmount
			balance.UsedAmount = &used
		}
		clone.Balance = &balance
	}
	return &clone
}

func (s *AccountTestService) FetchUpstreamPricingGroups(ctx context.Context, account *Account) ([]UpstreamPricingGroup, error) {
	snapshot, err := s.fetchUpstreamPricingCatalog(ctx, account)
	if err != nil {
		return nil, err
	}
	groups := make([]UpstreamPricingGroup, 0, len(snapshot.GroupRatios))
	for key, ratio := range snapshot.GroupRatios {
		name := strings.TrimSpace(snapshot.GroupNames[key])
		if name == "" {
			name = key
		}
		groups = append(groups, UpstreamPricingGroup{Key: key, Name: name, Ratio: ratio})
	}
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Name) < strings.ToLower(groups[j].Name)
	})
	return groups, nil
}

func (s *AccountTestService) fetchUpstreamPricingCatalog(ctx context.Context, account *Account) (*UpstreamPricingSnapshot, error) {
	if s == nil || s.httpUpstream == nil {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Upstream HTTP client is not configured"}
	}
	if account == nil {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Account is required"}
	}
	if account.Type != AccountTypeAPIKey {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorUnsupported, Message: "Only API-key upstream accounts support pricing lookup"}
	}

	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	if baseURL == "" {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "The account does not have an upstream base URL"}
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Invalid upstream base URL", Err: err}
	}

	candidates, err := buildUpstreamPricingCandidates(normalizedBaseURL)
	if err != nil {
		return nil, &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Invalid upstream pricing URL", Err: err}
	}
	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	dashboardToken := normalizeUpstreamDashboardToken(account.GetCredential("upstream_dashboard_access_token"))
	dashboardUserID := strings.TrimSpace(account.GetCredential("upstream_dashboard_user_id"))
	credentialDigest := sha256.Sum256([]byte(apiKey + "\x00" + dashboardToken + "\x00" + dashboardUserID))
	cacheKey := fmt.Sprintf("%d|%s|%x", account.ID, candidates[0].URL, credentialDigest)
	if cached := s.getCachedUpstreamPricing(cacheKey, time.Now()); cached != nil {
		return cached, nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var lastErr error
	for _, candidate := range candidates {
		req, reqErr := http.NewRequestWithContext(requestCtx, http.MethodGet, candidate.URL, nil)
		if reqErr != nil {
			lastErr = reqErr
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "58Token/1.0")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}

		resp, requestErr := s.doUpstreamModelsRequest(req, upstreamModelsProxyURL(account), account)
		if requestErr != nil {
			lastErr = requestErr
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, upstreamPricingBodyLimit+1))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if int64(len(body)) > upstreamPricingBodyLimit {
			lastErr = fmt.Errorf("response exceeds %d bytes", upstreamPricingBodyLimit)
			continue
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			lastErr = fmt.Errorf("pricing endpoint returned HTTP %d", resp.StatusCode)
			continue
		}

		snapshot, parseErr := parseUpstreamPricingSnapshot(body)
		if parseErr != nil {
			lastErr = parseErr
			continue
		}
		snapshot.Source = candidate.Source
		snapshot.Endpoint = candidate.URL
		snapshot.CheckedAt = time.Now().UTC()
		snapshot.RatioScope = upstreamPricingRatioScopeBase
		if dashboardToken != "" {
			effectiveRatios, effectiveNames, effectiveEndpoint, effectiveErr := s.fetchEffectiveUpstreamGroupRatios(requestCtx, normalizedBaseURL, dashboardToken, dashboardUserID, account)
			if effectiveErr == nil {
				snapshot.GroupRatios = effectiveRatios
				for key, name := range effectiveNames {
					snapshot.GroupNames[key] = name
				}
				snapshot.RatioScope = upstreamPricingRatioScopeEffective
				snapshot.Endpoint = effectiveEndpoint
			}
			// Effective group APIs are optional across New API/Sub2API forks. A
			// valid public catalog is still useful, so do not fail the whole row
			// when only the dashboard-specific endpoint is unavailable.
			if balance, balanceErr := s.fetchUpstreamBalance(requestCtx, normalizedBaseURL, dashboardToken, dashboardUserID, account); balanceErr == nil {
				snapshot.Balance = balance
			}
		}
		s.cacheUpstreamPricing(cacheKey, snapshot, time.Now().Add(upstreamPricingCacheTTL))
		return snapshot, nil
	}

	if lastErr == nil {
		lastErr = errors.New("no pricing endpoint candidates were available")
	}
	// Some Sub2API deployments disable the public pricing endpoint entirely.
	// Their authenticated model-plaza endpoint still exposes the effective
	// group multipliers, so use it as the catalog when a dashboard token exists.
	if dashboardToken != "" {
		if groupSnapshot, groupErr := s.fetchSub2APIGroupCatalog(ctx, normalizedBaseURL, dashboardToken, account); groupErr == nil {
			groupSnapshot.CheckedAt = time.Now().UTC()
			if balance, balanceErr := s.fetchUpstreamBalance(ctx, normalizedBaseURL, dashboardToken, "", account); balanceErr == nil {
				groupSnapshot.Balance = balance
			}
			s.cacheUpstreamPricing(cacheKey, groupSnapshot, time.Now().Add(upstreamPricingCacheTTL))
			return groupSnapshot, nil
		} else if plazaSnapshot, plazaErr := s.fetchSub2APIModelPlaza(ctx, normalizedBaseURL, dashboardToken, account); plazaErr == nil {
			plazaSnapshot.CheckedAt = time.Now().UTC()
			if balance, balanceErr := s.fetchUpstreamBalance(ctx, normalizedBaseURL, dashboardToken, "", account); balanceErr == nil {
				plazaSnapshot.Balance = balance
			}
			s.cacheUpstreamPricing(cacheKey, plazaSnapshot, time.Now().Add(upstreamPricingCacheTTL))
			return plazaSnapshot, nil
		} else {
			lastErr = fmt.Errorf("Sub2API groups endpoint: %v; model plaza: %w", groupErr, plazaErr)
		}
	}
	return nil, &UpstreamPricingError{
		Kind:    UpstreamPricingErrorUpstream,
		Message: "No compatible New API or Sub2API pricing endpoint was found",
		Err:     lastErr,
	}
}

func (s *AccountTestService) fetchSub2APIGroupCatalog(ctx context.Context, baseURL, dashboardToken string, account *Account) (*UpstreamPricingSnapshot, error) {
	candidates, err := buildUpstreamPricingCandidates(baseURL)
	if err != nil {
		return nil, err
	}
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Source == "newapi" {
			roots = append(roots, strings.TrimSuffix(candidate.URL, "/api/pricing"))
		}
	}
	var lastErr error
	for _, root := range upstreamPricingDedupeStrings(roots) {
		for _, prefix := range []string{"/api/v1", "/api", ""} {
			availableEndpoint := root + prefix + "/groups/available"
			available, requestErr := s.fetchUpstreamJSONObject(ctx, availableEndpoint, dashboardToken, account)
			if requestErr != nil {
				lastErr = requestErr
				continue
			}
			items, ok := available["data"].([]any)
			if !ok {
				lastErr = errors.New("Sub2API available groups response did not contain data")
				continue
			}
			ratios := make(map[string]float64, len(items))
			names := make(map[string]string, len(items))
			for _, item := range items {
				group, ok := item.(map[string]any)
				if !ok {
					continue
				}
				id, idOK := numberFromAny(group["id"])
				ratio, ratioOK := numberFromAny(group["rate_multiplier"])
				if !idOK || !ratioOK {
					continue
				}
				key := strconv.FormatFloat(id, 'f', -1, 64)
				ratios[key] = ratio
				names[key] = upstreamPricingFirstString(group["name"], group["description"], key)
			}

			if ratesResponse, ratesErr := s.fetchUpstreamJSONObject(ctx, root+prefix+"/groups/rates", dashboardToken, account); ratesErr == nil {
				if effectiveRates, ok := ratesResponse["data"].(map[string]any); ok {
					for key, value := range effectiveRates {
						if ratio, ratioOK := numberFromAny(value); ratioOK {
							ratios[key] = ratio
						}
					}
				}
			}

			// Match the upstream model API key to its bound Sub2API group. This
			// prevents users from having to guess among all available groups.
			apiKey := strings.TrimSpace(account.GetCredential("api_key"))
			if apiKey != "" {
				if keysResponse, keysErr := s.fetchUpstreamJSONObject(ctx, root+prefix+"/keys", dashboardToken, account); keysErr == nil {
					if data, ok := keysResponse["data"].(map[string]any); ok {
						if keyItems, ok := data["items"].([]any); ok {
							for _, item := range keyItems {
								entry, ok := item.(map[string]any)
								if !ok || strings.TrimSpace(upstreamPricingFirstString(entry["key"])) != apiKey {
									continue
								}
								if groupID, groupOK := numberFromAny(entry["group_id"]); groupOK {
									selected := strconv.FormatFloat(groupID, 'f', -1, 64)
									for key := range ratios {
										if key != selected {
											delete(ratios, key)
											delete(names, key)
										}
									}
								}
								if strings.TrimSpace(upstreamPricingFirstString(entry["key"])) == apiKey {
									break
								}
							}
						}
					}
				}
			}
			if len(ratios) == 0 {
				lastErr = errors.New("Sub2API group endpoints did not contain usable groups")
				continue
			}
			return &UpstreamPricingSnapshot{Source: "sub2api", Endpoint: availableEndpoint, GroupRatios: ratios, GroupNames: names, Models: []UpstreamPricingModel{}}, nil
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no Sub2API group endpoint candidates were available")
	}
	return nil, lastErr
}

func (s *AccountTestService) fetchUpstreamJSONObject(ctx context.Context, endpoint, token string, account *Account) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "58Token/1.0")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.doUpstreamModelsRequest(req, upstreamModelsProxyURL(account), account)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, upstreamPricingBodyLimit+1))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("endpoint %s returned HTTP %d", endpoint, resp.StatusCode)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// fetchUpstreamBalance tries the authenticated profile endpoints used by the
// two common upstream families. Balance is deliberately best-effort: private
// forks often disable the profile endpoint while still exposing pricing.
func (s *AccountTestService) fetchUpstreamBalance(
	ctx context.Context,
	baseURL, dashboardToken, dashboardUserID string,
	account *Account,
) (*UpstreamBalance, error) {
	if strings.TrimSpace(dashboardToken) == "" {
		return nil, errors.New("dashboard access token is required for balance lookup")
	}
	candidates, err := buildUpstreamPricingCandidates(baseURL)
	if err != nil {
		return nil, err
	}
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Source == "newapi" {
			roots = append(roots, strings.TrimSuffix(candidate.URL, "/api/pricing"))
		}
	}
	paths := []string{
		"/api/v1/auth/me", // Sub2API
		"/api/auth/me",
		"/auth/me",
		"/api/user/self", // New API
		"/api/v1/user/self",
		"/api/user/me",
	}
	var lastErr error
	seen := make(map[string]struct{})
	for _, root := range upstreamPricingDedupeStrings(roots) {
		for _, path := range paths {
			endpoint := root + path
			if _, ok := seen[endpoint]; ok {
				continue
			}
			seen[endpoint] = struct{}{}
			raw, requestErr := s.fetchUpstreamJSONObjectWithUser(ctx, endpoint, dashboardToken, dashboardUserID, account)
			if requestErr != nil {
				lastErr = requestErr
				continue
			}
			balance, parseErr := parseUpstreamBalance(raw)
			if parseErr != nil {
				lastErr = parseErr
				continue
			}
			if balance.Unit == "quota" {
				if status, statusErr := s.fetchUpstreamJSONObjectWithUser(ctx, root+"/api/status", dashboardToken, dashboardUserID, account); statusErr == nil {
					if quotaPerUnit, ok := findUpstreamNumber(status, "quota_per_unit"); ok && quotaPerUnit > 0 {
						balance.RawAmount = balance.Amount
						balance.Amount /= quotaPerUnit
						balance.Unit = "currency"
						balance.Currency = "USD"
						if balance.UsedAmount != nil {
							used := *balance.UsedAmount / quotaPerUnit
							balance.UsedAmount = &used
						}
					}
				}
			}
			balance.Source = "newapi"
			if strings.Contains(path, "/auth/me") {
				balance.Source = "sub2api"
			}
			balance.Endpoint = endpoint
			balance.CheckedAt = time.Now().UTC()
			return balance, nil
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no upstream balance endpoint candidates were available")
	}
	return nil, lastErr
}

func (s *AccountTestService) fetchUpstreamJSONObjectWithUser(ctx context.Context, endpoint, token, userID string, account *Account) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "58Token/1.0")
	req.Header.Set("Authorization", "Bearer "+token)
	if strings.TrimSpace(userID) != "" {
		req.Header.Set("New-Api-User", strings.TrimSpace(userID))
	}
	resp, err := s.doUpstreamModelsRequest(req, upstreamModelsProxyURL(account), account)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, upstreamPricingBodyLimit+1))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("endpoint %s returned HTTP %d", endpoint, resp.StatusCode)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func parseUpstreamBalance(raw map[string]any) (*UpstreamBalance, error) {
	containers := []any{raw}
	if data, ok := raw["data"]; ok {
		containers = append([]any{data}, containers...)
	}
	for _, container := range containers {
		fields, ok := container.(map[string]any)
		if !ok {
			continue
		}
		for _, key := range []string{"balance", "available_balance", "remaining_balance", "wallet_balance", "credit", "credits", "quota", "remaining_quota"} {
			amount, amountOK := numberFromAny(fields[key])
			if !amountOK {
				continue
			}
			unit := "currency"
			if strings.Contains(key, "quota") {
				unit = "quota"
			}
			balance := &UpstreamBalance{Amount: amount, RawAmount: amount, Unit: unit}
			for _, usedKey := range []string{"used_balance", "used_quota", "used_credits"} {
				if used, usedOK := numberFromAny(fields[usedKey]); usedOK {
					balance.UsedAmount = &used
					break
				}
			}
			balance.Currency = upstreamPricingFirstString(fields["currency"], fields["currency_code"])
			return balance, nil
		}
	}
	return nil, errors.New("upstream profile response did not contain a balance or quota")
}

func findUpstreamNumber(raw map[string]any, key string) (float64, bool) {
	if value, ok := numberFromAny(raw[key]); ok {
		return value, true
	}
	if data, ok := raw["data"].(map[string]any); ok {
		return numberFromAny(data[key])
	}
	return 0, false
}

func normalizeUpstreamDashboardToken(value string) string {
	token := strings.TrimSpace(value)
	token = strings.Trim(token, "\"")
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return token
}

func (s *AccountTestService) fetchSub2APIModelPlaza(ctx context.Context, baseURL, dashboardToken string, account *Account) (*UpstreamPricingSnapshot, error) {
	candidates, err := buildUpstreamPricingCandidates(baseURL)
	if err != nil {
		return nil, err
	}
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Source == "newapi" {
			roots = append(roots, strings.TrimSuffix(candidate.URL, "/api/pricing"))
		}
	}
	var lastErr error
	for _, root := range upstreamPricingDedupeStrings(roots) {
		for _, path := range []string{"/api/v1/model-plaza", "/api/model-plaza", "/model-plaza"} {
			endpoint := root + path
			req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if reqErr != nil {
				lastErr = reqErr
				continue
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", "58Token/1.0")
			req.Header.Set("Authorization", "Bearer "+dashboardToken)
			resp, requestErr := s.doUpstreamModelsRequest(req, upstreamModelsProxyURL(account), account)
			if requestErr != nil {
				lastErr = requestErr
				continue
			}
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, upstreamPricingBodyLimit+1))
			_ = resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
				continue
			}
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				lastErr = fmt.Errorf("Sub2API model plaza returned HTTP %d", resp.StatusCode)
				continue
			}
			ratios, names, parseErr := parseSub2APIModelPlaza(body)
			if parseErr != nil {
				lastErr = parseErr
				continue
			}
			return &UpstreamPricingSnapshot{
				Source:      "sub2api",
				Endpoint:    endpoint,
				GroupRatios: ratios,
				GroupNames:  names,
				Models:      []UpstreamPricingModel{},
			}, nil
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no Sub2API model-plaza endpoint candidates were available")
	}
	return nil, lastErr
}

func parseSub2APIModelPlaza(body []byte) (map[string]float64, map[string]string, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, nil, fmt.Errorf("parse Sub2API model plaza response: %w", err)
	}
	container := raw["data"]
	if container == nil {
		container = raw
	}
	data, ok := container.(map[string]any)
	if !ok {
		return nil, nil, errors.New("Sub2API model plaza response did not contain data")
	}
	items, ok := data["groups"].([]any)
	if !ok {
		return nil, nil, errors.New("Sub2API model plaza response did not contain groups")
	}
	ratios := make(map[string]float64, len(items))
	names := make(map[string]string, len(items))
	for _, item := range items {
		group, ok := item.(map[string]any)
		if !ok {
			continue
		}
		key := upstreamPricingFirstString(group["key"], group["id"])
		if key == "" {
			if id, idOK := numberFromAny(group["id"]); idOK {
				key = strconv.FormatFloat(id, 'f', -1, 64)
			}
		}
		if key == "" {
			key = upstreamPricingFirstString(group["name"])
		}
		if key == "" {
			continue
		}
		ratio, ratioOK := numberFromAny(group["user_rate_multiplier"])
		if !ratioOK {
			ratio, ratioOK = numberFromAny(group["rate_multiplier"])
		}
		if !ratioOK {
			continue
		}
		ratios[key] = ratio
		names[key] = upstreamPricingFirstString(group["name"], group["description"], key)
	}
	if len(ratios) == 0 {
		return nil, nil, errors.New("Sub2API model plaza response did not contain group ratios")
	}
	return ratios, names, nil
}

func (s *AccountTestService) fetchEffectiveUpstreamGroupRatios(
	ctx context.Context,
	baseURL, dashboardToken, dashboardUserID string,
	account *Account,
) (map[string]float64, map[string]string, string, error) {
	candidates, err := buildUpstreamPricingCandidates(baseURL)
	if err != nil {
		return nil, nil, "", &UpstreamPricingError{Kind: UpstreamPricingErrorConfiguration, Message: "Invalid upstream dashboard URL", Err: err}
	}
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Source == "newapi" {
			roots = append(roots, strings.TrimSuffix(candidate.URL, "/api/pricing"))
		}
	}

	var lastErr error
	for _, root := range upstreamPricingDedupeStrings(roots) {
		for _, path := range []string{"/api/user/self/groups", "/api/user/groups"} {
			// Older New API deployments expose the authenticated endpoint only
			// through /api/user/self/groups. Without a user id, /api/user/groups
			// is a public base-ratio endpoint and must not be treated as effective.
			if path == "/api/user/groups" && dashboardUserID == "" {
				continue
			}
			endpoint := root + path
			req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if reqErr != nil {
				lastErr = reqErr
				continue
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", "58Token/1.0")
			req.Header.Set("Authorization", "Bearer "+dashboardToken)
			if dashboardUserID != "" {
				req.Header.Set("New-Api-User", dashboardUserID)
			}

			resp, requestErr := s.doUpstreamModelsRequest(req, upstreamModelsProxyURL(account), account)
			if requestErr != nil {
				lastErr = requestErr
				continue
			}
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, upstreamPricingBodyLimit+1))
			_ = resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
				continue
			}
			if int64(len(body)) > upstreamPricingBodyLimit {
				lastErr = fmt.Errorf("response exceeds %d bytes", upstreamPricingBodyLimit)
				continue
			}
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				lastErr = fmt.Errorf("dashboard group endpoint returned HTTP %d", resp.StatusCode)
				continue
			}
			ratios, names, parseErr := parseEffectiveUpstreamGroups(body, path == "/api/user/self/groups")
			if parseErr != nil {
				lastErr = parseErr
				continue
			}
			return ratios, names, endpoint, nil
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no dashboard group endpoint candidates were available")
	}
	if ratiosSnapshot, plazaErr := s.fetchSub2APIModelPlaza(ctx, baseURL, dashboardToken, account); plazaErr == nil {
		return ratiosSnapshot.GroupRatios, ratiosSnapshot.GroupNames, ratiosSnapshot.Endpoint, nil
	}
	if dashboardUserID == "" {
		return nil, nil, "", &UpstreamPricingError{
			Kind:    UpstreamPricingErrorConfiguration,
			Message: "This upstream requires the upstream user ID (New-Api-User) together with the dashboard access token",
			Err:     lastErr,
		}
	}
	return nil, nil, "", &UpstreamPricingError{
		Kind:    UpstreamPricingErrorUpstream,
		Message: "The upstream dashboard access token could not read effective group ratios",
		Err:     lastErr,
	}
}

func parseEffectiveUpstreamGroups(body []byte, authenticated ...bool) (map[string]float64, map[string]string, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, nil, fmt.Errorf("parse dashboard group response: %w", err)
	}
	if success, ok := raw["success"].(bool); ok && !success {
		return nil, nil, errors.New("dashboard group response reported failure")
	}
	container := raw["data"]
	if container == nil {
		container = raw["usable_group"]
	}
	groups, ok := container.(map[string]any)
	if !ok {
		return nil, nil, errors.New("dashboard group response did not contain groups")
	}
	ratios := make(map[string]float64, len(groups))
	names := make(map[string]string, len(groups))
	allowUserOverride := len(authenticated) > 0 && authenticated[0]
	for key, item := range groups {
		switch value := item.(type) {
		case map[string]any:
			// The public legacy endpoint returns both base_ratio and ratio. That
			// ratio is not user-specific, so never label it as effective pricing.
			if _, hasBaseRatio := value["base_ratio"]; hasBaseRatio && !allowUserOverride {
				continue
			}
			ratio, ratioOK := numberFromAny(value["ratio"])
			if !ratioOK {
				ratio, ratioOK = numberFromAny(value["current_ratio"])
			}
			if !ratioOK {
				continue
			}
			ratios[key] = ratio
			names[key] = upstreamPricingFirstString(value["desc"], value["description"], value["name"], key)
		default:
			if ratio, ratioOK := numberFromAny(value); ratioOK {
				ratios[key] = ratio
				names[key] = key
			}
		}
	}
	if len(ratios) == 0 {
		return nil, nil, errors.New("dashboard group response did not contain effective ratios")
	}
	return ratios, names, nil
}

func filterUpstreamPricingToGroups(snapshot *UpstreamPricingSnapshot, groups []string) {
	if snapshot == nil || len(groups) == 0 {
		return
	}
	allowed := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		allowed[group] = struct{}{}
	}
	for key := range snapshot.GroupRatios {
		if _, ok := allowed[key]; !ok {
			delete(snapshot.GroupRatios, key)
			delete(snapshot.GroupNames, key)
		}
	}
}

func (s *AccountTestService) getCachedUpstreamPricing(key string, now time.Time) *UpstreamPricingSnapshot {
	if s == nil || key == "" {
		return nil
	}
	s.upstreamPricingCacheMu.Lock()
	defer s.upstreamPricingCacheMu.Unlock()
	entry, ok := s.upstreamPricingCache[key]
	if !ok || entry.Snapshot == nil || !now.Before(entry.ExpiresAt) {
		if ok {
			delete(s.upstreamPricingCache, key)
		}
		return nil
	}
	return entry.Snapshot
}

func (s *AccountTestService) cacheUpstreamPricing(key string, snapshot *UpstreamPricingSnapshot, expiresAt time.Time) {
	if s == nil || key == "" || snapshot == nil {
		return
	}
	s.upstreamPricingCacheMu.Lock()
	defer s.upstreamPricingCacheMu.Unlock()
	if s.upstreamPricingCache == nil {
		s.upstreamPricingCache = make(map[string]upstreamPricingCacheEntry)
	}
	s.upstreamPricingCache[key] = upstreamPricingCacheEntry{Snapshot: snapshot, ExpiresAt: expiresAt}
}

func buildUpstreamPricingCandidates(base string) ([]upstreamPricingCandidate, error) {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return nil, fmt.Errorf("parse upstream base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("upstream base URL must include a scheme and host")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""

	trimmed := *parsed
	trimmed.Path = trimKnownUpstreamAPIPath(trimmed.Path)
	trimmed.RawPath = ""

	origin := *parsed
	origin.Path = ""
	origin.RawPath = ""
	origin.RawQuery = ""
	origin.Fragment = ""

	roots := upstreamPricingDedupeStrings([]string{
		strings.TrimRight(trimmed.String(), "/"),
		strings.TrimRight(origin.String(), "/"),
	})
	candidates := make([]upstreamPricingCandidate, 0, len(roots)*2)
	seen := make(map[string]struct{}, len(roots)*2)
	for _, root := range roots {
		for _, item := range []struct {
			source string
			path   string
		}{
			{source: "newapi", path: "/api/pricing"},
			{source: "sub2api", path: "/api/v1/public/model-pricing"},
		} {
			candidateURL := root + item.path
			if _, ok := seen[candidateURL]; ok {
				continue
			}
			seen[candidateURL] = struct{}{}
			candidates = append(candidates, upstreamPricingCandidate{Source: item.source, URL: candidateURL})
		}
	}
	return candidates, nil
}

func trimKnownUpstreamAPIPath(path string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(path), "/")
	lower := strings.ToLower(trimmed)
	for _, suffix := range []string{
		"/v1/chat/completions",
		"/v1/responses",
		"/v1/messages",
		"/v1/models",
		"/v1beta/models",
		"/v1beta",
		"/v1",
	} {
		if strings.HasSuffix(lower, suffix) {
			return strings.TrimRight(trimmed[:len(trimmed)-len(suffix)], "/")
		}
	}
	return trimmed
}

func parseUpstreamPricingSnapshot(body []byte) (*UpstreamPricingSnapshot, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var raw map[string]any
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse pricing response: %w", err)
	}
	if success, ok := raw["success"].(bool); ok && !success {
		return nil, errors.New("pricing response reported failure")
	}

	snapshot := &UpstreamPricingSnapshot{
		GroupRatios: parseStringNumberMap(raw["group_ratio"]),
		GroupNames:  parseStringMap(raw["usable_group"]),
		Models:      []UpstreamPricingModel{},
	}
	if version, ok := raw["pricing_version"].(string); ok {
		snapshot.PricingVersion = strings.TrimSpace(version)
	}

	items, _ := raw["data"].([]any)
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := upstreamPricingFirstString(entry["model_name"], entry["model"], entry["id"])
		if name == "" {
			continue
		}
		model := UpstreamPricingModel{
			ModelName:              name,
			Tags:                   upstreamPricingFirstString(entry["tags"]),
			QuotaType:              int(numberOrZero(entry["quota_type"])),
			ModelRatio:             numberPointer(entry["model_ratio"]),
			CompletionRatio:        numberPointer(entry["completion_ratio"]),
			CacheRatio:             numberPointer(entry["cache_ratio"]),
			CreateCacheRatio:       numberPointer(entry["create_cache_ratio"]),
			ModelPrice:             numberPointer(entry["model_price"]),
			EnabledGroups:          stringSlice(entry["enable_groups"]),
			SupportedEndpointTypes: stringSlice(entry["supported_endpoint_types"]),
		}
		snapshot.Models = append(snapshot.Models, model)
	}

	if len(snapshot.Models) == 0 && len(snapshot.GroupRatios) == 0 {
		return nil, errors.New("pricing response did not contain model or group ratios")
	}
	sort.Slice(snapshot.Models, func(i, j int) bool {
		return strings.ToLower(snapshot.Models[i].ModelName) < strings.ToLower(snapshot.Models[j].ModelName)
	})
	return snapshot, nil
}

func parseStringNumberMap(value any) map[string]float64 {
	result := map[string]float64{}
	raw, ok := value.(map[string]any)
	if !ok {
		return result
	}
	for key, item := range raw {
		if number, ok := numberFromAny(item); ok {
			result[key] = number
		}
	}
	return result
}

func parseStringMap(value any) map[string]string {
	result := map[string]string{}
	raw, ok := value.(map[string]any)
	if !ok {
		return result
	}
	for key, item := range raw {
		if text, ok := item.(string); ok {
			result[key] = text
		}
	}
	return result
}

func numberPointer(value any) *float64 {
	number, ok := numberFromAny(value)
	if !ok {
		return nil
	}
	return &number
}

func numberOrZero(value any) float64 {
	number, _ := numberFromAny(value)
	return number
}

func numberFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func upstreamPricingFirstString(values ...any) string {
	for _, value := range values {
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]string); ok {
			return upstreamPricingDedupeStrings(typed)
		}
		return []string{}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return upstreamPricingDedupeStrings(result)
}

func upstreamPricingDedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
