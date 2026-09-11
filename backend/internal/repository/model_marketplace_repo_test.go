package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestModelMarketplaceRepositoryListEnabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id",
		"model_name",
		"pricing_aliases",
		"vendor_name",
		"groups",
		"tags",
		"endpoints",
		"description",
		"official_prices",
		"sort_order",
		"enabled",
		"created_at",
		"updated_at",
	}).AddRow(
		int64(1),
		"gpt-5.5",
		[]byte(`[]`),
		"OpenAI",
		[]byte(`["Codex Lite","Codex Pro"]`),
		[]byte(`["Reasoning","Tools"]`),
		[]byte(`["openai"]`),
		"",
		[]byte(`{"input":[{"label":"<=272K","value":5}],"output":30,"cacheWrite":null,"cacheRead":0.5}`),
		10,
		true,
		now,
		now,
	)
	mock.ExpectQuery("SELECT id, model_name, pricing_aliases, vendor_name, groups, tags, endpoints, description, official_prices, sort_order, enabled, created_at, updated_at").
		WillReturnRows(rows)

	repo := NewModelMarketplaceRepository(db)
	items, err := repo.ListEnabled(context.Background())
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, items, 1)
	require.Equal(t, "gpt-5.5", items[0].ModelName)
	require.Equal(t, []string{"Codex Lite", "Codex Pro"}, items[0].Groups)
	require.JSONEq(t, `[{"label":"<=272K","value":5}]`, string(items[0].OfficialPrices.Input))
	require.JSONEq(t, `30`, string(items[0].OfficialPrices.Output))
	require.JSONEq(t, `null`, string(items[0].OfficialPrices.CacheWrite))
	require.JSONEq(t, `0.5`, string(items[0].OfficialPrices.CacheRead))
}

func TestMarketplaceListAdminWithSingleConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mock.ExpectQuery("FROM model_marketplace_group_configs").WillReturnRows(sqlmock.NewRows([]string{
		"id", "source_group_id", "source_group_name", "source_platform", "source_rate_multiplier",
		"source_sort_order", "source_models", "display_name", "display_multiplier", "enabled", "source_active",
		"is_virtual",
	}).AddRow(1, 11, "Source", "openai", 1, 0, `["Custom-Model"]`, "Display", 2, true, true, false)).RowsWillBeClosed()
	mock.ExpectQuery("FROM model_marketplace_model_configs").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{
		"id", "source_model_name", "display_name", "enabled", "sort_order", "is_virtual",
		"capability_tags_configured", "capability_tags",
	}).AddRow(2, "Custom-Model", "Display Model", false, 7, false, false, []byte(`["Reasoning"]`))).RowsWillBeClosed()

	groups, err := NewModelMarketplaceRepository(db).ListAdmin(ctx)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Len(t, groups[0].Models, 1)
	require.False(t, groups[0].Models[0].Enabled)
	require.Equal(t, 7, groups[0].Models[0].SortOrder)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMarketplaceRepositoryHasSynced(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	synced, err := NewModelMarketplaceRepository(db).HasSynced(context.Background())
	require.NoError(t, err)
	require.True(t, synced)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMarketplaceListConfiguredEnabledKeepsLegacyItemsBeforeFirstSync(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "count"}).AddRow(false, 0))
	mock.ExpectQuery("SELECT id, model_name, pricing_aliases").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "model_name", "pricing_aliases", "vendor_name", "groups", "tags",
			"endpoints", "description", "official_prices", "sort_order", "enabled",
			"created_at", "updated_at",
		}).AddRow(
			int64(1),
			"gpt-5.5",
			[]byte(`[]`),
			"OpenAI",
			[]byte(`["Codex Pro"]`),
			[]byte(`[]`),
			[]byte(`["openai"]`),
			"",
			[]byte(`{}`),
			1,
			true,
			time.Now(),
			time.Now(),
		))

	items, err := NewModelMarketplaceRepository(db).ListConfiguredEnabled(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "gpt-5.5", items[0].ModelName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMarketplaceListConfiguredEnabledHidesAllItemsAfterEmptySync(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "count"}).AddRow(true, 0))

	items, err := NewModelMarketplaceRepository(db).ListConfiguredEnabled(context.Background())
	require.NoError(t, err)
	require.Empty(t, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMarketplaceListConfiguredEnabledUsesCurrentGroupStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "count"}).AddRow(true, 1))
	mock.ExpectQuery("JOIN groups current_group").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"model_name",
			"display_name",
			"pricing_aliases",
			"vendor_name",
			"group_name",
			"source_group_id",
			"group_multiplier",
			"tags",
			"endpoints",
			"description",
			"official_prices",
			"sort_order",
		}).AddRow(
			int64(0),
			"claude-sonnet-4-6",
			"claude-sonnet-4-6",
			[]byte(`["claude-sonnet-4-6"]`),
			"Anthropic",
			"测试专属分组",
			int64(3),
			1.0,
			[]byte(`["Reasoning"]`),
			[]byte(`["anthropic"]`),
			"",
			[]byte(`{}`),
			0,
		))

	items, err := NewModelMarketplaceRepository(db).ListConfiguredEnabled(context.Background())

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "测试专属分组", items[0].Groups[0])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMarketplaceRepositorySyncOfficialPricesStoresPerMillionTokenValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE model_marketplace_items").
		WithArgs("gpt-5.6-sol", `{"input":4,"output":20,"cacheWrite":5,"cacheRead":0.4}`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = NewModelMarketplaceRepository(db).SyncOfficialPrices(context.Background(), map[string]service.MarketplaceOfficialPrice{
		"gpt-5.6-sol": {
			Input:      marketplacePricePtr(4),
			Output:     marketplacePricePtr(20),
			CacheWrite: marketplacePricePtr(5),
			CacheRead:  marketplacePricePtr(0.4),
		},
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func marketplacePricePtr(value float64) *float64 {
	return &value
}
