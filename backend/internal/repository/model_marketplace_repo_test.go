package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
