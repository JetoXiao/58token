//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAI429FastPath_MarksOAuthAccountCoolingDown(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	apiKeyAccount := &Account{ID: 43, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	shouldDisable := svc.handleOpenAIAccountUpstreamError(context.Background(), account, http.StatusTooManyRequests, http.Header{}, nil)
	apiKeyShouldDisable := svc.handleOpenAIAccountUpstreamError(context.Background(), apiKeyAccount, http.StatusTooManyRequests, http.Header{}, nil)

	require.False(t, shouldDisable)
	require.False(t, apiKeyShouldDisable)
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(apiKeyAccount))
}

func TestOpenAIRuntimeBlock_AppliesToOpenAIAPIKeyWhenRateLimitServiceStopsScheduling(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 44, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	svc.BlockAccountScheduling(account, time.Time{}, "custom_error_code")

	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIRuntimeBlock_DoesNotApplyToOtherPlatforms(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 45, Platform: PlatformGemini, Type: AccountTypeOAuth}

	svc.BlockAccountScheduling(account, time.Time{}, "custom_error_code")

	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIRuntimeBlocker_IgnoresNonOpenAIFromRateLimitService(t *testing.T) {
	gateway := &OpenAIGatewayService{}
	repo := &rateLimitAccountRepoStub{}
	rateLimitService := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	rateLimitService.SetAccountRuntimeBlocker(gateway)
	account := &Account{ID: 45, Platform: PlatformGemini, Type: AccountTypeOAuth}

	shouldDisable := rateLimitService.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte("forbidden"))

	require.True(t, shouldDisable)
	require.False(t, gateway.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIRuntimeBlock_DoesNotShortenExistingBlock(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 46, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	longUntil := time.Now().Add(10 * time.Minute)

	svc.BlockAccountScheduling(account, longUntil, "oauth_401")
	svc.BlockAccountScheduling(account, time.Time{}, "upstream_disable")

	value, ok := svc.openaiAccountRuntimeBlockUntil.Load(account.ID)
	require.True(t, ok)
	actualUntil, ok := value.(time.Time)
	require.True(t, ok)
	require.WithinDuration(t, longUntil, actualUntil, time.Second)
}

func TestOpenAIRuntimeBlock_ClearAccountSchedulingBlock(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 47, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	svc.BlockAccountScheduling(account, time.Now().Add(time.Minute), "429")
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))

	svc.ClearAccountSchedulingBlock(account.ID)
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIConsecutiveFailureBreaker_BlocksAfterThreshold(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 48, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	failoverErr := &UpstreamFailoverError{StatusCode: http.StatusBadGateway}

	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))

	require.True(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIStreamTransportFailureBreaker_BlocksAfterThreshold(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 481, Platform: PlatformOpenAI, Type: AccountTypeOAuth}

	require.False(t, svc.RecordOpenAIAccountStreamTransportFailure(account))
	require.False(t, svc.RecordOpenAIAccountStreamTransportFailure(account))
	require.True(t, svc.RecordOpenAIAccountStreamTransportFailure(account))
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIConsecutiveFailureBreaker_SuccessClearsFailures(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 49, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	failoverErr := &UpstreamFailoverError{StatusCode: http.StatusBadGateway}

	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))

	svc.ReportOpenAIAccountScheduleResult(account.ID, true, nil)

	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestOpenAIConsecutiveFailureBreaker_ResetsAfterWindow(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 50, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	failoverErr := &UpstreamFailoverError{StatusCode: http.StatusBadGateway}

	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	value, ok := svc.openaiAccountConsecutiveFailures.Load(account.ID)
	require.True(t, ok)
	state, ok := value.(*openAIAccountConsecutiveFailureState)
	require.True(t, ok)
	state.mu.Lock()
	state.lastFailureAt = time.Now().Add(-openAIAccountFailureBreakerWindow - time.Second)
	state.mu.Unlock()

	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.False(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.True(t, svc.RecordOpenAIAccountFailover(account, failoverErr))
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestShouldStopOpenAIOAuth429Failover_OnlyDuringStorm(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	apiKeyAccount := &Account{ID: 43, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	require.False(t, svc.ShouldStopOpenAIOAuth429Failover(account, http.StatusTooManyRequests, 1))

	for i := 0; i < openAIOAuth429StormThreshold; i++ {
		svc.recordOpenAIOAuth429()
	}

	require.True(t, svc.ShouldStopOpenAIOAuth429Failover(account, http.StatusTooManyRequests, 1))
	require.False(t, svc.ShouldStopOpenAIOAuth429Failover(apiKeyAccount, http.StatusTooManyRequests, 1))
	require.False(t, svc.ShouldStopOpenAIOAuth429Failover(account, http.StatusInternalServerError, 1))
	require.False(t, svc.ShouldStopOpenAIOAuth429Failover(account, http.StatusTooManyRequests, 0))
}
