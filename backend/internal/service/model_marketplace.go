package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type MarketplacePriceTier struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type MarketplaceOfficialPrices struct {
	Input      json.RawMessage `json:"input"`
	Output     json.RawMessage `json:"output"`
	CacheWrite json.RawMessage `json:"cacheWrite"`
	CacheRead  json.RawMessage `json:"cacheRead"`
}

type ModelMarketplaceItem struct {
	ID             int64                     `json:"id"`
	ModelName      string                    `json:"model_name"`
	PricingAliases []string                  `json:"pricing_aliases"`
	VendorName     string                    `json:"vendor_name"`
	Groups         []string                  `json:"groups"`
	Tags           []string                  `json:"tags"`
	Endpoints      []string                  `json:"endpoints"`
	Description    string                    `json:"description"`
	OfficialPrices MarketplaceOfficialPrices `json:"official_prices"`
	SortOrder      int                       `json:"sort_order"`
	Enabled        bool                      `json:"enabled"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

type ModelMarketplaceRepository interface {
	ListEnabled(ctx context.Context) ([]ModelMarketplaceItem, error)
}

type ModelMarketplaceService struct {
	repo ModelMarketplaceRepository
}

var retiredMarketplaceModels = map[string]struct{}{
	"gpt-5.2":       {},
	"gpt-5.3-codex": {},
}

func NewModelMarketplaceService(repo ModelMarketplaceRepository) *ModelMarketplaceService {
	return &ModelMarketplaceService{repo: repo}
}

func (s *ModelMarketplaceService) ListEnabled(ctx context.Context) ([]ModelMarketplaceItem, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("model marketplace repository is not configured")
	}
	items, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	filtered := items[:0]
	for _, item := range items {
		if IsRetiredMarketplaceModel(item.ModelName) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func IsRetiredMarketplaceModel(modelName string) bool {
	_, ok := retiredMarketplaceModels[strings.ToLower(strings.TrimSpace(modelName))]
	return ok
}
