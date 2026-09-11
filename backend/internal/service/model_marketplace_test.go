package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type modelMarketplaceRepoStub struct {
	items []ModelMarketplaceItem
}

func (r modelMarketplaceRepoStub) ListEnabled(context.Context) ([]ModelMarketplaceItem, error) {
	return r.items, nil
}

type configuredModelMarketplaceRepoStub struct {
	items []ModelMarketplaceItem
}

func (r configuredModelMarketplaceRepoStub) ListConfiguredEnabled(context.Context) ([]ModelMarketplaceItem, error) {
	return r.items, nil
}

func TestModelMarketplaceServiceListEnabledFiltersRetiredModels(t *testing.T) {
	svc := NewModelMarketplaceService(modelMarketplaceRepoStub{items: []ModelMarketplaceItem{
		{ModelName: "gpt-5.2", Enabled: true},
		{ModelName: "gpt-5.3-codex", Enabled: true},
		{ModelName: "gpt-5.5", Enabled: true},
	}})

	items, err := svc.ListEnabled(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "gpt-5.5", items[0].ModelName)
}

func TestModelMarketplaceServiceListEnabledUsesConfiguredItemsWhenConfiguredRepositoryExists(t *testing.T) {
	svc := NewConfiguredModelMarketplaceService(
		modelMarketplaceRepoStub{items: []ModelMarketplaceItem{{ModelName: "legacy-model"}}},
		configuredModelMarketplaceRepoStub{items: []ModelMarketplaceItem{{ModelName: "configured-model"}}},
	)

	items, err := svc.ListEnabled(context.Background())

	require.NoError(t, err)
	require.Equal(t, []string{"configured-model"}, []string{items[0].ModelName})
}
