package service

import (
	"context"
	"strings"
	"time"
)

func (s *OpsService) GetTTFTAnalysis(ctx context.Context, filter *OpsTTFTAnalysisFilter) (*OpsTTFTAnalysisResponse, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if s.opsRepo == nil {
		return emptyOpsTTFTAnalysis(filter), nil
	}
	normalized := normalizeOpsTTFTAnalysisFilter(filter)
	out, err := s.opsRepo.GetTTFTAnalysis(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = emptyOpsTTFTAnalysis(normalized)
	}
	if out.Reasons == nil {
		out.Reasons = []*OpsTTFTReasonItem{}
	}
	if out.TopModels == nil {
		out.TopModels = []*OpsTTFTTopItem{}
	}
	if out.TopAccounts == nil {
		out.TopAccounts = []*OpsTTFTTopItem{}
	}
	if out.TopGroups == nil {
		out.TopGroups = []*OpsTTFTTopItem{}
	}
	if out.TopAPIKeys == nil {
		out.TopAPIKeys = []*OpsTTFTTopItem{}
	}
	if out.AccountParticipation == nil {
		out.AccountParticipation = []*OpsTTFTAccountParticipationItem{}
	}
	if out.SlowRequests == nil {
		out.SlowRequests = []*OpsTTFTSlowRequest{}
	}
	if out.Recommendations == nil {
		out.Recommendations = []*OpsTTFTRecommendation{}
	}
	return out, nil
}

func normalizeOpsTTFTAnalysisFilter(filter *OpsTTFTAnalysisFilter) *OpsTTFTAnalysisFilter {
	now := time.Now()
	out := &OpsTTFTAnalysisFilter{
		StartTime:       now.Add(-1 * time.Hour),
		EndTime:         now,
		SlowThresholdMs: 1000,
		TopLimit:        10,
	}
	if filter != nil {
		*out = *filter
	}
	if out.EndTime.IsZero() {
		out.EndTime = now
	}
	if out.StartTime.IsZero() {
		out.StartTime = out.EndTime.Add(-1 * time.Hour)
	}
	if out.StartTime.After(out.EndTime) {
		out.StartTime, out.EndTime = out.EndTime, out.StartTime
	}
	if out.SlowThresholdMs <= 0 {
		out.SlowThresholdMs = 1000
	}
	if out.TopLimit <= 0 {
		out.TopLimit = 10
	}
	if out.TopLimit > 50 {
		out.TopLimit = 50
	}
	out.Platform = strings.TrimSpace(strings.ToLower(out.Platform))
	return out
}

func emptyOpsTTFTAnalysis(filter *OpsTTFTAnalysisFilter) *OpsTTFTAnalysisResponse {
	normalized := normalizeOpsTTFTAnalysisFilter(filter)
	return &OpsTTFTAnalysisResponse{
		StartTime:       normalized.StartTime,
		EndTime:         normalized.EndTime,
		Platform:        normalized.Platform,
		GroupID:         normalized.GroupID,
		SlowThresholdMs: normalized.SlowThresholdMs,
		Summary: OpsTTFTSummary{
			ByRouteSource: map[string]int{},
		},
		Reasons:              []*OpsTTFTReasonItem{},
		TopModels:            []*OpsTTFTTopItem{},
		TopAccounts:          []*OpsTTFTTopItem{},
		TopGroups:            []*OpsTTFTTopItem{},
		TopAPIKeys:           []*OpsTTFTTopItem{},
		AccountParticipation: []*OpsTTFTAccountParticipationItem{},
		SlowRequests:         []*OpsTTFTSlowRequest{},
		Recommendations:      []*OpsTTFTRecommendation{},
	}
}
