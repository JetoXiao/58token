package service

import "testing"

func TestCalculateRechargeBonus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		paymentAmount float64
		threshold     float64
		bonusAmount   float64
		want          float64
	}{
		{name: "below threshold", paymentAmount: 99.99, threshold: 100, bonusAmount: 10, want: 0},
		{name: "one step", paymentAmount: 100, threshold: 100, bonusAmount: 10, want: 10},
		{name: "multiple steps", paymentAmount: 250, threshold: 100, bonusAmount: 10, want: 20},
		{name: "disabled threshold", paymentAmount: 250, threshold: 0, bonusAmount: 10, want: 0},
		{name: "disabled amount", paymentAmount: 250, threshold: 100, bonusAmount: 0, want: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := calculateRechargeBonus(tt.paymentAmount, tt.threshold, tt.bonusAmount); got != tt.want {
				t.Fatalf("calculateRechargeBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateCreditedBalanceIncludesRechargeBonus(t *testing.T) {
	t.Parallel()

	got := calculateCreditedBalance(250, 1, 100, 10)
	if got != 270 {
		t.Fatalf("calculateCreditedBalance() = %v, want 270", got)
	}
}
