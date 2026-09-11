package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type marketplaceGroupSourceStub struct {
	groups []Group
}

func (s marketplaceGroupSourceStub) ListActive(context.Context) ([]Group, error) {
	return s.groups, nil
}

type marketplaceModelSourceStub struct {
	modelsByGroup map[int64][]string
}

func (s marketplaceModelSourceStub) GetAvailableModels(_ context.Context, groupID *int64, _ string) []string {
	if groupID == nil {
		return nil
	}
	return append([]string(nil), s.modelsByGroup[*groupID]...)
}

type marketplaceModelSourceWithInvalidationStub struct {
	marketplaceModelSourceStub
	invalidated bool
}

func (s *marketplaceModelSourceWithInvalidationStub) InvalidateAvailableModelsCache(*int64, string) {
	s.invalidated = true
}

type marketplaceMaintenanceStoreStub struct {
	sources               []MarketplaceGroupSource
	preserveDisplayConfig bool
	groups                []MarketplaceGroupConfig
}

func (s *marketplaceMaintenanceStoreStub) ListEnabled(context.Context) ([]ModelMarketplaceItem, error) {
	return nil, nil
}

func (s *marketplaceMaintenanceStoreStub) ListAdmin(context.Context) ([]MarketplaceGroupConfig, error) {
	return s.groups, nil
}

func (s *marketplaceMaintenanceStoreStub) HasSynced(context.Context) (bool, error) {
	return true, nil
}

func (s *marketplaceMaintenanceStoreStub) Sync(
	_ context.Context,
	sources []MarketplaceGroupSource,
	preserveDisplayConfig bool,
) error {
	s.sources = sources
	s.preserveDisplayConfig = preserveDisplayConfig
	return nil
}

func (s *marketplaceMaintenanceStoreStub) SaveGroups(context.Context, []MarketplaceGroupConfig) error {
	return nil
}

func (s *marketplaceMaintenanceStoreStub) ListConfiguredEnabled(context.Context) ([]ModelMarketplaceItem, error) {
	return nil, nil
}

func (s *marketplaceMaintenanceStoreStub) UpdateGroup(context.Context, int64, UpdateMarketplaceGroupInput) (*MarketplaceGroupConfig, error) {
	return nil, nil
}

func TestModelMarketplaceMaintenanceServiceSyncBuildsOneSourcePerActiveGroup(t *testing.T) {
	store := &marketplaceMaintenanceStoreStub{
		groups: []MarketplaceGroupConfig{
			{SourceGroupID: 11, SourceGroupName: "Claude Plus"},
		},
	}
	svc := NewModelMarketplaceMaintenanceService(
		store,
		marketplaceGroupSourceStub{groups: []Group{
			{ID: 11, Name: "Claude Plus", Platform: PlatformAnthropic, RateMultiplier: 1.25, SortOrder: 3},
			{ID: 22, Name: "Codex Pro", Platform: PlatformOpenAI, RateMultiplier: 1.5, SortOrder: 5},
		}},
		marketplaceModelSourceStub{
			modelsByGroup: map[int64][]string{
				11: {"claude-sonnet-4-6", "claude-sonnet-4-6", " claude-opus-4-7 "},
				22: {"gpt-5.5"},
			},
		},
	)

	groups, err := svc.Sync(context.Background(), true)

	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.True(t, store.preserveDisplayConfig)
	require.Len(t, store.sources, 2)
	require.Equal(t, MarketplaceGroupSource{
		SourceGroupID:   11,
		SourceGroupName: "Claude Plus",
		Platform:        PlatformAnthropic,
		RateMultiplier:  1.25,
		SortOrder:       3,
		Models:          []string{"claude-opus-4-7", "claude-sonnet-4-6"},
	}, store.sources[0])
	require.Equal(t, MarketplaceGroupSource{
		SourceGroupID:   22,
		SourceGroupName: "Codex Pro",
		Platform:        PlatformOpenAI,
		RateMultiplier:  1.5,
		SortOrder:       5,
		Models:          []string{"gpt-5.5"},
	}, store.sources[1])
}

func TestModelMarketplaceMaintenanceServiceSyncDoesNotFallbackToPlatformDefaults(t *testing.T) {
	store := &marketplaceMaintenanceStoreStub{}
	svc := NewModelMarketplaceMaintenanceService(
		store,
		marketplaceGroupSourceStub{groups: []Group{
			{ID: 11, Name: "Codex Pro", Platform: PlatformOpenAI, RateMultiplier: 1.5},
		}},
		marketplaceModelSourceStub{modelsByGroup: map[int64][]string{}},
	)

	_, err := svc.Sync(context.Background(), false)

	require.NoError(t, err)
	require.Len(t, store.sources, 1)
	require.Empty(t, store.sources[0].Models)
}

func TestModelMarketplaceMaintenanceServiceSyncInvalidatesModelSourceCache(t *testing.T) {
	store := &marketplaceMaintenanceStoreStub{}
	modelSource := &marketplaceModelSourceWithInvalidationStub{
		marketplaceModelSourceStub: marketplaceModelSourceStub{
			modelsByGroup: map[int64][]string{11: {"gpt-5.5"}},
		},
	}
	svc := NewModelMarketplaceMaintenanceService(
		store,
		marketplaceGroupSourceStub{groups: []Group{
			{ID: 11, Name: "Codex Pro", Platform: PlatformOpenAI, RateMultiplier: 1},
		}},
		modelSource,
	)

	_, err := svc.Sync(context.Background(), false)

	require.NoError(t, err)
	require.True(t, modelSource.invalidated)
}
