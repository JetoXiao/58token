package service

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	openAIAccountStateUpdateTimeout       = 5 * time.Second
	openAIOAuth429FallbackCooldown        = 5 * time.Second
	openAIStopSchedulingBridgeCooldown    = 2 * time.Minute
	openAIAccountFailureBreakerThreshold  = 3
	openAIAccountFailureBreakerWindow     = 2 * time.Minute
	openAIAccountFailureBreakerCooldown   = openAIStopSchedulingBridgeCooldown
	openAIOAuth429StormWindow             = 10 * time.Second
	openAIOAuth429StormThreshold          = 20
	openAIOAuth429StormMaxAccountSwitches = 1
)

type openAIAccountConsecutiveFailureState struct {
	mu             sync.Mutex
	count          int
	firstFailureAt time.Time
	lastFailureAt  time.Time
	lastStatusCode int
	lastReason     string
}

func openAIAccountStateContext(ctx context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if ctx != nil {
		base = context.WithoutCancel(ctx)
	}
	return context.WithTimeout(base, openAIAccountStateUpdateTimeout)
}

func isOpenAIOAuthAccount(account *Account) bool {
	return account != nil && account.Platform == PlatformOpenAI && account.Type == AccountTypeOAuth
}

func isOpenAIAccount(account *Account) bool {
	return account != nil && account.Platform == PlatformOpenAI
}

func (s *OpenAIGatewayService) handleOpenAIAccountUpstreamError(ctx context.Context, account *Account, statusCode int, headers http.Header, responseBody []byte) bool {
	stateCtx, cancel := openAIAccountStateContext(ctx)
	defer cancel()

	if s == nil || account == nil {
		return false
	}
	if statusCode == http.StatusTooManyRequests {
		s.markOpenAIOAuth429RateLimited(stateCtx, account, headers, responseBody)
	}
	if s.rateLimitService == nil {
		return false
	}
	shouldDisable := s.rateLimitService.HandleUpstreamError(stateCtx, account, statusCode, headers, responseBody)
	if shouldDisable {
		s.BlockAccountScheduling(account, time.Time{}, "upstream_disable")
	}
	return shouldDisable
}

func (s *OpenAIGatewayService) RecordOpenAIAccountFailover(account *Account, failoverErr *UpstreamFailoverError) bool {
	if s == nil || failoverErr == nil {
		return false
	}
	return s.recordOpenAIAccountConsecutiveFailure(account, failoverErr.StatusCode, "failover")
}

// RecordOpenAIAccountStreamTransportFailure counts an upstream SSE connection
// interruption after output began. It is intentionally separate from generic
// forwarding errors so client disconnects and response validation errors do not
// make an account unavailable.
func (s *OpenAIGatewayService) RecordOpenAIAccountStreamTransportFailure(account *Account) bool {
	return s.recordOpenAIAccountConsecutiveFailure(account, http.StatusBadGateway, "stream_transport")
}

func (s *OpenAIGatewayService) recordOpenAIAccountConsecutiveFailure(account *Account, statusCode int, reason string) bool {
	if s == nil || !isOpenAIAccount(account) || account.ID <= 0 {
		return false
	}
	if s.isOpenAIAccountRuntimeBlocked(account) {
		return true
	}

	now := time.Now()
	value, _ := s.openaiAccountConsecutiveFailures.LoadOrStore(account.ID, &openAIAccountConsecutiveFailureState{})
	state, ok := value.(*openAIAccountConsecutiveFailureState)
	if !ok || state == nil {
		state = &openAIAccountConsecutiveFailureState{}
		s.openaiAccountConsecutiveFailures.Store(account.ID, state)
	}

	state.mu.Lock()
	if state.lastFailureAt.IsZero() || now.Sub(state.lastFailureAt) > openAIAccountFailureBreakerWindow {
		state.count = 0
		state.firstFailureAt = now
	}
	state.count++
	state.lastFailureAt = now
	state.lastStatusCode = statusCode
	state.lastReason = reason
	count := state.count
	firstFailureAt := state.firstFailureAt
	shouldBlock := count >= openAIAccountFailureBreakerThreshold
	state.mu.Unlock()

	if !shouldBlock {
		return false
	}

	until := now.Add(openAIAccountFailureBreakerCooldown)
	s.openaiAccountConsecutiveFailures.Delete(account.ID)
	s.BlockAccountScheduling(account, until, "consecutive_failures")
	slog.Warn("openai_account_consecutive_failures_blocked",
		"account_id", account.ID,
		"status_code", statusCode,
		"reason", reason,
		"count", count,
		"first_failure_at", firstFailureAt.Format(time.RFC3339),
		"cooldown_until", until.Format(time.RFC3339),
	)
	return true
}

