package service

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/assert"
)

func TestAffiliateRebateBaseAmountUsesActualPaidRMBAmount(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
		Amount:      30,
		PayAmount:   30,
		FeeRate:     0,
	}

	baseAmount, currency, eligible := (&PaymentService{}).affiliateRebateBaseAmountUSD(context.Background(), order)

	assert.True(t, eligible)
	assert.Equal(t, "USD", currency)
	assert.InDelta(t, 30, baseAmount, 1e-9)
	assert.InDelta(t, 1.5, roundTo(baseAmount*0.05, 8), 1e-9)
}

func TestAffiliateRebateBaseAmountConvertsActualPaidUSDTToRMBEquivalent(t *testing.T) {
	t.Parallel()

	providerKey := payment.TypeInfini
	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeUSDT,
		ProviderKey: &providerKey,
		ProviderSnapshot: map[string]any{
			"schema_version": 2,
			"provider_key":   payment.TypeInfini,
			"currency":       payment.TypeUSDT,
		},
		Amount:    30,
		PayAmount: 4.2857,
		FeeRate:   0,
	}

	baseAmount, currency, eligible := (&PaymentService{}).affiliateRebateBaseAmountUSD(context.Background(), order)

	assert.True(t, eligible)
	assert.Equal(t, "USD", currency)
	assert.InDelta(t, 29.9999, baseAmount, 1e-4)
	assert.InDelta(t, 1.5, roundTo(baseAmount*0.05, 2), 1e-9)
}

func TestAffiliateEligiblePaidAmountDoesNotRemovePaymentFee(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
		PayAmount:   103,
		FeeRate:     3,
	}

	assert.InDelta(t, 103, affiliateEligiblePaidAmount(order), 1e-9)
}
