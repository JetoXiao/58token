package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestInfiniCreatePaymentReturnsCheckoutURLAndSignsRequest(t *testing.T) {
	var authHeader string
	var digestHeader string
	var dateHeader string
	var requestBody map[string]any

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, infiniOrderCreatePath, r.URL.Path)
		authHeader = r.Header.Get("Authorization")
		digestHeader = r.Header.Get("Digest")
		dateHeader = r.Header.Get("Date")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))
		require.Equal(t, "sub2_test_order", requestBody["client_reference"])
		require.Equal(t, []any{float64(infiniCryptoPayMethod)}, requestBody["pay_methods"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order_id":"ord_infini_123","request_id":"req_1","checkout_url":"https://checkout.infini.money/pay/abc","client_reference":"sub2_test_order"}`))
	}))
	defer server.Close()

	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       "https://openapi.infini.money",
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)
	prov.config["apiBase"] = server.URL
	prov.httpClient = server.Client()

	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:   "sub2_test_order",
		Amount:    "4.1667",
		ReturnURL: "https://app.example.com/payment/result",
		Subject:   "Balance recharge",
	})
	require.NoError(t, err)
	require.Equal(t, "ord_infini_123", resp.TradeNo)
	require.Equal(t, "https://checkout.infini.money/pay/abc", resp.PayURL)
	require.Equal(t, payment.TypeUSDT, resp.Currency)
	require.Contains(t, authHeader, `Signature keyId="merchant-001"`)
	require.NotEmpty(t, digestHeader)
	require.NotEmpty(t, dateHeader)
}

func TestInfiniCreatePaymentParsesWrappedResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, infiniOrderCreatePath, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"order_id":"ord_wrapped_123","request_id":"req_1","checkout_url":"https://checkout.infini.money/pay/wrapped","client_reference":"sub2_test_order"}}`))
	}))
	defer server.Close()

	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniProdAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)
	prov.config["apiBase"] = server.URL
	prov.httpClient = server.Client()

	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:   "sub2_test_order",
		Amount:    "4.1667",
		ReturnURL: "https://app.example.com/payment/result",
		Subject:   "Balance recharge",
	})
	require.NoError(t, err)
	require.Equal(t, "ord_wrapped_123", resp.TradeNo)
	require.Equal(t, "https://checkout.infini.money/pay/wrapped", resp.PayURL)
}

func TestInfiniVerifyNotificationMapsCompletedOrder(t *testing.T) {
	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniSandboxAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)

	raw := `{"event":"order.completed","order_id":"ord_123","client_reference":"sub2_abc","amount":"4.1667","currency":"USD","status":"paid","amount_confirmed":"4.1667"}`
	headers := signedInfiniWebhookHeaders(raw, "whsec", time.Now())
	notification, err := prov.VerifyNotification(context.Background(), raw, headers)
	require.NoError(t, err)
	require.NotNil(t, notification)
	require.Equal(t, "ord_123", notification.TradeNo)
	require.Equal(t, "sub2_abc", notification.OrderID)
	require.InDelta(t, 4.1667, notification.Amount, 0.000001)
	require.Equal(t, payment.NotificationStatusSuccess, notification.Status)
	require.Equal(t, payment.TypeUSDT, notification.Metadata["currency"])
}

func TestInfiniVerifyNotificationUsesConfirmedAmount(t *testing.T) {
	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniSandboxAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)

	raw := `{"event":"order.completed","order_id":"ord_123","client_reference":"sub2_abc","amount":"2.00","currency":"USD","status":"paid","amount_confirmed":"0.2858"}`
	headers := signedInfiniWebhookHeaders(raw, "whsec", time.Now())
	notification, err := prov.VerifyNotification(context.Background(), raw, headers)
	require.NoError(t, err)
	require.NotNil(t, notification)
	require.InDelta(t, 0.2858, notification.Amount, 0.000001)
}

func TestInfiniQueryOrderUsesPayStatusForSettlement(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, infiniOrderCreatePath, r.URL.Path)
		require.Equal(t, "ord_paid_123", r.URL.Query().Get("order_id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order_id":"ord_paid_123","status":"processing","pay_status":"paid","amount":"4.1667","currency":"USD","amount_confirmed":"4.1667","client_reference":"sub2_abc"}`))
	}))
	defer server.Close()

	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniSandboxAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)
	prov.config["apiBase"] = server.URL
	prov.httpClient = server.Client()

	resp, err := prov.QueryOrder(context.Background(), "ord_paid_123")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, resp.Status)
	require.InDelta(t, 4.1667, resp.Amount, 0.000001)
	require.Equal(t, "sub2_abc", resp.Metadata["client_reference"])
}

func TestInfiniQueryOrderUsesConfirmedAmountAndOverpaidStatus(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, infiniOrderCreatePath, r.URL.Path)
		require.Equal(t, "ord_overpaid_123", r.URL.Query().Get("order_id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order_id":"ord_overpaid_123","status":"overpaid","amount":"2.00","currency":"USD","amount_confirmed":"0.2858","client_reference":"sub2_abc"}`))
	}))
	defer server.Close()

	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniSandboxAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)
	prov.config["apiBase"] = server.URL
	prov.httpClient = server.Client()

	resp, err := prov.QueryOrder(context.Background(), "ord_overpaid_123")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, resp.Status)
	require.InDelta(t, 0.2858, resp.Amount, 0.000001)
}

func TestInfiniVerifyNotificationParsesWrappedPayload(t *testing.T) {
	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniSandboxAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)

	raw := `{"event":"order.completed","data":{"order_id":"ord_wrapped_123","client_reference":"sub2_wrapped","amount":"4.1667","currency":"USD","status":"paid","amount_confirmed":"4.1667"}}`
	headers := signedInfiniWebhookHeaders(raw, "whsec", time.Now())
	notification, err := prov.VerifyNotification(context.Background(), raw, headers)
	require.NoError(t, err)
	require.NotNil(t, notification)
	require.Equal(t, "ord_wrapped_123", notification.TradeNo)
	require.Equal(t, "sub2_wrapped", notification.OrderID)
	require.InDelta(t, 4.1667, notification.Amount, 0.000001)
}

func TestInfiniVerifyNotificationRejectsBadSignature(t *testing.T) {
	prov, err := NewInfini("inst_1", map[string]string{
		"keyId":         "merchant-001",
		"secretKey":     "secret",
		"webhookSecret": "whsec",
		"apiBase":       infiniSandboxAPIBase,
		"currency":      payment.TypeUSDT,
	})
	require.NoError(t, err)

	raw := `{"event":"order.completed","order_id":"ord_123","client_reference":"sub2_abc","amount":"4.1667","status":"paid"}`
	headers := signedInfiniWebhookHeaders(raw, "wrong-secret", time.Now())
	_, err = prov.VerifyNotification(context.Background(), raw, headers)
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "signature")
}

func signedInfiniWebhookHeaders(rawBody, secret string, now time.Time) map[string]string {
	timestamp := strconv.FormatInt(now.UTC().Unix(), 10)
	eventID := "evt_test_123"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "." + eventID + "." + rawBody))
	return map[string]string{
		"x-webhook-timestamp": timestamp,
		"x-webhook-event-id":  eventID,
		"x-webhook-signature": hex.EncodeToString(mac.Sum(nil)),
	}
}
