package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestExpireTimedOutOrdersAutoCancelsPendingOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderTimeoutTestClient(t)

	user, err := client.User.Create().
		SetEmail("auto-cancel-timeout@example.com").
		SetPasswordHash("hash").
		SetUsername("auto-cancel-timeout-user").
		Save(ctx)
	require.NoError(t, err)

	expiredOrder, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(99).
		SetPayAmount(99).
		SetFeeRate(0).
		SetRechargeCode("TIMEOUT-CANCEL").
		SetOutTradeNo("sub2_timeout_cancel").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(-time.Minute)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	activeOrder, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(129).
		SetPayAmount(129).
		SetFeeRate(0).
		SetRechargeCode("ACTIVE-PENDING").
		SetOutTradeNo("sub2_active_pending").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient:       client,
		registry:        payment.NewRegistry(),
		providersLoaded: true,
	}

	cancelled, err := svc.ExpireTimedOutOrders(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, cancelled)

	reloadedExpired, err := client.PaymentOrder.Get(ctx, expiredOrder.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCancelled, reloadedExpired.Status)

	reloadedActive, err := client.PaymentOrder.Get(ctx, activeOrder.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPending, reloadedActive.Status)
}

func newPaymentOrderTimeoutTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:payment_order_timeout?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}
