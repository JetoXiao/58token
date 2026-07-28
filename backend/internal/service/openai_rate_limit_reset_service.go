package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	httppool "github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/google/uuid"
)

var (
	openAIRateLimitResetCreditsURL       = "https://chatgpt.com/backend-api/wham/rate-limit-reset-credits"
	openAIConsumeRateLimitResetCreditURL = "https://chatgpt.com/backend-api/wham/rate-limit-reset-credits/consume"

	ErrOpenAIRateLimitResetUnsupported   = errors.New("rate-limit reset credits are only available for OpenAI OAuth accounts")
	ErrOpenAIRateLimitResetProxyRequired = errors.New("the account must have a configured proxy before rate-limit credits can be queried or consumed")
)

type OpenAIRateLimitResetCredit struct {
	ID          string  `json:"id"`
	ResetType   string  `json:"reset_type"`
	Status      string  `json:"status"`
	GrantedAt   string  `json:"granted_at"`
	ExpiresAt   *string `json:"expires_at"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type OpenAIRateLimitResetCredits struct {
	Credits        []OpenAIRateLimitResetCredit `json:"credits"`
	AvailableCount int64                        `json:"available_count"`
}

type ConsumeOpenAIRateLimitResetCreditResult struct {
	Code         string `json:"code"`
	WindowsReset int64  `json:"windows_reset"`
}

type consumeOpenAIRateLimitResetCreditRequest struct {
	RedeemRequestID string `json:"redeem_request_id"`
}

func (s *AccountUsageService) GetOpenAIRateLimitResetCredits(ctx context.Context, accountID int64) (*OpenAIRateLimitResetCredits, error) {
	account, err := s.getOpenAIResetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var result OpenAIRateLimitResetCredits
	if err := s.doOpenAIResetCreditRequest(ctx, account, http.MethodGet, openAIRateLimitResetCreditsURL, nil, &result); err != nil {
		return nil, err
	}
	if result.Credits == nil {
		result.Credits = []OpenAIRateLimitResetCredit{}
	}
	return &result, nil
}

func (s *AccountUsageService) ConsumeOpenAIRateLimitResetCredit(ctx context.Context, accountID int64) (*ConsumeOpenAIRateLimitResetCreditResult, error) {
	account, err := s.getOpenAIResetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(consumeOpenAIRateLimitResetCreditRequest{RedeemRequestID: uuid.NewString()})
	if err != nil {
		return nil, fmt.Errorf("marshal rate-limit reset request: %w", err)
	}

	var result ConsumeOpenAIRateLimitResetCreditResult
	if err := s.doOpenAIResetCreditRequest(ctx, account, http.MethodPost, openAIConsumeRateLimitResetCreditURL, payload, &result); err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.openAIProbeCache.Delete(accountID)
	}
	return &result, nil
}

func (s *AccountUsageService) getOpenAIResetAccount(ctx context.Context, accountID int64) (*Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account failed: %w", err)
	}
	if account == nil || !account.IsOpenAIOAuth() {
		return nil, ErrOpenAIRateLimitResetUnsupported
	}
	if account.ProxyID == nil || account.Proxy == nil || strings.TrimSpace(account.Proxy.URL()) == "" {
		return nil, ErrOpenAIRateLimitResetProxyRequired
	}
	if strings.TrimSpace(account.GetOpenAIAccessToken()) == "" {
		return nil, errors.New("OpenAI OAuth access token is unavailable")
	}
	return account, nil
}

func (s *AccountUsageService) doOpenAIResetCreditRequest(
	ctx context.Context,
	account *Account,
	method string,
	endpoint string,
	payload []byte,
	destination any,
) error {
	requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, method, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create OpenAI rate-limit reset request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+account.GetOpenAIAccessToken())
	request.Header.Set("User-Agent", codexCLIUserAgent)
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	if chatGPTAccountID := strings.TrimSpace(account.GetChatGPTAccountID()); chatGPTAccountID != "" {
		request.Header.Set("ChatGPT-Account-Id", chatGPTAccountID)
	}
	if s.identityCache != nil {
		if fingerprint, fingerprintErr := s.identityCache.GetFingerprint(requestCtx, account.ID); fingerprintErr == nil && fingerprint != nil && strings.TrimSpace(fingerprint.UserAgent) != "" {
			request.Header.Set("User-Agent", strings.TrimSpace(fingerprint.UserAgent))
		}
	}

	// This operation intentionally has no direct-connect fallback: both the query and
	// the destructive consume request must use the proxy assigned to this account.
	client, err := httppool.GetClient(httppool.Options{
		ProxyURL:              account.Proxy.URL(),
		Timeout:               15 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("build proxied OpenAI rate-limit reset client: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("proxied OpenAI rate-limit reset request failed: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read OpenAI rate-limit reset response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("OpenAI rate-limit reset returned status %d", response.StatusCode)
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return fmt.Errorf("decode OpenAI rate-limit reset response: %w", err)
	}
	return nil
}
