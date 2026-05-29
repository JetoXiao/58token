package service

import (
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestPaymentDashboardStatsUsesRMBReportingCurrency(t *testing.T) {
	now := time.Now()
	orders := []*dbent.PaymentOrder{
		{UserID: 1, UserEmail: "rmb@example.com", PaymentType: payment.TypeAlipay, PayAmount: 100, PaidAt: &now},
		{UserID: 2, UserEmail: "usdt@example.com", PaymentType: payment.TypeUSDT, PayAmount: 2, PaidAt: &now},
	}

	st := &DashboardStats{}
	computeBasicStats(st, orders, now.Add(-time.Hour), 7)
	require.Equal(t, 114.0, st.TotalAmount)
	require.Equal(t, 114.0, st.TodayAmount)
	require.Equal(t, 57.0, st.AvgAmount)

	methods := buildMethodDistribution(orders, 7)
	require.Len(t, methods, 2)
	byType := map[string]PaymentMethodStat{}
	for _, method := range methods {
		byType[method.Type] = method
	}
	require.Equal(t, "CNY", byType[payment.TypeAlipay].Currency)
	require.Equal(t, 100.0, byType[payment.TypeAlipay].Amount)
	require.Equal(t, 100.0, byType[payment.TypeAlipay].AmountCNY)
	require.Equal(t, "USD", byType[payment.TypeUSDT].Currency)
	require.Equal(t, 2.0, byType[payment.TypeUSDT].Amount)
	require.Equal(t, 14.0, byType[payment.TypeUSDT].AmountCNY)

	topUsers := buildTopUsers(orders, 7)
	require.Len(t, topUsers, 2)
	require.Equal(t, int64(1), topUsers[0].UserID)
	require.Equal(t, 100.0, topUsers[0].Amount)
	require.Equal(t, 100.0, topUsers[0].RMBAmount)
	require.Equal(t, 14.0, topUsers[1].Amount)
	require.Equal(t, 2.0, topUsers[1].USDTAmount)
}
