package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RuntimeModelProbeService actively validates quarantined account/model pairs
// without waiting for user traffic. It uses the existing account test service,
// keeps one probe per pair, and backs off after repeated failures.
type RuntimeModelProbeService struct {
	accountRepo AccountRepository
	accountTest *AccountTestService
	rateLimit   *RateLimitService
	gateway     *GatewayService
	openAI      *OpenAIGatewayService
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
	mu          sync.Mutex
	next        map[accountModelRuntimeProbeKey]time.Time
	failures    map[accountModelRuntimeProbeKey]int
}

func (s *RuntimeModelProbeService) SetGatewayServices(gateway *GatewayService, openAI *OpenAIGatewayService) {
	if s == nil {
		return
	}
	s.gateway = gateway
	s.openAI = openAI
}

type accountModelRuntimeProbeKey struct {
	accountID int64
	model     string
}

func NewRuntimeModelProbeService(repo AccountRepository, accountTest *AccountTestService, rateLimit *RateLimitService, interval time.Duration) *RuntimeModelProbeService {
	if interval <= 0 {
		interval = time.Minute
	}
	return &RuntimeModelProbeService{accountRepo: repo, accountTest: accountTest, rateLimit: rateLimit, interval: interval, stopCh: make(chan struct{}), next: make(map[accountModelRuntimeProbeKey]time.Time), failures: make(map[accountModelRuntimeProbeKey]int)}
}

func (s *RuntimeModelProbeService) Start() {
	if s == nil || s.accountRepo == nil || s.accountTest == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		t := time.NewTicker(s.interval)
		defer t.Stop()
		s.runOnce()
		for {
			select {
			case <-t.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *RuntimeModelProbeService) Schedule(accountID int64, model string, delay time.Duration) {
	if s == nil || accountID <= 0 || model == "" {
		return
	}
	if delay <= 0 {
		delay = time.Minute
	}
	key := accountModelRuntimeProbeKey{accountID, model}
	s.mu.Lock()
	if current, ok := s.next[key]; !ok || time.Now().Add(delay).Before(current) {
		s.next[key] = time.Now().Add(delay)
	}
	s.mu.Unlock()
}

func (s *RuntimeModelProbeService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}

func (s *RuntimeModelProbeService) runOnce() {
	now := time.Now()
	// Dedicated breakers expose their in-memory keys; persisted account state is
	// handled by AccountExpiryService. This service only probes known active
	// accounts and pairs that have been scheduled for a retry.
	s.mu.Lock()
	due := make([]accountModelRuntimeProbeKey, 0)
	for key, at := range s.next {
		if !now.Before(at) {
			due = append(due, key)
		}
	}
	s.mu.Unlock()
	for _, key := range due {
		go s.runProbe(key)
	}
}

// RunOnceForTest exposes one scan for unit tests without starting a goroutine.
func (s *RuntimeModelProbeService) RunOnceForTest() { s.runOnce() }

func (s *RuntimeModelProbeService) runProbe(key accountModelRuntimeProbeKey) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	account, err := s.accountRepo.GetByID(ctx, key.accountID)
	if err != nil || account == nil || !account.IsActive() {
		s.clear(key)
		return
	}
	result, err := s.accountTest.RunTestBackground(ctx, key.accountID, key.model)
	if err == nil && result != nil && result.Status == "success" {
		if s.rateLimit != nil {
			_, _ = s.rateLimit.RecoverAccountAfterSuccessfulTest(ctx, key.accountID)
		}
		if account.Platform == PlatformAnthropic && s.gateway != nil {
			s.gateway.ReportAnthropicAccountScheduleResultForModel(key.accountID, key.model, true)
		} else if account.Platform == PlatformOpenAI && s.openAI != nil {
			s.openAI.ReportOpenAIAccountScheduleResultForModel(key.accountID, key.model, true, nil)
		} else if s.gateway != nil {
			s.gateway.ReportAccountModelScheduleResult(key.accountID, account.Platform, key.model, true)
		}
		s.clear(key)
		slog.Info("account_model_probe_recovered", "account_id", key.accountID, "platform", account.Platform, "model", key.model)
		return
	}
	s.mu.Lock()
	n := s.failures[key] + 1
	s.failures[key] = n
	delay := 5 * time.Minute
	for i := 1; i < n && delay < time.Hour; i++ {
		delay *= 2
	}
	s.next[key] = time.Now().Add(delay)
	s.mu.Unlock()
	slog.Warn("account_model_probe_failed", "account_id", key.accountID, "platform", account.Platform, "model", key.model, "next_probe_in", delay, "error", err)
}

func (s *RuntimeModelProbeService) clear(key accountModelRuntimeProbeKey) {
	s.mu.Lock()
	delete(s.next, key)
	delete(s.failures, key)
	s.mu.Unlock()
}
