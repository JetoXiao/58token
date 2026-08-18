package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

const (
	anthropicAccountModelFailureBreakerThreshold = 3
	anthropicAccountModelFailureBreakerWindow    = 2 * time.Minute
	anthropicAccountModelFailureBreakerCooldown  = 2 * time.Minute
)

type anthropicAccountModelRuntimeKey struct {
	accountID int64
	model     string
}

type anthropicAccountModelFailureState struct {
	mu           sync.Mutex
	requestTimes map[string]time.Time
}

type anthropicAccountModelRuntimeState struct {
	blockUntil   sync.Map // key: anthropicAccountModelRuntimeKey, value: time.Time
	failureState sync.Map // key: anthropicAccountModelRuntimeKey, value: *anthropicAccountModelFailureState
}

func normalizeAnthropicAccountModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

func isAnthropicAccountModelBreakerStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	default:
		return statusCode >= 500 && statusCode <= 599
	}
}

func anthropicModelFailureRequestKey(ctx context.Context) string {
	if ctx != nil {
		if clientRequestID, _ := ctx.Value(ctxkey.ClientRequestID).(string); strings.TrimSpace(clientRequestID) != "" {
			return "client:" + strings.TrimSpace(clientRequestID)
		}
		if requestID, _ := ctx.Value(ctxkey.RequestID).(string); strings.TrimSpace(requestID) != "" {
			return "local:" + strings.TrimSpace(requestID)
		}
	}
	// Requests normally carry one of the IDs above. A stable anonymous key is
	// safer than counting every retry in one request as a distinct failure.
	return "anonymous"
}

// RecordAnthropicAccountFailoverForModel records repeated upstream failures for
// one Anthropic account/model pair. It intentionally does not disable the
// account globally, so other Claude models can continue using the channel.
func (s *GatewayService) RecordAnthropicAccountFailoverForModel(ctx context.Context, account *Account, requestedModel string, failoverErr *UpstreamFailoverError) bool {
	model := normalizeAnthropicAccountModel(requestedModel)
	if s == nil || account == nil || account.Platform != PlatformAnthropic || account.ID <= 0 || model == "" || failoverErr == nil || failoverErr.RequestScoped || !isAnthropicAccountModelBreakerStatus(failoverErr.StatusCode) {
		return false
	}
	if s.isAnthropicAccountModelRuntimeBlocked(account, model) {
		return true
	}

	now := time.Now()
	key := anthropicAccountModelRuntimeKey{accountID: account.ID, model: model}
	value, _ := s.anthropicModelRuntime.failureState.LoadOrStore(key, &anthropicAccountModelFailureState{})
	state, ok := value.(*anthropicAccountModelFailureState)
	if !ok || state == nil {
		state = &anthropicAccountModelFailureState{}
		s.anthropicModelRuntime.failureState.Store(key, state)
	}

	state.mu.Lock()
	if state.requestTimes == nil {
		state.requestTimes = make(map[string]time.Time, anthropicAccountModelFailureBreakerThreshold)
	}
	for requestKey, failedAt := range state.requestTimes {
		if now.Sub(failedAt) > anthropicAccountModelFailureBreakerWindow {
			delete(state.requestTimes, requestKey)
		}
	}
	state.requestTimes[anthropicModelFailureRequestKey(ctx)] = now
	count := len(state.requestTimes)
	state.mu.Unlock()

	if count < anthropicAccountModelFailureBreakerThreshold {
		return false
	}

	until := now.Add(anthropicAccountModelFailureBreakerCooldown)
	s.anthropicModelRuntime.failureState.Delete(key)
	s.blockAnthropicAccountModelScheduling(account.ID, model, until)
	slog.Warn("anthropic_account_model_failures_blocked",
		"account_id", account.ID,
		"model", model,
		"status_code", failoverErr.StatusCode,
		"distinct_requests", count,
		"cooldown_until", until.Format(time.RFC3339),
	)
	return true
}

// ReportAnthropicAccountScheduleResultForModel clears a model breaker after a
// successful request. A request admitted after cooldown is the half-open probe.
func (s *GatewayService) ReportAnthropicAccountScheduleResultForModel(accountID int64, requestedModel string, success bool) {
	if !success {
		return
	}
	model := normalizeAnthropicAccountModel(requestedModel)
	if s == nil || accountID <= 0 || model == "" {
		return
	}
	key := anthropicAccountModelRuntimeKey{accountID: accountID, model: model}
	s.anthropicModelRuntime.failureState.Delete(key)
	s.anthropicModelRuntime.blockUntil.Delete(key)
}

func (s *GatewayService) blockAnthropicAccountModelScheduling(accountID int64, model string, until time.Time) {
	model = normalizeAnthropicAccountModel(model)
	if s == nil || accountID <= 0 || model == "" {
		return
	}
	key := anthropicAccountModelRuntimeKey{accountID: accountID, model: model}
	for {
		current, loaded := s.anthropicModelRuntime.blockUntil.Load(key)
		if !loaded {
			actual, stored := s.anthropicModelRuntime.blockUntil.LoadOrStore(key, until)
			if !stored {
				return
			}
			current = actual
		}
		currentUntil, ok := current.(time.Time)
		if ok && currentUntil.After(until) {
			return
		}
		if s.anthropicModelRuntime.blockUntil.CompareAndSwap(key, current, until) {
			return
		}
	}
}

func (s *GatewayService) primeAnthropicAccountModelHalfOpen(key anthropicAccountModelRuntimeKey, now time.Time) {
	if s == nil {
		return
	}
	state := &anthropicAccountModelFailureState{requestTimes: make(map[string]time.Time, anthropicAccountModelFailureBreakerThreshold)}
	for i := 1; i < anthropicAccountModelFailureBreakerThreshold; i++ {
		state.requestTimes[fmt.Sprintf("half-open:%d", i)] = now
	}
	s.anthropicModelRuntime.failureState.Store(key, state)
}

func (s *GatewayService) isAnthropicAccountModelRuntimeBlocked(account *Account, requestedModel string) bool {
	model := normalizeAnthropicAccountModel(requestedModel)
	if s == nil || account == nil || account.Platform != PlatformAnthropic || account.ID <= 0 || model == "" {
		return false
	}
	key := anthropicAccountModelRuntimeKey{accountID: account.ID, model: model}
	value, ok := s.anthropicModelRuntime.blockUntil.Load(key)
	if !ok {
		return false
	}
	cooldownUntil, ok := value.(time.Time)
	if !ok || cooldownUntil.IsZero() {
		s.anthropicModelRuntime.blockUntil.Delete(key)
		return false
	}
	if time.Now().Before(cooldownUntil) {
		return true
	}
	// Expiry is the automatic half-open probe. A failed probe immediately
	// reopens the circuit; a successful probe clears the state above.
	s.anthropicModelRuntime.blockUntil.Delete(key)
	s.primeAnthropicAccountModelHalfOpen(key, time.Now())
	slog.Info("anthropic_account_model_half_open_probe", "account_id", account.ID, "model", model)
	return false
}
