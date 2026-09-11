package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type modelMarketplaceRepository struct {
	db *sql.DB
}

func NewModelMarketplaceRepository(db *sql.DB) *modelMarketplaceRepository {
	return &modelMarketplaceRepository{db: db}
}

func ProvideModelMarketplaceRepository(
	repo *modelMarketplaceRepository,
) service.ModelMarketplaceRepository {
	return repo
}

func ProvideModelMarketplaceConfiguredRepository(
	repo *modelMarketplaceRepository,
) service.ModelMarketplaceConfiguredRepository {
	return repo
}

func ProvideModelMarketplaceMaintenanceStore(
	repo *modelMarketplaceRepository,
) service.ModelMarketplaceMaintenanceStore {
	return repo
}

func ProvideMarketplaceOfficialPriceStore(
	repo *modelMarketplaceRepository,
) service.MarketplaceOfficialPriceStore {
	return repo
}

func (r *modelMarketplaceRepository) SyncOfficialPrices(
	ctx context.Context,
	prices map[string]service.MarketplaceOfficialPrice,
) error {
	if len(prices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin model marketplace official price sync: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for modelName, price := range prices {
		payload := marketplaceOfficialPricePayload{
			Input:      price.Input,
			Output:     price.Output,
			CacheWrite: price.CacheWrite,
			CacheRead:  price.CacheRead,
		}
		if payload.Input == nil && payload.Output == nil &&
			payload.CacheWrite == nil && payload.CacheRead == nil {
			continue
		}
		data, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return fmt.Errorf("marshal official price for %s: %w", modelName, marshalErr)
		}

		if _, err = tx.ExecContext(ctx, `
			UPDATE model_marketplace_items
			SET official_prices = $2::jsonb,
				updated_at = NOW()
			WHERE (
				LOWER(model_name) = LOWER($1)
				OR EXISTS (
					 SELECT 1
					 FROM jsonb_array_elements_text(pricing_aliases) AS pricing_alias(alias_name)
					 WHERE LOWER(alias_name) = LOWER($1)
				)
			)
			AND official_prices IS DISTINCT FROM $2::jsonb
		`, modelName, string(data)); err != nil {
			return fmt.Errorf("sync official price for %s: %w", modelName, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit model marketplace official price sync: %w", err)
	}
	return nil
}

type marketplaceOfficialPricePayload struct {
	Input      *float64 `json:"input"`
	Output     *float64 `json:"output"`
	CacheWrite *float64 `json:"cacheWrite"`
	CacheRead  *float64 `json:"cacheRead"`
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

func (r *modelMarketplaceRepository) ListAdmin(ctx context.Context) ([]service.MarketplaceGroupConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			source_group_id,
			source_group_name,
			source_platform,
			source_rate_multiplier,
			source_sort_order,
			source_models,
			display_name,
			display_multiplier,
			enabled,
			source_active,
			is_virtual
		FROM model_marketplace_group_configs
		WHERE is_virtual = FALSE
		ORDER BY source_active DESC, is_virtual ASC, source_sort_order ASC,
			source_group_id ASC NULLS LAST, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list model marketplace group configs: %w", err)
	}
	defer rows.Close()

	groups := make([]service.MarketplaceGroupConfig, 0)
	for rows.Next() {
		var group service.MarketplaceGroupConfig
		var sourceModelsJSON []byte
		var sourceGroupID sql.NullInt64
		if err := rows.Scan(
			&group.ID,
			&sourceGroupID,
			&group.SourceGroupName,
			&group.Platform,
			&group.SourceRate,
			&group.SourceSortOrder,
			&sourceModelsJSON,
			&group.DisplayName,
			&group.DisplayMultiplier,
			&group.Enabled,
			&group.SourceActive,
			&group.IsVirtual,
		); err != nil {
			return nil, fmt.Errorf("scan model marketplace group config: %w", err)
		}
		if sourceGroupID.Valid {
			group.SourceGroupID = sourceGroupID.Int64
		}
		if err := unmarshalJSONList(sourceModelsJSON, &group.SourceModels); err != nil {
			return nil, fmt.Errorf("decode source models for group %d: %w", group.SourceGroupID, err)
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate model marketplace group configs: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close model marketplace group configs: %w", err)
	}
	for index := range groups {
		groups[index].Models, err = r.listMarketplaceModelConfigs(ctx, groups[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return groups, nil
}

func (r *modelMarketplaceRepository) HasSynced(ctx context.Context) (bool, error) {
	var synced bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM model_marketplace_sync_state
			WHERE id = TRUE
		)
	`).Scan(&synced); err != nil {
		return false, fmt.Errorf("check model marketplace sync state: %w", err)
	}
	return synced, nil
}

func (r *modelMarketplaceRepository) listMarketplaceModelConfigs(
	ctx context.Context,
	groupConfigID int64,
) ([]service.MarketplaceModelConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			mc.id,
			mc.source_model_name,
			mc.display_name,
			mc.enabled,
			mc.sort_order,
			mc.is_virtual,
			(mc.capability_tags IS NOT NULL) AS capability_tags_configured,
			COALESCE(
				mc.capability_tags,
				COALESCE((
					SELECT mi.tags
					FROM model_marketplace_items mi
					WHERE LOWER(mi.model_name) = LOWER(mc.source_model_name)
					   OR EXISTS (
							SELECT 1
							FROM jsonb_array_elements_text(mi.pricing_aliases) AS pricing_alias(alias_name)
							WHERE LOWER(pricing_alias.alias_name) = LOWER(mc.source_model_name)
					   )
					ORDER BY CASE
						WHEN LOWER(mi.model_name) = LOWER(mc.source_model_name) THEN 0
						ELSE 1
					END, mi.id
					LIMIT 1
				), '[]'::jsonb)
			) AS capability_tags
		FROM model_marketplace_model_configs mc
		WHERE mc.group_config_id = $1
		  AND mc.is_virtual = FALSE
		  AND EXISTS (
				SELECT 1
				FROM model_marketplace_group_configs g
				WHERE g.id = $1
				  AND EXISTS (
						SELECT 1
						FROM jsonb_array_elements_text(g.source_models) AS source_model(model_name)
						WHERE source_model.model_name = mc.source_model_name
				  )
		  )
		ORDER BY mc.sort_order ASC, mc.source_model_name ASC
	`, groupConfigID)
	if err != nil {
		return nil, fmt.Errorf("list model marketplace model configs for group %d: %w", groupConfigID, err)
	}
	defer rows.Close()

	models := make([]service.MarketplaceModelConfig, 0)
	for rows.Next() {
		var model service.MarketplaceModelConfig
		var capabilityTagsJSON []byte
		if err := rows.Scan(
			&model.ID,
			&model.SourceModelName,
			&model.DisplayName,
			&model.Enabled,
			&model.SortOrder,
			&model.IsVirtual,
			&model.CapabilityTagsConfigured,
			&capabilityTagsJSON,
		); err != nil {
			return nil, fmt.Errorf("scan model marketplace model config: %w", err)
		}
		if err := unmarshalJSONList(capabilityTagsJSON, &model.CapabilityTags); err != nil {
			return nil, fmt.Errorf("decode capability tags for %s: %w", model.SourceModelName, err)
		}
		models = append(models, model)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate model marketplace model configs: %w", err)
	}
	return models, nil
}

func (r *modelMarketplaceRepository) Sync(
	ctx context.Context,
	sources []service.MarketplaceGroupSource,
	preserveDisplayConfig bool,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin model marketplace sync: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		UPDATE model_marketplace_group_configs
		SET source_active = FALSE, updated_at = NOW()
		WHERE is_virtual = FALSE
	`); err != nil {
		return fmt.Errorf("mark inactive model marketplace groups: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		DELETE FROM model_marketplace_group_configs
		WHERE is_virtual = TRUE
	`); err != nil {
		return fmt.Errorf("remove legacy virtual model marketplace groups: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		DELETE FROM model_marketplace_model_configs
		WHERE is_virtual = TRUE
	`); err != nil {
		return fmt.Errorf("remove legacy virtual model marketplace models: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO model_marketplace_sync_state (id, synced_at)
		VALUES (TRUE, NOW())
		ON CONFLICT (id) DO UPDATE SET synced_at = EXCLUDED.synced_at
	`); err != nil {
		return fmt.Errorf("record model marketplace sync state: %w", err)
	}

	activeIDs := make([]int64, 0, len(sources))
	for _, source := range sources {
		activeIDs = append(activeIDs, source.SourceGroupID)
		var groupConfigID int64
		err = tx.QueryRowContext(ctx, `
			SELECT id
			FROM model_marketplace_group_configs
			WHERE source_group_id = $1 AND is_virtual = FALSE
		`, source.SourceGroupID).Scan(&groupConfigID)
		if err == sql.ErrNoRows {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO model_marketplace_group_configs (
					source_group_id,
					source_group_name,
					source_platform,
					source_rate_multiplier,
					source_sort_order,
					source_models,
					display_name,
					display_multiplier,
					enabled,
					source_active,
					is_virtual
				) VALUES (
					$1::bigint,
					$2::varchar,
					$3::varchar,
					$4::numeric,
					$5::integer,
					$6::jsonb,
					CASE WHEN $7::boolean THEN ''::varchar ELSE $2::varchar END,
					GREATEST($4::numeric, 0.0001::numeric),
					NOT $7::boolean,
					TRUE,
					FALSE
				)
				RETURNING id
			`,
				source.SourceGroupID,
				source.SourceGroupName,
				source.Platform,
				source.RateMultiplier,
				source.SortOrder,
				mustMarshalJSON(source.Models),
				preserveDisplayConfig,
			).Scan(&groupConfigID)
		} else if err == nil {
			if preserveDisplayConfig {
				_, err = tx.ExecContext(ctx, `
					UPDATE model_marketplace_group_configs
					SET source_group_name = $2,
						source_platform = $3,
						source_rate_multiplier = $4,
						source_sort_order = $5,
						source_models = $6::jsonb,
						source_active = TRUE,
						is_virtual = FALSE,
						updated_at = NOW()
					WHERE id = $1
				`,
					groupConfigID,
					source.SourceGroupName,
					source.Platform,
					source.RateMultiplier,
					source.SortOrder,
					mustMarshalJSON(source.Models),
				)
			} else {
				_, err = tx.ExecContext(ctx, `
					UPDATE model_marketplace_group_configs
					SET source_group_name = $2,
						source_platform = $3,
						source_rate_multiplier = $4,
						source_sort_order = $5,
						source_models = $6::jsonb,
						display_name = $2,
						display_multiplier = GREATEST($4::numeric, 0.0001::numeric),
						enabled = TRUE,
						source_active = TRUE,
						is_virtual = FALSE,
						updated_at = NOW()
					WHERE id = $1
				`,
					groupConfigID,
					source.SourceGroupName,
					source.Platform,
					source.RateMultiplier,
					source.SortOrder,
					mustMarshalJSON(source.Models),
				)
			}
		}
		if err != nil {
			return fmt.Errorf("upsert model marketplace group %d: %w", source.SourceGroupID, err)
		}

		for index, modelName := range source.Models {
			if preserveDisplayConfig {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO model_marketplace_model_configs (
						group_config_id, source_model_name, display_name, enabled, sort_order, is_virtual
					) VALUES ($1, $2, '', FALSE, $3, FALSE)
					ON CONFLICT (group_config_id, source_model_name) DO UPDATE SET
						sort_order = EXCLUDED.sort_order,
						is_virtual = FALSE,
						updated_at = NOW()
				`, groupConfigID, modelName, index)
			} else {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO model_marketplace_model_configs (
						group_config_id, source_model_name, display_name, enabled, sort_order, is_virtual, capability_tags
					) VALUES ($1, $2, '', TRUE, $3, FALSE, NULL)
					ON CONFLICT (group_config_id, source_model_name) DO UPDATE SET
						display_name = '',
						enabled = TRUE,
						sort_order = EXCLUDED.sort_order,
						is_virtual = FALSE,
						capability_tags = NULL,
						updated_at = NOW()
				`, groupConfigID, modelName, index)
			}
			if err != nil {
				return fmt.Errorf("upsert model marketplace model %s for group %d: %w", modelName, source.SourceGroupID, err)
			}
		}
		if !preserveDisplayConfig {
			if len(source.Models) == 0 {
				_, err = tx.ExecContext(ctx, `
					DELETE FROM model_marketplace_model_configs
					WHERE group_config_id = $1
					  AND is_virtual = FALSE
				`, groupConfigID)
			} else {
				_, err = tx.ExecContext(ctx, `
					DELETE FROM model_marketplace_model_configs
					WHERE group_config_id = $1
					  AND is_virtual = FALSE
					  AND NOT (source_model_name = ANY($2::text[]))
				`, groupConfigID, pq.Array(source.Models))
			}
			if err != nil {
				return fmt.Errorf("remove stale model marketplace models for group %d: %w", source.SourceGroupID, err)
			}
		}
	}

	if len(activeIDs) == 0 {
		if _, err = tx.ExecContext(ctx, `
			UPDATE model_marketplace_group_configs
			SET source_active = FALSE,
				enabled = CASE WHEN $1::boolean THEN enabled ELSE FALSE END,
				updated_at = NOW()
			WHERE is_virtual = FALSE
		`, preserveDisplayConfig); err != nil {
			return fmt.Errorf("clear model marketplace groups after empty sync: %w", err)
		}
	} else if _, err = tx.ExecContext(ctx, `
		UPDATE model_marketplace_group_configs
		SET source_active = FALSE,
			enabled = CASE WHEN $1::boolean THEN enabled ELSE FALSE END,
			updated_at = NOW()
		WHERE is_virtual = FALSE
		  AND (source_group_id IS NULL OR NOT (source_group_id = ANY($2::bigint[])))
	`, preserveDisplayConfig, pqInt64Array(activeIDs)); err != nil {
		return fmt.Errorf("finalize inactive model marketplace groups: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit model marketplace sync: %w", err)
	}
	return nil
}

func (r *modelMarketplaceRepository) SaveGroups(
	ctx context.Context,
	groups []service.MarketplaceGroupConfig,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin model marketplace group save: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	savedGroupIDs := make([]int64, 0, len(groups))
	for _, group := range groups {
		isVirtual := group.IsVirtual || group.SourceGroupID <= 0
		var groupConfigID int64
		sourceGroupName := strings.TrimSpace(group.SourceGroupName)
		platform := strings.TrimSpace(group.Platform)
		displayName := strings.TrimSpace(group.DisplayName)
		displayMultiplier := group.DisplayMultiplier
		if displayMultiplier <= 0 {
			displayMultiplier = 1
		}

		if isVirtual {
			if sourceGroupName == "" {
				sourceGroupName = displayName
			}
			if group.SourceRate <= 0 {
				group.SourceRate = displayMultiplier
			}
			sourceModels := marketplaceModelNames(group.Models)
			if group.ID > 0 {
				result, execErr := tx.ExecContext(ctx, `
					UPDATE model_marketplace_group_configs
					SET source_group_id = NULL,
						source_group_name = $2,
						source_platform = $3,
						source_rate_multiplier = $4,
						source_sort_order = $5,
						source_models = $6::jsonb,
						display_name = $7,
						display_multiplier = $8,
						enabled = $9,
						source_active = TRUE,
						is_virtual = TRUE,
						updated_at = NOW()
					WHERE id = $1 AND is_virtual = TRUE
				`,
					group.ID,
					sourceGroupName,
					platform,
					group.SourceRate,
					group.SourceSortOrder,
					mustMarshalJSON(sourceModels),
					displayName,
					displayMultiplier,
					group.Enabled,
				)
				if execErr != nil {
					err = execErr
				} else {
					var affected int64
					affected, err = result.RowsAffected()
					if err == nil && affected == 0 {
						err = fmt.Errorf("virtual model marketplace group %d not found", group.ID)
					}
				}
				groupConfigID = group.ID
			} else {
				err = tx.QueryRowContext(ctx, `
					INSERT INTO model_marketplace_group_configs (
						source_group_id, source_group_name, source_platform,
						source_rate_multiplier, source_sort_order, source_models,
						display_name, display_multiplier, enabled, source_active, is_virtual
					) VALUES (NULL, $1, $2, $3, $4, $5::jsonb, $6, $7, $8, TRUE, TRUE)
					RETURNING id
				`,
					sourceGroupName,
					platform,
					group.SourceRate,
					group.SourceSortOrder,
					mustMarshalJSON(sourceModels),
					displayName,
					displayMultiplier,
					group.Enabled,
				).Scan(&groupConfigID)
			}
		} else {
			if group.ID > 0 {
				result, execErr := tx.ExecContext(ctx, `
					UPDATE model_marketplace_group_configs
					SET display_name = $2,
						display_multiplier = $3,
						enabled = $4,
						is_virtual = FALSE,
						updated_at = NOW()
					WHERE id = $1 AND is_virtual = FALSE
				`, group.ID, displayName, displayMultiplier, group.Enabled)
				if execErr != nil {
					err = execErr
				} else {
					var affected int64
					affected, err = result.RowsAffected()
					if err == nil && affected == 0 {
						err = fmt.Errorf("model marketplace group %d not found", group.ID)
					}
				}
				groupConfigID = group.ID
			} else {
				err = tx.QueryRowContext(ctx, `
					SELECT id
					FROM model_marketplace_group_configs
					WHERE source_group_id = $1 AND is_virtual = FALSE
				`, group.SourceGroupID).Scan(&groupConfigID)
				if err == nil {
					_, err = tx.ExecContext(ctx, `
						UPDATE model_marketplace_group_configs
						SET display_name = $2,
							display_multiplier = $3,
							enabled = $4,
							updated_at = NOW()
						WHERE id = $1
					`, groupConfigID, displayName, displayMultiplier, group.Enabled)
				}
			}
		}
		if err != nil {
			return fmt.Errorf("save model marketplace group %s: %w", sourceGroupName, err)
		}
		savedGroupIDs = append(savedGroupIDs, groupConfigID)

		modelNames := make([]string, 0, len(group.Models))
		for index, model := range group.Models {
			modelName := strings.TrimSpace(model.SourceModelName)
			if modelName == "" {
				continue
			}
			modelNames = append(modelNames, modelName)
			modelIsVirtual := isVirtual || model.IsVirtual
			capabilityTags := nullableMarketplaceCapabilityTags(model.CapabilityTags)
			if _, err = tx.ExecContext(ctx, `
			INSERT INTO model_marketplace_model_configs (
				group_config_id, source_model_name, display_name, enabled, sort_order, is_virtual, capability_tags
			) VALUES ($1, $2, '', $3, $4, $5, $6::jsonb)
			ON CONFLICT (group_config_id, source_model_name) DO UPDATE SET
				display_name = '',
				enabled = EXCLUDED.enabled,
				sort_order = EXCLUDED.sort_order,
				is_virtual = EXCLUDED.is_virtual,
				capability_tags = EXCLUDED.capability_tags,
					updated_at = NOW()
			`, groupConfigID, modelName, model.Enabled, index, modelIsVirtual, capabilityTags); err != nil {
				return fmt.Errorf("save model marketplace model %s: %w", modelName, err)
			}
		}
		if len(modelNames) == 0 {
			_, err = tx.ExecContext(ctx, `
				DELETE FROM model_marketplace_model_configs
				WHERE group_config_id = $1
			`, groupConfigID)
		} else {
			_, err = tx.ExecContext(ctx, `
				DELETE FROM model_marketplace_model_configs
				WHERE group_config_id = $1
				  AND NOT (source_model_name = ANY($2::text[]))
			`, groupConfigID, pq.Array(modelNames))
		}
		if err != nil {
			return fmt.Errorf("remove deleted model marketplace models for group %d: %w", groupConfigID, err)
		}
	}

	if len(savedGroupIDs) == 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM model_marketplace_group_configs`)
	} else {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM model_marketplace_group_configs
			WHERE NOT (id = ANY($1::bigint[]))
		`, pqInt64Array(savedGroupIDs))
	}
	if err != nil {
		return fmt.Errorf("remove deleted model marketplace groups: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit model marketplace group save: %w", err)
	}
	return nil
}

func marketplaceModelNames(models []service.MarketplaceModelConfig) []string {
	names := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		name := strings.TrimSpace(model.SourceModelName)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names
}

func nullableMarketplaceCapabilityTags(tags []string) any {
	if tags == nil {
		return nil
	}
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		switch strings.ToLower(tag) {
		case "openai", "openai api", "openai-compatible", "openai-compatible api",
			"anthropic", "anthropic api", "gemini", "gemini api", "antigravity":
			continue
		}
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, tag)
	}
	return mustMarshalJSON(normalized)
}

func (r *modelMarketplaceRepository) ListConfiguredEnabled(ctx context.Context) ([]service.ModelMarketplaceItem, error) {
	var synced bool
	var count int
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			EXISTS (SELECT 1 FROM model_marketplace_sync_state WHERE id = TRUE),
			COUNT(*)
		FROM model_marketplace_group_configs
		WHERE is_virtual = FALSE
	`).Scan(&synced, &count); err != nil {
		return nil, fmt.Errorf("check configured model marketplace state: %w", err)
	}
	if !synced && count == 0 {
		return r.ListEnabled(ctx)
	}
	if synced && count == 0 {
		return []service.ModelMarketplaceItem{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(mi.id, 0),
			mc.source_model_name,
			mc.source_model_name,
			COALESCE(mi.pricing_aliases, '[]'::jsonb),
			COALESCE(mi.vendor_name, CASE
				WHEN LOWER(g.source_platform) = 'anthropic' THEN 'Anthropic'
				WHEN LOWER(g.source_platform) = 'openai' THEN 'OpenAI'
				ELSE g.source_platform
			END),
			COALESCE(NULLIF(g.display_name, ''), g.source_group_name),
			COALESCE(g.source_group_id, 0),
			g.display_multiplier,
			COALESCE(mc.capability_tags, COALESCE(mi.tags, '[]'::jsonb)),
			COALESCE(mi.endpoints, '[]'::jsonb),
			COALESCE(mi.description, ''),
			COALESCE(mi.official_prices, '{}'::jsonb),
			(g.source_sort_order * 100000) + mc.sort_order
		FROM model_marketplace_group_configs g
		JOIN model_marketplace_model_configs mc ON mc.group_config_id = g.id
		JOIN groups current_group ON current_group.id = g.source_group_id
		LEFT JOIN model_marketplace_items mi ON
			LOWER(mi.model_name) = LOWER(mc.source_model_name)
			OR EXISTS (
				SELECT 1
				FROM jsonb_array_elements_text(mi.pricing_aliases) AS pricing_alias(alias_name)
				WHERE LOWER(pricing_alias.alias_name) = LOWER(mc.source_model_name)
			)
		WHERE g.enabled = TRUE
		  AND g.is_virtual = FALSE
		  AND current_group.status = 'active'
		  AND current_group.deleted_at IS NULL
		  AND mc.enabled = TRUE
		  AND mc.is_virtual = FALSE
		  AND (mi.id IS NULL OR mi.enabled = TRUE)
		  AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements_text(g.source_models) AS source_model(model_name)
				WHERE source_model.model_name = mc.source_model_name
		  )
		ORDER BY g.source_sort_order ASC, g.source_group_id ASC, mc.sort_order ASC, mc.source_model_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list configured model marketplace items: %w", err)
	}
	defer rows.Close()

	items := make([]service.ModelMarketplaceItem, 0)
	for rows.Next() {
		var item service.ModelMarketplaceItem
		var pricingAliasesJSON, tagsJSON, endpointsJSON, officialPricesJSON []byte
		var groupName, displayName string
		if err := rows.Scan(
			&item.ID,
			&item.ModelName,
			&displayName,
			&pricingAliasesJSON,
			&item.VendorName,
			&groupName,
			&item.SourceGroupID,
			&item.GroupMultiplier,
			&tagsJSON,
			&endpointsJSON,
			&item.Description,
			&officialPricesJSON,
			&item.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan configured model marketplace item: %w", err)
		}
		item.DisplayName = strings.TrimSpace(displayName)
		item.Groups = []string{strings.TrimSpace(groupName)}
		item.Enabled = true
		if err := unmarshalJSONList(pricingAliasesJSON, &item.PricingAliases); err != nil {
			return nil, fmt.Errorf("decode configured pricing aliases for %s: %w", item.ModelName, err)
		}
		if err := unmarshalJSONList(tagsJSON, &item.Tags); err != nil {
			return nil, fmt.Errorf("decode configured tags for %s: %w", item.ModelName, err)
		}
		if err := unmarshalJSONList(endpointsJSON, &item.Endpoints); err != nil {
			return nil, fmt.Errorf("decode configured endpoints for %s: %w", item.ModelName, err)
		}
		if err := json.Unmarshal(officialPricesJSON, &item.OfficialPrices); err != nil {
			return nil, fmt.Errorf("decode configured official prices for %s: %w", item.ModelName, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate configured model marketplace items: %w", err)
	}
	return items, nil
}

func (r *modelMarketplaceRepository) UpdateGroup(
	ctx context.Context,
	sourceGroupID int64,
	input service.UpdateMarketplaceGroupInput,
) (*service.MarketplaceGroupConfig, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin model marketplace group update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var groupConfigID int64
	if err = tx.QueryRowContext(ctx, `
		UPDATE model_marketplace_group_configs
		SET display_name = $2,
			display_multiplier = $3,
			enabled = $4,
			updated_at = NOW()
		WHERE source_group_id = $1
		RETURNING id
	`, sourceGroupID, strings.TrimSpace(input.DisplayName), input.DisplayMultiplier, input.Enabled).Scan(&groupConfigID); err != nil {
		return nil, fmt.Errorf("update model marketplace group %d: %w", sourceGroupID, err)
	}

	for _, model := range input.Models {
		capabilityTags := nullableMarketplaceCapabilityTags(model.CapabilityTags)
		if _, err = tx.ExecContext(ctx, `
			UPDATE model_marketplace_model_configs
			SET display_name = $3, enabled = $4, sort_order = $5, capability_tags = $6::jsonb
			WHERE group_config_id = $1 AND source_model_name = $2
		`, groupConfigID, strings.TrimSpace(model.SourceModelName), strings.TrimSpace(model.DisplayName), model.Enabled, model.SortOrder, capabilityTags); err != nil {
			return nil, fmt.Errorf("update model marketplace model %s for group %d: %w", model.SourceModelName, sourceGroupID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit model marketplace group update: %w", err)
	}
	groups, err := r.ListAdmin(ctx)
	if err != nil {
		return nil, err
	}
	for index := range groups {
		if groups[index].SourceGroupID == sourceGroupID {
			return &groups[index], nil
		}
	}
	return nil, fmt.Errorf("model marketplace group %d disappeared after update", sourceGroupID)
}

func mustMarshalJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func pqInt64Array(values []int64) any {
	return pq.Array(values)
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
