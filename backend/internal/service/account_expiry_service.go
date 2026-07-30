package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// expiredAccountRuntimeStateCleaner is implemented by repositories that can remove
// persisted runtime blocks once their own deadline has elapsed.
//
// It intentionally stays separate from AccountRepository so existing lightweight
// repository implementations do not need a database-only maintenance method.
type expiredAccountRuntimeStateCleaner interface {
	ClearExpiredRuntimeState(ctx context.Context, now time.Time) ([]int64, error)
}

// AccountExpiryService periodically maintains time-based account state. Besides
// pausing expired accounts, it removes expired runtime blocks so an account that
// has completed a Gemini/OpenAI/Anthropic cooldown is fully restored in the
// database and scheduler cache.
type AccountExpiryService struct {
	accountRepo AccountRepository
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func NewAccountExpiryService(accountRepo AccountRepository, interval time.Duration) *AccountExpiryService {
	return &AccountExpiryService{
		accountRepo: accountRepo,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (s *AccountExpiryService) Start() {
	if s == nil || s.accountRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *AccountExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *AccountExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	now := time.Now()

	updated, err := s.accountRepo.AutoPauseExpiredAccounts(ctx, now)
	if err != nil {
		log.Printf("[AccountExpiry] Auto pause expired accounts failed: %v", err)
	} else if updated > 0 {
		log.Printf("[AccountExpiry] Auto paused %d expired accounts", updated)
	}

	cleaner, ok := s.accountRepo.(expiredAccountRuntimeStateCleaner)
	if !ok {
		return
	}
	recovered, err := cleaner.ClearExpiredRuntimeState(ctx, now)
	if err != nil {
		log.Printf("[AccountExpiry] Clear expired account runtime state failed: %v", err)
		return
	}
	if len(recovered) > 0 {
		log.Printf("[AccountExpiry] Restored %d accounts after runtime cooldown", len(recovered))
	}
}
