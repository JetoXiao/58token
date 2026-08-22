//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRuntimeBreakerPolicyMatchesObservedTrafficTiers(t *testing.T) {
	high := runtimeBreakerPolicyFor(PlatformAnthropic, "claude-opus-4-8")
	quiet := runtimeBreakerPolicyFor(PlatformAnthropic, "claude-haiku-4-5")
	openAI := runtimeBreakerPolicyFor(PlatformOpenAI, "gpt-5.6-sol")
	require.Equal(t, 2*time.Minute, high.Window)
	require.Equal(t, 2*time.Minute, openAI.Cooldown)
	require.Equal(t, 15*time.Minute, quiet.Window)
	require.Equal(t, 10*time.Minute, quiet.Cooldown)
}
