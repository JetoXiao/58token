package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	infiniProdAPIBase      = "https://openapi.infini.money"
	infiniSandboxAPIBase   = "https://openapi-sandbox.infini.money"
	infiniHTTPTimeout      = 15 * time.Second
	infiniMaxResponseSize  = 1 << 20
	infiniMaxErrorSummary  = 512
	infiniWebhookTolerance = 5 * time.Minute
	infiniDNSFallbackDelay = 3 * time.Second

	infiniOrderCreatePath = "/v1/acquiring/order"
	infiniCryptoPayMethod = 1

	infiniEventOrderCompleted   = "order.completed"
	infiniEventOrderProcessing  = "order.processing"
	infiniEventOrderExpired     = "order.expired"
	infiniEventOrderLatePayment = "order.late_payment"
	infiniStatusPaid            = "paid"
	infiniStatusProcessing      = "processing"
	infiniStatusPending         = "pending"
	infiniStatusExpired         = "expired"
	infiniStatusPartialPaid     = "partial_paid"
)

type Infini struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

func NewInfini(instanceID string, config map[string]string) (*Infini, error) {
	for _, k := range []string{"keyId", "secretKey", "webhookSecret", "apiBase"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("infini config missing required key: %s", k)
		}
	}
	cfg := cloneStringMap(config)
	apiBase, err := normalizeInfiniAPIBase(cfg["apiBase"])
	if err != nil {
		return nil, err
	}
	cfg["apiBase"] = apiBase
	currency, err := payment.NormalizePaymentCurrency(cfg["currency"])
	if err != nil {
		return nil, fmt.Errorf("infini config currency: %w", err)
	}
	if currency != strings.ToUpper(payment.TypeUSDT) {
		return nil, fmt.Errorf("infini config currency must be USDT")
	}
	cfg["currency"] = currency
	if strings.TrimSpace(cfg["fiatCurrency"]) == "" {
		cfg["fiatCurrency"] = "USD"
	}
	return &Infini{
		instanceID: instanceID,
		config:     cfg,
		httpClient: newInfiniHTTPClient(),
	}, nil
}

func newInfiniHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = infiniDialContext
	return &http.Client{
		Timeout:   infiniHTTPTimeout,
		Transport: transport,
	}
}

func infiniDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: infiniHTTPTimeout}
	conn, err := dialer.DialContext(ctx, network, address)
	if err == nil || !isInfiniDNSLookupError(err) {
		return conn, err
	}

	host, port, splitErr := net.SplitHostPort(address)
	if splitErr != nil || !isInfiniAPIHost(host) {
		return nil, err
	}

	ips, resolveErr := lookupInfiniHostWithFallbackDNS(ctx, host)
	if resolveErr != nil || len(ips) == 0 {
		return nil, err
	}

	lastErr := err
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, lastErr
}

func isInfiniDNSLookupError(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such host") || strings.Contains(msg, "lookup")
}

func isInfiniAPIHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "openapi.infini.money" || host == "openapi-sandbox.infini.money"
}

func lookupInfiniHostWithFallbackDNS(ctx context.Context, host string) ([]net.IPAddr, error) {
	var lastErr error
	for _, server := range []string{
		"223.5.5.5:53",
		"119.29.29.29:53",
		"1.1.1.1:53",
		"8.8.8.8:53",
	} {
		server := server
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: infiniDNSFallbackDelay}
				return d.DialContext(ctx, network, server)
			},
		}
		ips, err := resolver.LookupIPAddr(ctx, host)
		if err == nil && len(ips) > 0 {
			return ips, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func normalizeInfiniAPIBase(raw string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = infiniProdAPIBase
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", fmt.Errorf("infini apiBase must be an HTTPS URL")
	}
	host := strings.ToLower(parsed.Host)
	if host != "openapi.infini.money" && host != "openapi-sandbox.infini.money" {
		return "", fmt.Errorf("infini apiBase host must be openapi.infini.money or openapi-sandbox.infini.money")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.RawPath = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "/v1/acquiring" {
		parsed.Path = ""
	}
	if parsed.Path != "" {
		return "", fmt.Errorf("infini apiBase must not include a path")
	}
	return parsed.String(), nil
}

func (i *Infini) Name() string        { return "Infini" }
func (i *Infini) ProviderKey() string { return payment.TypeInfini }
func (i *Infini) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeUSDT}
}

