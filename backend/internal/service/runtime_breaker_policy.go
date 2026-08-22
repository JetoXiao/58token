package service

import (
	"strings"
	"time"
)

type runtimeBreakerPolicy struct {
	Threshold int
	Window    time.Duration
	Cooldown  time.Duration
}

// Policies are based on production traffic: Opus/Sonnet and OpenAI are high
// volume; long-tail Claude models need a longer observation window.
func runtimeBreakerPolicyFor(platform, model string) runtimeBreakerPolicy {
	model = strings.ToLower(strings.TrimSpace(model))
	if platform == PlatformOpenAI {
		return runtimeBreakerPolicy{3, 2 * time.Minute, 2 * time.Minute}
	}
	if platform == PlatformAnthropic {
		switch model {
		case "claude-opus-4-8", "claude-sonnet-5":
			return runtimeBreakerPolicy{3, 2 * time.Minute, 2 * time.Minute}
		case "claude-fable-5":
			return runtimeBreakerPolicy{3, 5 * time.Minute, 5 * time.Minute}
		default:
			return runtimeBreakerPolicy{3, 15 * time.Minute, 10 * time.Minute}
		}
	}
	return runtimeBreakerPolicy{3, 5 * time.Minute, 5 * time.Minute}
}
