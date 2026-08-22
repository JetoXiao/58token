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

const ()

type accountModelRuntimeKey struct {
	accountID int64
	model     string
}
type accountModelFailureState struct {
	mu           sync.Mutex
	requestTimes map[string]time.Time
}
type genericAccountModelRuntimeState struct {
	blockUntil   sync.Map
	failureState sync.Map
}

func (s *GatewayService) isGenericAccountModelRuntimeBlocked(account *Account, requestedModel string) bool {
	return s.isAccountModelRuntimeBlocked(account, requestedModel)
}

func normalizeAccountModelRuntimeModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

func accountModelFailureRequestKey(ctx context.Context, now time.Time) string {
	if ctx != nil {
		if id, _ := ctx.Value(ctxkey.ClientRequestID).(string); strings.TrimSpace(id) != "" {
			return "client:" + strings.TrimSpace(id)
		}
		if id, _ := ctx.Value(ctxkey.RequestID).(string); strings.TrimSpace(id) != "" {
			return "local:" + strings.TrimSpace(id)
		}
	}
	return fmt.Sprintf("anonymous:%d", now.UnixNano())
}

func isAccountModelBreakerStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	default:
		return statusCode >= 500 && statusCode <= 599
	}
}

// RecordAccountFailoverForModel is the generic account+model breaker for
// GatewayService platforms without a dedicated runtime breaker.
func (s *GatewayService) RecordAccountFailoverForModel(ctx context.Context, account *Account, requestedModel string, failoverErr *UpstreamFailoverError) bool {
	model := normalizeAccountModelRuntimeModel(requestedModel)
	if s == nil || account == nil || account.ID <= 0 || model == "" || account.Platform == PlatformAnthropic || account.Platform == PlatformOpenAI || failoverErr == nil || (failoverErr.RequestScoped && !failoverErr.ModelScoped) || !isAccountModelBreakerStatus(failoverErr.StatusCode) {
		return false
	}
	if s.isAccountModelRuntimeBlocked(account, model) {
		return true
	}
	now := time.Now()
	policy := runtimeBreakerPolicyFor(account.Platform, model)
	key := accountModelRuntimeKey{account.ID, model}
	value, _ := s.accountModelRuntime.failureState.LoadOrStore(key, &accountModelFailureState{})
	state, _ := value.(*accountModelFailureState)
	if state == nil {
		state = &accountModelFailureState{}
		s.accountModelRuntime.failureState.Store(key, state)
	}
	state.mu.Lock()
	if state.requestTimes == nil {
		state.requestTimes = make(map[string]time.Time, policy.Threshold)
	}
	for k, at := range state.requestTimes {
		if now.Sub(at) > policy.Window {
			delete(state.requestTimes, k)
		}
	}
	state.requestTimes[accountModelFailureRequestKey(ctx, now)] = now
	count := len(state.requestTimes)
	state.mu.Unlock()
	if count < policy.Threshold {
		return false
	}
	until := now.Add(policy.Cooldown)
	s.accountModelRuntime.failureState.Delete(key)
	s.blockAccountModelScheduling(account.ID, model, until)
	if s.runtimeModelProbe != nil {
		s.runtimeModelProbe.Schedule(account.ID, model, policy.Cooldown)
	}
	slog.Warn("account_model_failures_blocked", "account_id", account.ID, "platform", account.Platform, "model", model, "status_code", failoverErr.StatusCode, "distinct_requests", count, "cooldown_until", until.Format(time.RFC3339))
	return true
}

func (s *GatewayService) ReportAccountModelScheduleResult(accountID int64, platform, requestedModel string, success bool) {
	if s == nil || !success || accountID <= 0 || platform == PlatformAnthropic || platform == PlatformOpenAI {
		return
	}
	model := normalizeAccountModelRuntimeModel(requestedModel)
	if model == "" {
		return
	}
	key := accountModelRuntimeKey{accountID, model}
	s.accountModelRuntime.failureState.Delete(key)
	s.accountModelRuntime.blockUntil.Delete(key)
}

func (s *GatewayService) blockAccountModelScheduling(accountID int64, model string, until time.Time) {
	if s == nil || accountID <= 0 || model == "" {
		return
	}
	key := accountModelRuntimeKey{accountID, model}
	for {
		current, loaded := s.accountModelRuntime.blockUntil.Load(key)
		if !loaded {
			actual, stored := s.accountModelRuntime.blockUntil.LoadOrStore(key, until)
			if !stored {
				return
			}
			current = actual
		}
		if currentUntil, ok := current.(time.Time); ok && currentUntil.After(until) {
			return
		}
		if s.accountModelRuntime.blockUntil.CompareAndSwap(key, current, until) {
			return
		}
	}
}

func (s *GatewayService) isAccountModelRuntimeBlocked(account *Account, requestedModel string) bool {
	model := normalizeAccountModelRuntimeModel(requestedModel)
	if s == nil || account == nil || account.ID <= 0 || model == "" || account.Platform == PlatformAnthropic || account.Platform == PlatformOpenAI {
		return false
	}
	key := accountModelRuntimeKey{account.ID, model}
	value, ok := s.accountModelRuntime.blockUntil.Load(key)
	if !ok {
		return false
	}
	cooldownUntil, ok := value.(time.Time)
	if !ok || cooldownUntil.IsZero() {
		s.accountModelRuntime.blockUntil.Delete(key)
		return false
	}
	if time.Now().Before(cooldownUntil) {
		return true
	}
	s.accountModelRuntime.blockUntil.Delete(key)
	now := time.Now()
	s.accountModelRuntime.failureState.Store(key, &accountModelFailureState{requestTimes: map[string]time.Time{"half-open:1": now, "half-open:2": now}})
	slog.Info("account_model_half_open_probe", "account_id", account.ID, "platform", account.Platform, "model", model)
	return false
}
