//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type runtimeStateRecoveryRepoStub struct {
	mockAccountRepoForGemini
	autoPauseCalls int
	recoveryCalls  int
	recovered      []int64
	recoveryErr    error
}

func (r *runtimeStateRecoveryRepoStub) AutoPauseExpiredAccounts(context.Context, time.Time) (int64, error) {
	r.autoPauseCalls++
	return 0, nil
}

func (r *runtimeStateRecoveryRepoStub) ClearExpiredRuntimeState(context.Context, time.Time) ([]int64, error) {
	r.recoveryCalls++
	return r.recovered, r.recoveryErr
}

func TestAccountExpiryService_RunOnceRecoversExpiredRuntimeState(t *testing.T) {
	repo := &runtimeStateRecoveryRepoStub{recovered: []int64{58, 67}}
	svc := NewAccountExpiryService(repo, time.Minute)

	svc.runOnce()

	require.Equal(t, 1, repo.autoPauseCalls)
	require.Equal(t, 1, repo.recoveryCalls)
}

func TestAccountExpiryService_RunOnceKeepsRunningWhenRecoveryFails(t *testing.T) {
	repo := &runtimeStateRecoveryRepoStub{recoveryErr: errors.New("database unavailable")}
	svc := NewAccountExpiryService(repo, time.Minute)

	require.NotPanics(t, svc.runOnce)
	require.Equal(t, 1, repo.autoPauseCalls)
	require.Equal(t, 1, repo.recoveryCalls)
}
