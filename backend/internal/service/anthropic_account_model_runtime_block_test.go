//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestAnthropicModelFailureBreaker_IsolatesAccountAndModel(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{ID: 701, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true}
	err := &UpstreamFailoverError{StatusCode: http.StatusForbidden}
	requestContext := func(id string) context.Context {
		return context.WithValue(context.Background(), ctxkey.ClientRequestID, id)
	}

	require.False(t, svc.RecordAnthropicAccountFailoverForModel(requestContext("r-1"), account, "claude-sonnet-5", err))
	require.False(t, svc.RecordAnthropicAccountFailoverForModel(requestContext("r-2"), account, "claude-sonnet-5", err))
	require.True(t, svc.RecordAnthropicAccountFailoverForModel(requestContext("r-3"), account, "claude-sonnet-5", err))

	require.True(t, svc.isAnthropicAccountModelRuntimeBlocked(account, "claude-sonnet-5"))
	require.False(t, svc.isAnthropicAccountModelRuntimeBlocked(account, "claude-opus-4-8"))
	require.True(t, svc.isAccountSchedulableForModelSelection(context.Background(), account, "claude-opus-4-8"))
	require.False(t, svc.isAccountSchedulableForModelSelection(context.Background(), account, "claude-sonnet-5"))
}

func TestAnthropicModelFailureBreaker_HalfOpenProbeRecovers(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{ID: 702, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true}
	model := "claude-sonnet-5"
	key := anthropicAccountModelRuntimeKey{accountID: account.ID, model: model}
	svc.anthropicModelRuntime.blockUntil.Store(key, time.Now().Add(-time.Second))

	require.False(t, svc.isAnthropicAccountModelRuntimeBlocked(account, model))
	require.True(t, svc.RecordAnthropicAccountFailoverForModel(context.WithValue(context.Background(), ctxkey.RequestID, "probe"), account, model, &UpstreamFailoverError{StatusCode: http.StatusBadGateway}))
	require.True(t, svc.isAnthropicAccountModelRuntimeBlocked(account, model))

	svc.ReportAnthropicAccountScheduleResultForModel(account.ID, model, true)
	require.False(t, svc.isAnthropicAccountModelRuntimeBlocked(account, model))
}

func TestAnthropicModelFailureBreaker_IgnoresRequestScopedErrors(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{ID: 703, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true}
	err := &UpstreamFailoverError{StatusCode: http.StatusBadRequest, RequestScoped: true}

	for i := 0; i < 5; i++ {
		require.False(t, svc.RecordAnthropicAccountFailoverForModel(context.Background(), account, "claude-sonnet-5", err))
	}
	require.False(t, svc.isAnthropicAccountModelRuntimeBlocked(account, "claude-sonnet-5"))
}
