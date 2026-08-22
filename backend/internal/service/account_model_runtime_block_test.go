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

func TestGenericAccountModelBreakerIsolatesModelAndRecovers(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{ID: 801, Platform: PlatformGemini, Status: StatusActive, Schedulable: true}
	err := &UpstreamFailoverError{StatusCode: http.StatusBadGateway}
	ctx := func(id string) context.Context {
		return context.WithValue(context.Background(), ctxkey.ClientRequestID, id)
	}

	require.False(t, svc.RecordAccountFailoverForModel(ctx("a"), account, "gemini-2.5-pro", err))
	require.False(t, svc.RecordAccountFailoverForModel(ctx("b"), account, "gemini-2.5-pro", err))
	require.True(t, svc.RecordAccountFailoverForModel(ctx("c"), account, "gemini-2.5-pro", err))
	require.True(t, svc.isGenericAccountModelRuntimeBlocked(account, "gemini-2.5-pro"))
	require.False(t, svc.isGenericAccountModelRuntimeBlocked(account, "gemini-2.5-flash"))

	key := accountModelRuntimeKey{account.ID, "gemini-2.5-pro"}
	svc.accountModelRuntime.blockUntil.Store(key, time.Now().Add(-time.Second))
	require.False(t, svc.isGenericAccountModelRuntimeBlocked(account, "gemini-2.5-pro"))
	svc.ReportAccountModelScheduleResult(account.ID, account.Platform, "gemini-2.5-pro", true)
	require.False(t, svc.isGenericAccountModelRuntimeBlocked(account, "gemini-2.5-pro"))
}
