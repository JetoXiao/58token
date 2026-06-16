package service

import "time"

type OpsTTFTAnalysisFilter struct {
	StartTime time.Time
	EndTime   time.Time

	Platform string
	GroupID  *int64

	SlowThresholdMs int
	TopLimit        int
}

type OpsTTFTSummary struct {
	RequestCount          int64                     `json:"request_count"`
	FirstTokenSampleCount int64                     `json:"first_token_sample_count"`
	SlowRequestCount      int64                     `json:"slow_request_count"`
	SlowRate              float64                   `json:"slow_rate"`
	CacheHitCount         int64                     `json:"cache_hit_count"`
	CacheHitRate          float64                   `json:"cache_hit_rate"`
	ByRouteSource         map[string]int            `json:"by_route_source"`
	RouteSources          []*OpsTTFTRouteSourceItem `json:"route_sources"`

	TTFT     OpsPercentiles `json:"ttft"`
	Duration OpsPercentiles `json:"duration"`
	// PreUpstream is platform-side time before waiting on the upstream first token:
	// auth + routing/account selection + queue wait + connection pick.
	PreUpstream OpsPercentiles `json:"pre_upstream"`
}

type OpsTTFTRouteSourceItem struct {
	Source    string  `json:"source"`
	Count     int64   `json:"count"`
	Share     float64 `json:"share"`
	SlowCount int64   `json:"slow_count"`
	SlowRate  float64 `json:"slow_rate"`
}

type OpsTTFTReasonItem struct {
	Reason     string  `json:"reason"`
	Count      int64   `json:"count"`
	Share      float64 `json:"share"`
	AvgTTFTMs  *int    `json:"avg_ttft_ms,omitempty"`
	P95TTFTMs  *int    `json:"p95_ttft_ms,omitempty"`
	Suggestion string  `json:"suggestion"`
}

type OpsTTFTTopItem struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	SlowCount  int64   `json:"slow_count"`
	SlowRate   float64 `json:"slow_rate"`
	AvgTTFTMs  *int    `json:"avg_ttft_ms,omitempty"`
	P95TTFTMs  *int    `json:"p95_ttft_ms,omitempty"`
	P99TTFTMs  *int    `json:"p99_ttft_ms,omitempty"`
	MaxTTFTMs  *int    `json:"max_ttft_ms,omitempty"`
	RouteHint  string  `json:"route_hint,omitempty"`
	Suggestion string  `json:"suggestion,omitempty"`
}

type OpsTTFTAccountParticipationItem struct {
	AccountID           int64   `json:"account_id"`
	AccountName         string  `json:"account_name"`
	AccountType         string  `json:"account_type"`
	Platform            string  `json:"platform"`
	Status              string  `json:"status"`
	Schedulable         bool    `json:"schedulable"`
	InObservedGroup     bool    `json:"in_observed_group"`
	ObservedGroupIDs    []int64 `json:"observed_group_ids,omitempty"`
	RequestCount        int64   `json:"request_count"`
	SlowCount           int64   `json:"slow_count"`
	SlowRate            float64 `json:"slow_rate"`
	AvgTTFTMs           *int    `json:"avg_ttft_ms,omitempty"`
	P95TTFTMs           *int    `json:"p95_ttft_ms,omitempty"`
	MaxTTFTMs           *int    `json:"max_ttft_ms,omitempty"`
	ParticipationReason string  `json:"participation_reason"`
}

type OpsTTFTSlowRequest struct {
	CreatedAt               time.Time      `json:"created_at"`
	RequestID               string         `json:"request_id"`
	Platform                string         `json:"platform"`
	Model                   string         `json:"model"`
	UserID                  *int64         `json:"user_id,omitempty"`
	APIKeyID                *int64         `json:"api_key_id,omitempty"`
	AccountID               *int64         `json:"account_id,omitempty"`
	GroupID                 *int64         `json:"group_id,omitempty"`
	UserEmail               string         `json:"user_email,omitempty"`
	APIKeyName              string         `json:"api_key_name,omitempty"`
	AccountName             string         `json:"account_name,omitempty"`
	GroupName               string         `json:"group_name,omitempty"`
	RequestType             string         `json:"request_type,omitempty"`
	RouteSource             string         `json:"route_source"`
	CacheStatus             string         `json:"cache_status,omitempty"`
	SlowReason              string         `json:"slow_reason"`
	SlowDetail              string         `json:"slow_detail,omitempty"`
	FirstTokenMs            *int           `json:"first_token_ms,omitempty"`
	DurationMs              *int           `json:"duration_ms,omitempty"`
	AuthLatencyMs           *int           `json:"auth_latency_ms,omitempty"`
	RoutingMs               *int           `json:"routing_latency_ms,omitempty"`
	UpstreamMs              *int           `json:"upstream_latency_ms,omitempty"`
	ResponseMs              *int           `json:"response_latency_ms,omitempty"`
	QueueWaitMs             *int           `json:"queue_wait_ms,omitempty"`
	ConnPickMs              *int           `json:"conn_pick_ms,omitempty"`
	PreUpstreamMs           *int           `json:"pre_upstream_ms,omitempty"`
	SchedulerCandidateCount *int           `json:"scheduler_candidate_count,omitempty"`
	SchedulerTopK           *int           `json:"scheduler_top_k,omitempty"`
	SchedulerLatencyMs      *int           `json:"scheduler_latency_ms,omitempty"`
	SchedulerLayer          string         `json:"scheduler_layer,omitempty"`
	SchedulerReason         string         `json:"scheduler_reason,omitempty"`
	SchedulerLoadSkew       *float64       `json:"scheduler_load_skew,omitempty"`
	RequestParams           map[string]any `json:"request_params,omitempty"`
}

type OpsTTFTRecommendation struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Action   string `json:"action"`
}

type OpsTTFTAnalysisResponse struct {
	StartTime            time.Time                          `json:"start_time"`
	EndTime              time.Time                          `json:"end_time"`
	Platform             string                             `json:"platform,omitempty"`
	GroupID              *int64                             `json:"group_id,omitempty"`
	SlowThresholdMs      int                                `json:"slow_threshold_ms"`
	Summary              OpsTTFTSummary                     `json:"summary"`
	Reasons              []*OpsTTFTReasonItem               `json:"reasons"`
	TopModels            []*OpsTTFTTopItem                  `json:"top_models"`
	TopAccounts          []*OpsTTFTTopItem                  `json:"top_accounts"`
	TopGroups            []*OpsTTFTTopItem                  `json:"top_groups"`
	TopAPIKeys           []*OpsTTFTTopItem                  `json:"top_api_keys"`
	AccountParticipation []*OpsTTFTAccountParticipationItem `json:"account_participation"`
	SlowRequests         []*OpsTTFTSlowRequest              `json:"slow_requests"`
	Recommendations      []*OpsTTFTRecommendation           `json:"recommendations"`
}