func (s *OpenAIGatewayService) clearOpenAIAccountConsecutiveFailures(accountID int64) {
	if s == nil || accountID <= 0 {
		return
	}
	s.openaiAccountConsecutiveFailures.Delete(accountID)
}

func (s *OpenAIGatewayService) markOpenAIOAuth429RateLimited(ctx context.Context, account *Account, headers http.Header, responseBody []byte) {
	if s == nil || !isOpenAIOAuthAccount(account) {
		return
	}
	s.recordOpenAIOAuth429()

	cooldownUntil := time.Now().Add(openAIOAuth429FallbackCooldown)
	if s.rateLimitService != nil {
		if resetAt := s.rateLimitService.calculateOpenAI429ResetTime(headers); resetAt != nil && resetAt.After(time.Now()) {
			cooldownUntil = *resetAt
		} else if resetUnix := parseOpenAIRateLimitResetTime(responseBody); resetUnix != nil {
			if resetAt := time.Unix(*resetUnix, 0); resetAt.After(time.Now()) {
				cooldownUntil = resetAt
			}
		} else if cooldown, ok := s.rateLimitService.get429FallbackCooldown(ctx, account); ok && cooldown > 0 {
			cooldownUntil = time.Now().Add(cooldown)
		}
	}
	s.BlockAccountScheduling(account, cooldownUntil, "429")
}

func (s *OpenAIGatewayService) BlockAccountScheduling(account *Account, until time.Time, reason string) {
	if s == nil || !isOpenAIAccount(account) {
		return
	}
	now := time.Now()
	blockUntil := until
	if blockUntil.IsZero() || !blockUntil.After(now) {
		blockUntil = now.Add(openAIStopSchedulingBridgeCooldown)
	}

	for {
		current, loaded := s.openaiAccountRuntimeBlockUntil.Load(account.ID)
		if !loaded {
			actual, stored := s.openaiAccountRuntimeBlockUntil.LoadOrStore(account.ID, blockUntil)
			if !stored {
				return
			}
			current = actual
		}

		currentUntil, ok := current.(time.Time)
		if !ok || currentUntil.IsZero() {
			if s.openaiAccountRuntimeBlockUntil.CompareAndSwap(account.ID, current, blockUntil) {
				return
			}
			continue
		}
		if currentUntil.After(blockUntil) {
			return
		}
		if s.openaiAccountRuntimeBlockUntil.CompareAndSwap(account.ID, current, blockUntil) {
			return
		}
	}
}

func (s *OpenAIGatewayService) ClearAccountSchedulingBlock(accountID int64) {
	if s == nil || accountID <= 0 {
		return
	}
	s.openaiAccountRuntimeBlockUntil.Delete(accountID)
	s.clearOpenAIAccountConsecutiveFailures(accountID)
}

func (s *OpenAIGatewayService) isOpenAIAccountRuntimeBlocked(account *Account) bool {
	if s == nil || !isOpenAIAccount(account) {
		return false
	}
	value, ok := s.openaiAccountRuntimeBlockUntil.Load(account.ID)
	if !ok {
		return false
	}
	cooldownUntil, ok := value.(time.Time)
	if !ok || cooldownUntil.IsZero() {
		s.openaiAccountRuntimeBlockUntil.Delete(account.ID)
		return false
	}
	if time.Now().Before(cooldownUntil) {
		return true
	}
	s.openaiAccountRuntimeBlockUntil.Delete(account.ID)
	return false
}

func (s *OpenAIGatewayService) recordOpenAIOAuth429() {
	if s == nil {
		return
	}
	now := time.Now()
	windowStart := s.openaiOAuth429WindowStartUnixNano.Load()
	if windowStart == 0 || now.Sub(time.Unix(0, windowStart)) >= openAIOAuth429StormWindow {
		if s.openaiOAuth429WindowStartUnixNano.CompareAndSwap(windowStart, now.UnixNano()) {
			s.openaiOAuth429WindowCount.Store(1)
			return
		}
	}
	s.openaiOAuth429WindowCount.Add(1)
}

func (s *OpenAIGatewayService) isOpenAIOAuth429Storm() bool {
	if s == nil {
		return false
	}
	windowStart := s.openaiOAuth429WindowStartUnixNano.Load()
	if windowStart == 0 || time.Since(time.Unix(0, windowStart)) >= openAIOAuth429StormWindow {
		return false
	}
	return s.openaiOAuth429WindowCount.Load() >= openAIOAuth429StormThreshold
}

func (s *OpenAIGatewayService) ShouldStopOpenAIOAuth429Failover(account *Account, statusCode int, failedSwitches int) bool {
	if statusCode != http.StatusTooManyRequests || failedSwitches < openAIOAuth429StormMaxAccountSwitches {
		return false
	}
	if !isOpenAIOAuthAccount(account) {
		return false
	}
	return s.isOpenAIOAuth429Storm()
}