func (i *Infini) MerchantIdentityMetadata() map[string]string {
	if i == nil {
		return nil
	}
	return map[string]string{
		"currency":      payment.TypeUSDT,
		"fiat_currency": strings.ToUpper(strings.TrimSpace(i.config["fiatCurrency"])),
	}
}

func (i *Infini) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("infini create order: invalid amount %s", req.Amount)
	}

	payload := infiniCreateOrderRequest{
		Amount:          amount.String(),
		Currency:        strings.ToUpper(strings.TrimSpace(i.config["fiatCurrency"])),
		RequestID:       uuid.NewString(),
		ClientReference: req.OrderID,
		OrderDesc:       req.Subject,
		SuccessURL:      req.ReturnURL,
		FailureURL:      req.ReturnURL,
		PayMethods:      []int{infiniCryptoPayMethod},
	}
	if alias := strings.TrimSpace(i.config["merchantAlias"]); alias != "" {
		payload.MerchantAlias = alias
	}

	var created infiniCreateOrderResponse
	if err := i.doJSON(ctx, http.MethodPost, infiniOrderCreatePath, payload, &created); err != nil {
		return nil, fmt.Errorf("infini create order: %w", err)
	}
	if strings.TrimSpace(created.OrderID) == "" || strings.TrimSpace(created.CheckoutURL) == "" {
		return nil, fmt.Errorf("infini create order: missing order_id or checkout_url")
	}
	return &payment.CreatePaymentResponse{
		TradeNo:    created.OrderID,
		PayURL:     created.CheckoutURL,
		Currency:   payment.TypeUSDT,
		ResultType: payment.CreatePaymentResultOrderCreated,
	}, nil
}

func (i *Infini) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	orderID := strings.TrimSpace(tradeNo)
	if orderID == "" {
		return nil, fmt.Errorf("infini query order: missing order id")
	}
	path := infiniOrderCreatePath + "?order_id=" + url.QueryEscape(orderID)
	var order infiniOrderStatusResponse
	if err := i.doJSON(ctx, http.MethodGet, path, nil, &order); err != nil {
		return nil, fmt.Errorf("infini query order: %w", err)
	}
	amount := infiniSettledAmount(order.Amount, order.AmountConfirmed)
	status := infiniProviderStatus(order.PayStatus, order.ExceptionTags)
	if strings.TrimSpace(order.PayStatus) == "" {
		status = infiniProviderStatus(order.Status, order.ExceptionTags)
	}
	if status != payment.ProviderStatusPaid {
		if fallbackStatus := infiniProviderStatus(order.Status, order.ExceptionTags); fallbackStatus == payment.ProviderStatusPaid {
			status = fallbackStatus
		}
	}
	return &payment.QueryOrderResponse{
		TradeNo:  order.OrderID,
		Status:   status,
		Amount:   amount,
		Metadata: infiniOrderMetadata(order.OrderID, order.Currency, order.ClientReference, ""),
	}, nil
}

func (i *Infini) VerifyNotification(_ context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	eventID, err := verifyInfiniWebhookSignature(rawBody, headers, i.config["webhookSecret"], time.Now())
	if err != nil {
		return nil, err
	}

	event, err := decodeInfiniWebhookPayload([]byte(rawBody))
	if err != nil {
		return nil, fmt.Errorf("infini parse webhook: %w", err)
	}
	if strings.TrimSpace(event.ClientReference) == "" || strings.TrimSpace(event.OrderID) == "" {
		return nil, fmt.Errorf("infini webhook missing order_id or client_reference")
	}

	status := infiniProviderStatus(event.Status, event.ExceptionTags)
	switch strings.TrimSpace(event.Event) {
	case infiniEventOrderCompleted:
		status = payment.ProviderStatusSuccess
	case infiniEventOrderProcessing:
		status = payment.ProviderStatusPending
	case infiniEventOrderExpired, infiniEventOrderLatePayment:
		if status == payment.ProviderStatusPaid {
			status = payment.ProviderStatusSuccess
			break
		}
		status = payment.ProviderStatusFailed
	default:
		return nil, nil
	}
	if status != payment.ProviderStatusSuccess {
		return nil, nil
	}

	amount := infiniSettledAmount(event.Amount, event.AmountConfirmed)
	return &payment.PaymentNotification{
		TradeNo:  event.OrderID,
		OrderID:  event.ClientReference,
		Amount:   amount,
		Status:   payment.NotificationStatusSuccess,
		RawData:  rawBody,
		Metadata: infiniOrderMetadata(event.OrderID, event.Currency, event.ClientReference, eventID),
	}, nil
}

