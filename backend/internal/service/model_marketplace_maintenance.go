package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// MarketplaceGroupSource is the source snapshot of one active gateway group.
// The source fields are refreshed by Sync; display fields are preserved or reset
// according to the sync mode selected by the administrator.
type MarketplaceGroupSource struct {
	SourceGroupID   int64    `json:"source_group_id"`
	SourceGroupName string   `json:"source_group_name"`
	Platform        string   `json:"platform"`
	RateMultiplier  float64  `json:"rate_multiplier"`
	SortOrder       int      `json:"sort_order"`
	Models          []string `json:"models"`
}

type MarketplaceModelConfig struct {
	ID                       int64    `json:"id,omitempty"`
	SourceModelName          string   `json:"source_model_name"`
	DisplayName              string   `json:"display_name,omitempty"`
	CapabilityTags           []string `json:"capability_tags,omitempty"`
	CapabilityTagsConfigured bool     `json:"capability_tags_configured"`
	Enabled                  bool     `json:"enabled"`
	SortOrder                int      `json:"sort_order"`
	IsVirtual                bool     `json:"is_virtual"`
}

// MarketplaceGroupConfig combines the latest source snapshot with admin-owned
// presentation settings.
type MarketplaceGroupConfig struct {
	ID                int64                    `json:"id"`
	SourceGroupID     int64                    `json:"source_group_id"`
	SourceGroupName   string                   `json:"source_group_name"`
	Platform          string                   `json:"platform"`
	SourceRate        float64                  `json:"source_rate_multiplier"`
	SourceSortOrder   int                      `json:"source_sort_order"`
	SourceModels      []string                 `json:"source_models"`
	DisplayName       string                   `json:"display_name"`
	DisplayMultiplier float64                  `json:"display_multiplier"`
	Enabled           bool                     `json:"enabled"`
	SourceActive      bool                     `json:"source_active"`
	IsVirtual         bool                     `json:"is_virtual"`
	Models            []MarketplaceModelConfig `json:"models"`
}

type SyncMarketplaceGroupsInput struct {
	PreserveDisplayConfig bool `json:"preserve_display_config"`
}

type SaveMarketplaceGroupsInput struct {
	Groups []MarketplaceGroupConfig `json:"groups"`
}

type UpdateMarketplaceGroupInput struct {
	DisplayName       string                   `json:"display_name"`
	DisplayMultiplier float64                  `json:"display_multiplier"`
	Enabled           bool                     `json:"enabled"`
	Models            []MarketplaceModelConfig `json:"models"`
}

type MarketplaceGroupSourceReader interface {
	ListActive(ctx context.Context) ([]Group, error)
}

type MarketplaceModelSourceReader interface {
	GetAvailableModels(ctx context.Context, groupID *int64, platform string) []string
}

type MarketplaceModelSourceCacheInvalidator interface {
	InvalidateAvailableModelsCache(groupID *int64, platform string)
}

type ModelMarketplaceMaintenanceStore interface {
	ModelMarketplaceRepository
	ListAdmin(ctx context.Context) ([]MarketplaceGroupConfig, error)
	HasSynced(ctx context.Context) (bool, error)
	Sync(ctx context.Context, sources []MarketplaceGroupSource, preserveDisplayConfig bool) error
	SaveGroups(ctx context.Context, groups []MarketplaceGroupConfig) error
	ListConfiguredEnabled(ctx context.Context) ([]ModelMarketplaceItem, error)
	UpdateGroup(ctx context.Context, sourceGroupID int64, input UpdateMarketplaceGroupInput) (*MarketplaceGroupConfig, error)
}

type ModelMarketplaceConfiguredRepository interface {
	ListConfiguredEnabled(ctx context.Context) ([]ModelMarketplaceItem, error)
}

type ModelMarketplaceMaintenanceService struct {
	store       ModelMarketplaceMaintenanceStore
	groupSource MarketplaceGroupSourceReader
	modelSource MarketplaceModelSourceReader
}

