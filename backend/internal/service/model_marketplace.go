package service

import (
	"context"
	"encoding/json"
	"fmt"
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

func NewModelMarketplaceService(repo ModelMarketplaceRepository) *ModelMarketplaceService {
	return &ModelMarketplaceService{repo: repo}
}

func (s *ModelMarketplaceService) ListEnabled(ctx context.Context) ([]ModelMarketplaceItem, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("model marketplace repository is not configured")
	}
	return s.repo.ListEnabled(ctx)
}