func (i *Infini) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("infini refund is not implemented")
}

func (i *Infini) doJSON(ctx context.Context, method, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, i.config["apiBase"]+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	i.signRequest(req, method, path, body)

	respBody, status, err := i.do(req)
	if err != nil {
		return err
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP %d: %s", status, summarizeInfiniResponse(respBody))
	}
	if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
		return nil
	}
	if err := decodeInfiniJSONResponse(respBody, out); err != nil {
		return err
	}
	return nil
}

func decodeInfiniJSONResponse(respBody []byte, out any) error {
	body := bytes.TrimSpace(respBody)
	var envelope struct {
		Code    any             `json:"code"`
		Message string          `json:"message"`
		Msg     string          `json:"msg"`
		Error   json.RawMessage `json:"error"`
		Data    json.RawMessage `json:"data"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil {
		for _, raw := range []json.RawMessage{envelope.Data, envelope.Result} {
			if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				continue
			}
			if err := json.Unmarshal(raw, out); err != nil {
				return fmt.Errorf("parse response data: %w", err)
			}
			return nil
		}
		if len(bytes.TrimSpace(envelope.Error)) > 0 && !bytes.Equal(bytes.TrimSpace(envelope.Error), []byte("null")) {
			return fmt.Errorf("infini error response: %s", summarizeInfiniResponse(body))
		}
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	return nil
}

func decodeInfiniWebhookPayload(rawBody []byte) (infiniWebhookPayload, error) {
	var event infiniWebhookPayload
	body := bytes.TrimSpace(rawBody)
	if len(body) == 0 {
		return event, fmt.Errorf("empty webhook body")
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return event, err
	}
	if strings.TrimSpace(event.OrderID) != "" || strings.TrimSpace(event.ClientReference) != "" {
		return event, nil
	}

	var envelope struct {
		Event     string          `json:"event"`
		EventType string          `json:"event_type"`
		Type      string          `json:"type"`
		Data      json.RawMessage `json:"data"`
		Result    json.RawMessage `json:"result"`
		Object    json.RawMessage `json:"object"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return event, err
	}

	for _, raw := range []json.RawMessage{envelope.Data, envelope.Result, envelope.Object} {
		if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		if err := json.Unmarshal(raw, &event); err != nil {
			return event, err
		}
		if event.Event == "" {
			event.Event = firstNonEmptyString(envelope.Event, envelope.EventType, envelope.Type)
		}
		return event, nil
	}

	return event, nil
}

func (i *Infini) signRequest(req *http.Request, method, path string, body []byte) {
	date := time.Now().UTC().Format(http.TimeFormat)
	keyID := strings.TrimSpace(i.config["keyId"])
	signingString := fmt.Sprintf("%s\n%s %s\ndate: %s\n", keyID, strings.ToUpper(method), path, date)
	mac := hmac.New(sha256.New, []byte(i.config["secretKey"]))
	_, _ = mac.Write([]byte(signingString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("Date", date)
	if len(body) > 0 {
		sum := sha256.Sum256(body)
		req.Header.Set("Digest", "SHA-256="+base64.StdEncoding.EncodeToString(sum[:]))
	}
	req.Header.Set("Authorization", fmt.Sprintf(`Signature keyId="%s",algorithm="hmac-sha256",headers="@request-target date",signature="%s"`, keyID, signature))
}

func (i *Infini) do(req *http.Request) ([]byte, int, error) {
	client := i.httpClient
	if client == nil {
		client = &http.Client{Timeout: infiniHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, infiniMaxResponseSize))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func verifyInfiniWebhookSignature(rawBody string, headers map[string]string, secret string, now time.Time) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", fmt.Errorf("infini webhookSecret not configured")
	}
	signature := strings.TrimSpace(headers["x-webhook-signature"])
	timestamp := strings.TrimSpace(headers["x-webhook-timestamp"])
	eventID := strings.TrimSpace(headers["x-webhook-event-id"])
	if signature == "" || timestamp == "" || eventID == "" {
		return "", fmt.Errorf("infini webhook missing required signature headers")
	}
	ts, err := parseInfiniWebhookTimestamp(timestamp)
	if err != nil {
		return "", err
	}
	if now.Sub(ts) > infiniWebhookTolerance || ts.Sub(now) > infiniWebhookTolerance {
		return "", fmt.Errorf("infini webhook timestamp outside tolerance")
	}
	signedContent := timestamp + "." + eventID + "." + rawBody
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signedContent))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(signature))) {
		return "", fmt.Errorf("infini webhook signature mismatch")
	}
	return eventID, nil
}