func NewModelMarketplaceMaintenanceService(
	store ModelMarketplaceMaintenanceStore,
	groupSource MarketplaceGroupSourceReader,
	modelSource MarketplaceModelSourceReader,
) *ModelMarketplaceMaintenanceService {
	return &ModelMarketplaceMaintenanceService{
		store:       store,
		groupSource: groupSource,
		modelSource: modelSource,
	}
}

func ProvideMarketplaceGroupSource(repo GroupRepository) MarketplaceGroupSourceReader {
	return repo
}

func ProvideMarketplaceModelSource(gateway *GatewayService) MarketplaceModelSourceReader {
	return gateway
}

func (s *ModelMarketplaceMaintenanceService) Sync(
	ctx context.Context,
	preserveDisplayConfig bool,
) ([]MarketplaceGroupConfig, error) {
	if s == nil || s.store == nil || s.groupSource == nil || s.modelSource == nil {
		return nil, fmt.Errorf("model marketplace maintenance service is not configured")
	}

	groups, err := s.groupSource.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active marketplace groups: %w", err)
	}

	sources := make([]MarketplaceGroupSource, 0, len(groups))
	for _, group := range groups {
		groupID := group.ID
		if invalidator, ok := s.modelSource.(MarketplaceModelSourceCacheInvalidator); ok {
			invalidator.InvalidateAvailableModelsCache(&groupID, group.Platform)
		}
		models := normalizeMarketplaceModels(
			s.modelSource.GetAvailableModels(ctx, &groupID, group.Platform),
		)
		sources = append(sources, MarketplaceGroupSource{
			SourceGroupID:   group.ID,
			SourceGroupName: strings.TrimSpace(group.Name),
			Platform:        strings.TrimSpace(group.Platform),
			RateMultiplier:  group.RateMultiplier,
			SortOrder:       group.SortOrder,
			Models:          models,
		})
	}

	if err := s.store.Sync(ctx, sources, preserveDisplayConfig); err != nil {
		return nil, fmt.Errorf("sync model marketplace groups: %w", err)
	}
	configs, err := s.store.ListAdmin(ctx)
	if err != nil {
		return nil, fmt.Errorf("list synced model marketplace groups: %w", err)
	}
	return configs, nil
}

func (s *ModelMarketplaceMaintenanceService) ListAdmin(ctx context.Context) ([]MarketplaceGroupConfig, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("model marketplace maintenance service is not configured")
	}
	return s.store.ListAdmin(ctx)
}

func (s *ModelMarketplaceMaintenanceService) HasSynced(ctx context.Context) (bool, error) {
	if s == nil || s.store == nil {
		return false, fmt.Errorf("model marketplace maintenance service is not configured")
	}
	return s.store.HasSynced(ctx)
}

func (s *ModelMarketplaceMaintenanceService) SaveGroups(
	ctx context.Context,
	input SaveMarketplaceGroupsInput,
) ([]MarketplaceGroupConfig, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("model marketplace maintenance service is not configured")
	}
	for _, group := range input.Groups {
		if group.DisplayMultiplier <= 0 {
			return nil, fmt.Errorf("display_multiplier must be greater than zero")
		}
	}
	if err := s.store.SaveGroups(ctx, input.Groups); err != nil {
		return nil, fmt.Errorf("save model marketplace groups: %w", err)
	}
	groups, err := s.store.ListAdmin(ctx)
	if err != nil {
		return nil, fmt.Errorf("list saved model marketplace groups: %w", err)
	}
	return groups, nil
}

func (s *ModelMarketplaceMaintenanceService) UpdateGroup(
	ctx context.Context,
	sourceGroupID int64,
	input UpdateMarketplaceGroupInput,
) (*MarketplaceGroupConfig, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("model marketplace maintenance service is not configured")
	}
	if sourceGroupID <= 0 {
		return nil, fmt.Errorf("source_group_id must be positive")
	}
	if input.DisplayMultiplier <= 0 {
		return nil, fmt.Errorf("display_multiplier must be greater than zero")
	}
	return s.store.UpdateGroup(ctx, sourceGroupID, input)
}

func normalizeMarketplaceModels(models []string) []string {
	seen := make(map[string]struct{}, len(models))
	normalized := make([]string, 0, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		normalized = append(normalized, model)
	}
	sort.Strings(normalized)
	return normalized
}
