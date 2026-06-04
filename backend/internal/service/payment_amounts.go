package service

import (
	"math"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/shopspring/decimal"
)

const (
	defaultBalanceRechargeMultiplier = 1.0
	defaultUsdtCnyExchangeRate       = 7.0
	defaultRechargeBonusThreshold    = 100.0
	defaultRechargeBonusAmount       = 10.0
)

func normalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier <= 0 {
		return defaultBalanceRechargeMultiplier
	}
	return multiplier
}

func normalizeUsdtCnyExchangeRate(rate float64) float64 {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate <= 0 {
		return defaultUsdtCnyExchangeRate
	}
	return rate
}

func normalizeRechargeBonusThreshold(threshold float64) float64 {
	if math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold < 0 {
		return defaultRechargeBonusThreshold
	}
	return threshold
}

func normalizeRechargeBonusAmount(amount float64) float64 {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < 0 {
		return defaultRechargeBonusAmount
	}
	return amount
}

func calculateRechargeBonus(paymentAmount, threshold, bonusAmount float64) float64 {
	threshold = normalizeRechargeBonusThreshold(threshold)
	bonusAmount = normalizeRechargeBonusAmount(bonusAmount)
	if paymentAmount <= 0 || threshold <= 0 || bonusAmount <= 0 {
		return 0
	}
	return decimal.NewFromFloat(paymentAmount).
		Div(decimal.NewFromFloat(threshold)).
		Floor().
		Mul(decimal.NewFromFloat(bonusAmount)).
		Round(2).
		InexactFloat64()
}

func calculateCreditedBalance(paymentAmount, multiplier, bonusThreshold, bonusAmount float64) float64 {
	base := decimal.NewFromFloat(paymentAmount).
		Mul(decimal.NewFromFloat(normalizeBalanceRechargeMultiplier(multiplier)))
	bonus := decimal.NewFromFloat(calculateRechargeBonus(paymentAmount, bonusThreshold, bonusAmount))
	return base.Add(bonus).
		Round(2).
		InexactFloat64()
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	if orderAmount <= 0 || payAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	fractionDigits := int32(payment.CurrencyMaxFractionDigits(currency))
	if math.Abs(refundAmount-orderAmount) <= paymentAmountToleranceForCurrency(currency) {
		return decimal.NewFromFloat(payAmount).Round(fractionDigits).InexactFloat64()
	}
	return decimal.NewFromFloat(payAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(fractionDigits).
		InexactFloat64()
}
