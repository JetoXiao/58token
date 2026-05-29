package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type modelMarketplaceRepository struct {
	db *sql.DB
}

func NewModelMarketplaceRepository(db *sql.DB) service.ModelMarketplaceRepository {
	return &modelMarketplaceRepository{db: db}
}

func (r *modelMarketplaceRepository) ListEnabled(ctx context.Context) ([]service.ModelMarketplaceItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, model_name, pricing_aliases, vendor_name, groups, tags, endpoints, description, official_prices, sort_order, enabled, created_at, updated_at
		FROM model_marketplace_items
		WHERE enabled = TRUE
		ORDER BY sort_order ASC, model_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list model marketplace items: %w", err)
	}
	defer rows.Close()

	items := make([]service.ModelMarketplaceItem, 0)
	for rows.Next() {
		var item service.ModelMarketplaceItem
		var pricingAliasesJSON, groupsJSON, tagsJSON, endpointsJSON, officialPricesJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.ModelName,
			&pricingAliasesJSON,
			&item.VendorName,
			&groupsJSON,
			&tagsJSON,
			&endpointsJSON,
			&item.Description,
			&officialPricesJSON,
			&item.SortOrder,
			&item.Enabled,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan model marketplace item: %w", err)
		}
		if err := unmarshalJSONList(pricingAliasesJSON, &item.PricingAliases); err != nil {
			return nil, fmt.Errorf("decode pricing aliases for %s: %w", item.ModelName, err)
		}
		if err := unmarshalJSONList(groupsJSON, &item.Groups); err != nil {
			return nil, fmt.Errorf("decode groups for %s: %w", item.ModelName, err)
		}
		if err := unmarshalJSONList(tagsJSON, &item.Tags); err != nil {
			return nil, fmt.Errorf("decode tags for %s: %w", item.ModelName, err)
		}
		if err := unmarshalJSONList(endpointsJSON, &item.Endpoints); err != nil {
			return nil, fmt.Errorf("decode endpoints for %s: %w", item.ModelName, err)
		}
		if err := json.Unmarshal(officialPricesJSON, &item.OfficialPrices); err != nil {
			return nil, fmt.Errorf("decode official prices for %s: %w", item.ModelName, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate model marketplace items: %w", err)
	}
	return items, nil
}

func unmarshalJSONList(data []byte, out *[]string) error {
	if len(data) == 0 {
		*out = []string{}
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	if *out == nil {
		*out = []string{}
	}
	return nil
}
