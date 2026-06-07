package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAIImageCostUsesExplicitGroupPriceBeforeChannelPrice(t *testing.T) {
	groupID := int64(901)
	groupPrice := 0.10
	svc := &OpenAIGatewayService{
		billingService: NewBillingService(&config.Config{}, nil),
		resolver:       newImagePriorityResolver(groupID, "gpt-image-2", 0.25),
	}

	cost := svc.calculateOpenAIImageCost(
		context.Background(),
		"gpt-image-2",
		&APIKey{
			GroupID: &groupID,
			Group: &Group{
				ID:           groupID,
				ImagePrice1K: &groupPrice,
			},
		},
		&OpenAIForwardResult{
			Model:      "gpt-image-2",
			ImageCount: 3,
			ImageSize:  ImageBillingSize1K,
		},
		1,
	)

	require.NotNil(t, cost)
	require.Equal(t, string(BillingModeImage), cost.BillingMode)
	require.InDelta(t, 0.30, cost.TotalCost, 1e-12)
	require.InDelta(t, 0.30, cost.ActualCost, 1e-12)
}

func TestGatewayImageCostUsesExplicitGroupPriceBeforeChannelPrice(t *testing.T) {
	groupID := int64(902)
	groupPrice := 0.15
	svc := &GatewayService{
		billingService: NewBillingService(&config.Config{}, nil),
		resolver:       newImagePriorityResolver(groupID, "gemini-image", 0.30),
	}

	cost := svc.calculateRecordUsageCost(
		context.Background(),
		&ForwardResult{
			Model:      "gemini-image",
			ImageCount: 2,
			ImageSize:  ImageBillingSize2K,
		},
		&APIKey{
			GroupID: &groupID,
			Group: &Group{
				ID:           groupID,
				ImagePrice2K: &groupPrice,
			},
		},
		"gemini-image",
		1,
		1,
		nil,
	)

	require.NotNil(t, cost)
	require.Equal(t, string(BillingModeImage), cost.BillingMode)
	require.InDelta(t, 0.30, cost.TotalCost, 1e-12)
	require.InDelta(t, 0.30, cost.ActualCost, 1e-12)
}

func newImagePriorityResolver(groupID int64, model string, channelPrice float64) *ModelPricingResolver {
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, model: model}] = &ChannelModelPricing{
		BillingMode:     BillingModeImage,
		PerRequestPrice: &channelPrice,
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.loadedAt = time.Now()

	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	return NewModelPricingResolver(channelService, billingService)
}