func parseInfiniWebhookTimestamp(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("infini webhook timestamp is empty")
	}
	seconds, err := decimal.NewFromString(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("infini webhook timestamp invalid: %w", err)
	}
	return time.Unix(seconds.IntPart(), 0), nil
}

func infiniProviderStatus(status string, tags []string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case infiniStatusPaid, "success", "succeeded", "completed", "confirmed", "overpaid", "fully_paid":
		return payment.ProviderStatusPaid
	case infiniStatusProcessing, infiniStatusPending, "created", "waiting", "unpaid", "confirming":
		return payment.ProviderStatusPending
	case infiniStatusExpired, infiniStatusPartialPaid, "failed", "failure", "cancelled", "canceled":
		for _, tag := range tags {
			if strings.EqualFold(strings.TrimSpace(tag), "late") {
				return payment.ProviderStatusPaid
			}
		}
		return payment.ProviderStatusFailed
	default:
		return payment.ProviderStatusPending
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func infiniAmount(raw any) float64 {
	switch v := raw.(type) {
	case string:
		d, err := decimal.NewFromString(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return d.InexactFloat64()
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		d, err := decimal.NewFromString(v.String())
		if err != nil {
			return 0
		}
		return d.InexactFloat64()
	default:
		return 0
	}
}

func infiniSettledAmount(orderAmount any, amountConfirmed any) float64 {
	if confirmed := infiniAmount(amountConfirmed); confirmed > 0 {
		return confirmed
	}
	return infiniAmount(orderAmount)
}

func infiniOrderMetadata(orderID, currency, clientReference, eventID string) map[string]string {
	metadata := map[string]string{
		"currency":         payment.TypeUSDT,
		"infini_order_id":  strings.TrimSpace(orderID),
		"client_reference": strings.TrimSpace(clientReference),
	}
	if c := strings.TrimSpace(currency); c != "" {
		metadata["infini_currency"] = strings.ToUpper(c)
	}
	if eventID = strings.TrimSpace(eventID); eventID != "" {
		metadata["webhook_event_id"] = eventID
	}
	return metadata
}

func summarizeInfiniResponse(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > infiniMaxErrorSummary {
		return s[:infiniMaxErrorSummary] + "...(truncated)"
	}
	return s
}

type infiniCreateOrderRequest struct {
	Amount          string `json:"amount"`
	Currency        string `json:"currency,omitempty"`
	RequestID       string `json:"request_id"`
	ClientReference string `json:"client_reference,omitempty"`
	OrderDesc       string `json:"order_desc,omitempty"`
	SuccessURL      string `json:"success_url,omitempty"`
	FailureURL      string `json:"failure_url,omitempty"`
	MerchantAlias   string `json:"merchant_alias,omitempty"`
	PayMethods      []int  `json:"pay_methods,omitempty"`
}

type infiniCreateOrderResponse struct {
	OrderID         string `json:"order_id"`
	RequestID       string `json:"request_id"`
	CheckoutURL     string `json:"checkout_url"`
	ClientReference string `json:"client_reference"`
}

type infiniOrderStatusResponse struct {
	OrderID          string   `json:"order_id"`
	Status           string   `json:"status"`
	PayStatus        string   `json:"pay_status"`
	Amount           any      `json:"amount"`
	Currency         string   `json:"currency"`
	AmountConfirming any      `json:"amount_confirming"`
	AmountConfirmed  any      `json:"amount_confirmed"`
	ExceptionTags    []string `json:"exception_tags"`
	ClientReference  string   `json:"client_reference"`
}

type infiniWebhookPayload struct {
	Event            string   `json:"event"`
	OrderID          string   `json:"order_id"`
	ClientReference  string   `json:"client_reference"`
	Amount           any      `json:"amount"`
	Currency         string   `json:"currency"`
	Status           string   `json:"status"`
	AmountConfirming any      `json:"amount_confirming"`
	AmountConfirmed  any      `json:"amount_confirmed"`
	ExceptionTags    []string `json:"exception_tags"`
}
