package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

// --- Dashboard & Analytics ---

func (s *PaymentService) GetDashboardStats(ctx context.Context, days int) (*DashboardStats, error) {
	if days <= 0 {
		days = 30
	}
	now := time.Now()
	since := now.AddDate(0, 0, -days)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	paidStatuses := []string{OrderStatusCompleted, OrderStatusPaid, OrderStatusRecharging}

	orders, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusIn(paidStatuses...),
			paymentorder.PaidAtGTE(since),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	exchangeRate := s.dashboardUsdtCnyExchangeRate(ctx)
	st := &DashboardStats{}
	computeBasicStats(st, orders, todayStart, exchangeRate)

	st.PendingOrders, err = s.entClient.PaymentOrder.Query().
		Where(paymentorder.StatusEQ(OrderStatusPending)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	st.DailySeries = buildDailySeries(orders, since, days, exchangeRate)
	st.PaymentMethods = buildMethodDistribution(orders, exchangeRate)
	st.TopUsers = buildTopUsers(orders, exchangeRate)

	return st, nil
}

func (s *PaymentService) dashboardUsdtCnyExchangeRate(ctx context.Context) float64 {
	if s == nil || s.configService == nil {
		return normalizeUsdtCnyExchangeRate(defaultUsdtCnyExchangeRate)
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil || cfg == nil {
		if err != nil {
			slog.Warn("payment dashboard: failed to load exchange rate, using default", "error", err)
		}
		return normalizeUsdtCnyExchangeRate(defaultUsdtCnyExchangeRate)
	}
	return normalizeUsdtCnyExchangeRate(cfg.UsdtCnyExchangeRate)
}

func computeBasicStats(st *DashboardStats, orders []*dbent.PaymentOrder, todayStart time.Time, exchangeRate float64) {
	var totalAmount, todayAmount float64
	var todayCount int
	for _, o := range orders {
		amountCNY := paymentOrderAmountCNY(o, exchangeRate)
		totalAmount += amountCNY
		if o.PaidAt != nil && !o.PaidAt.Before(todayStart) {
			todayAmount += amountCNY
			todayCount++
		}
	}
	st.TotalAmount = math.Round(totalAmount*100) / 100
	st.TodayAmount = math.Round(todayAmount*100) / 100
	st.TotalCount = len(orders)
	st.TodayCount = todayCount
	if st.TotalCount > 0 {
		st.AvgAmount = math.Round(totalAmount/float64(st.TotalCount)*100) / 100
	}
}

func buildDailySeries(orders []*dbent.PaymentOrder, since time.Time, days int, exchangeRate float64) []DailyStats {
	dailyMap := make(map[string]*DailyStats)
	for _, o := range orders {
		if o.PaidAt == nil {
			continue
		}
		date := o.PaidAt.Format("2006-01-02")
		ds, ok := dailyMap[date]
		if !ok {
			ds = &DailyStats{Date: date}
			dailyMap[date] = ds
		}
		ds.Amount += paymentOrderAmountCNY(o, exchangeRate)
		ds.Count++
	}
	series := make([]DailyStats, 0, days)
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i+1).Format("2006-01-02")
		if ds, ok := dailyMap[date]; ok {
			ds.Amount = math.Round(ds.Amount*100) / 100
			series = append(series, *ds)
		} else {
			series = append(series, DailyStats{Date: date})
		}
	}
	return series
}

func buildMethodDistribution(orders []*dbent.PaymentOrder, exchangeRate float64) []PaymentMethodStat {
	methodMap := make(map[string]*PaymentMethodStat)
	for _, o := range orders {
		ms, ok := methodMap[o.PaymentType]
		if !ok {
			ms = &PaymentMethodStat{Type: o.PaymentType, Currency: paymentOrderDisplayCurrency(o)}
			methodMap[o.PaymentType] = ms
		}
		ms.Amount += o.PayAmount
		ms.AmountCNY += paymentOrderAmountCNY(o, exchangeRate)
		ms.Count++
	}
	methods := make([]PaymentMethodStat, 0, len(methodMap))
	for _, ms := range methodMap {
		ms.Amount = math.Round(ms.Amount*100) / 100
		ms.AmountCNY = math.Round(ms.AmountCNY*100) / 100
		methods = append(methods, *ms)
	}
	return methods
}

func buildTopUsers(orders []*dbent.PaymentOrder, exchangeRate float64) []TopUserStat {
	userMap := make(map[int64]*TopUserStat)
	for _, o := range orders {
		us, ok := userMap[o.UserID]
		if !ok {
			us = &TopUserStat{UserID: o.UserID, Email: o.UserEmail}
			userMap[o.UserID] = us
		}
		if paymentOrderIsUSDT(o) {
			us.USDTAmount += o.PayAmount
		} else {
			us.RMBAmount += o.PayAmount
		}
		us.Amount += paymentOrderAmountCNY(o, exchangeRate)
	}
	userList := make([]*TopUserStat, 0, len(userMap))
	for _, us := range userMap {
		us.Amount = math.Round(us.Amount*100) / 100
		us.RMBAmount = math.Round(us.RMBAmount*100) / 100
		us.USDTAmount = math.Round(us.USDTAmount*100) / 100
		userList = append(userList, us)
	}
	sort.Slice(userList, func(i, j int) bool {
		return userList[i].Amount > userList[j].Amount
	})
	limit := topUsersLimit
	if len(userList) < limit {
		limit = len(userList)
	}
	result := make([]TopUserStat, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, *userList[i])
	}
	return result
}

func paymentOrderAmountCNY(o *dbent.PaymentOrder, exchangeRate float64) float64 {
	if o == nil {
		return 0
	}
	if paymentOrderIsUSDT(o) {
		return o.PayAmount * normalizeUsdtCnyExchangeRate(exchangeRate)
	}
	return o.PayAmount
}

func paymentOrderDisplayCurrency(o *dbent.PaymentOrder) string {
	if paymentOrderIsUSDT(o) {
		return "USD"
	}
	return payment.DefaultPaymentCurrency
}

func paymentOrderIsUSDT(o *dbent.PaymentOrder) bool {
	if o == nil {
		return false
	}
	if payment.GetBasePaymentType(o.PaymentType) == payment.TypeUSDT {
		return true
	}
	if o.ProviderKey != nil && payment.GetBasePaymentType(*o.ProviderKey) == payment.TypeUSDT {
		return true
	}
	return PaymentOrderCurrency(o) == payment.TypeUSDT
}

// --- Audit Logs ---

func (s *PaymentService) writeAuditLog(ctx context.Context, oid int64, action, op string, detail map[string]any) {
	dj, _ := json.Marshal(detail)
	_, err := s.entClient.PaymentAuditLog.Create().SetOrderID(strconv.FormatInt(oid, 10)).SetAction(action).SetDetail(string(dj)).SetOperator(op).Save(ctx)
	if err != nil {
		slog.Error("audit log failed", "orderID", oid, "action", action, "error", err)
	}
}

func (s *PaymentService) GetOrderAuditLogs(ctx context.Context, oid int64) ([]*dbent.PaymentAuditLog, error) {
	return s.entClient.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(oid, 10))).Order(paymentauditlog.ByCreatedAt()).All(ctx)
}
