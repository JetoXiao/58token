package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrTrialCardNotFound      = infraerrors.NotFound("TRIAL_CARD_NOT_FOUND", "trial card not found")
	ErrTrialCardInactive      = infraerrors.BadRequest("TRIAL_CARD_INACTIVE", "trial card is inactive")
	ErrTrialCardExpired       = infraerrors.BadRequest("TRIAL_CARD_EXPIRED", "trial card has expired")
	ErrTrialCardExhausted     = infraerrors.BadRequest("TRIAL_CARD_EXHAUSTED", "trial card has no remaining redemptions")
	ErrTrialCardUserLimit     = infraerrors.BadRequest("TRIAL_CARD_USER_LIMIT", "trial card redemption limit reached")
	ErrFreeQuotaGroupsMissing = infraerrors.BadRequest("FREE_QUOTA_GROUPS_MISSING", "free quota allowed groups are required")
)

const (
	FreeQuotaSourceInvitation = "invitation"
	FreeQuotaSourceTrialCard  = "trial_card"
)

type FreeQuotaLedger struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	SourceType      string     `json:"source_type"`
	SourceID        string     `json:"source_id,omitempty"`
	Amount          float64    `json:"amount"`
	RemainingAmount float64    `json:"remaining_amount"`
	AllowedGroupIDs []int64    `json:"allowed_group_ids"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TrialCard struct {
	ID             int64      `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Amount         float64    `json:"amount"`
	MaxRedemptions int        `json:"max_redemptions"`
	RedeemedCount  int        `json:"redeemed_count"`
	PerUserLimit   int        `json:"per_user_limit"`
	Status         string     `json:"status"`
	Notes          string     `json:"notes,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type FreeQuotaSettings struct {
	InvitationEnabled bool    `json:"invitation_enabled"`
	InvitationAmount  float64 `json:"invitation_amount"`
	GroupIDs          []int64 `json:"group_ids"`
	ShowLockedGroups  bool    `json:"show_locked_groups"`
	TransferOnPayment bool    `json:"transfer_on_payment"`
}

type FreeQuotaSummary struct {
	BalanceAmount   float64 `json:"balance_amount"`
	FreeQuotaAmount float64 `json:"free_quota_amount"`
	TotalAmount     float64 `json:"total_amount"`
}

type CreateTrialCardInput struct {
	Code           string
	Name           string
	Amount         float64
	MaxRedemptions int
	PerUserLimit   int
	Status         string
	Notes          string
	ExpiresAt      *time.Time
}

type UpdateTrialCardInput struct {
	Name           *string
	Amount         *float64
	MaxRedemptions *int
	PerUserLimit   *int
	Status         *string
	Notes          *string
	ExpiresAt      **time.Time
}

type FreeQuotaRepository interface {
	Grant(ctx context.Context, grant *FreeQuotaLedger) error
	GetAvailableForGroup(ctx context.Context, userID, groupID int64) (float64, error)
	RedeemTrialCard(ctx context.Context, userID int64, code string, allowedGroupIDs []int64) (*TrialCard, *FreeQuotaLedger, error)
	ListTrialCards(ctx context.Context, page, pageSize int) ([]TrialCard, int64, error)
	CreateTrialCard(ctx context.Context, input *CreateTrialCardInput) (*TrialCard, error)
	UpdateTrialCard(ctx context.Context, id int64, input *UpdateTrialCardInput) (*TrialCard, error)
	DeleteTrialCard(ctx context.Context, id int64) error
	MarkPaymentSucceeded(ctx context.Context, userID int64, amount float64, transferFreeQuota bool) error
	GetUserBalance(ctx context.Context, userID int64) (float64, error)
	GetUserFreeQuotaSummary(ctx context.Context, userID int64) (*FreeQuotaSummary, error)
}

type FreeQuotaService struct {
	repo           FreeQuotaRepository
	settingService *SettingService
}

func NewFreeQuotaService(repo FreeQuotaRepository, settingService *SettingService) *FreeQuotaService {
	return &FreeQuotaService{repo: repo, settingService: settingService}
}

func (s *FreeQuotaService) GrantInvitationQuota(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil || userID <= 0 {
		return nil
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.InvitationEnabled || settings.InvitationAmount <= 0 {
		return nil
	}
	if len(settings.GroupIDs) == 0 {
		return ErrFreeQuotaGroupsMissing
	}
	return s.repo.Grant(ctx, &FreeQuotaLedger{
		UserID:          userID,
		SourceType:      FreeQuotaSourceInvitation,
		Amount:          settings.InvitationAmount,
		RemainingAmount: settings.InvitationAmount,
		AllowedGroupIDs: settings.GroupIDs,
		Status:          StatusActive,
	})
}

func (s *FreeQuotaService) GetAvailableForGroup(ctx context.Context, userID, groupID int64) (float64, error) {
	if s == nil || s.repo == nil || userID <= 0 || groupID <= 0 {
		return 0, nil
	}
	return s.repo.GetAvailableForGroup(ctx, userID, groupID)
}

func (s *FreeQuotaService) GetUserFreeQuotaSummary(ctx context.Context, userID int64) (*FreeQuotaSummary, error) {
	if s == nil || s.repo == nil || userID <= 0 {
		return &FreeQuotaSummary{}, nil
	}
	return s.repo.GetUserFreeQuotaSummary(ctx, userID)
}

func (s *FreeQuotaService) RedeemTrialCard(ctx context.Context, userID int64, code string) (*TrialCard, *FreeQuotaLedger, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrTrialCardNotFound
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, nil, err
	}
	if len(settings.GroupIDs) == 0 {
		return nil, nil, ErrFreeQuotaGroupsMissing
	}
	return s.repo.RedeemTrialCard(ctx, userID, strings.TrimSpace(code), settings.GroupIDs)
}

func (s *FreeQuotaService) ListTrialCards(ctx context.Context, page, pageSize int) ([]TrialCard, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, nil
	}
	return s.repo.ListTrialCards(ctx, page, pageSize)
}

func (s *FreeQuotaService) CreateTrialCard(ctx context.Context, input *CreateTrialCardInput) (*TrialCard, error) {
	if s == nil || s.repo == nil {
		return nil, ErrTrialCardNotFound
	}
	return s.repo.CreateTrialCard(ctx, input)
}

func (s *FreeQuotaService) UpdateTrialCard(ctx context.Context, id int64, input *UpdateTrialCardInput) (*TrialCard, error) {
	if s == nil || s.repo == nil {
		return nil, ErrTrialCardNotFound
	}
	return s.repo.UpdateTrialCard(ctx, id, input)
}

func (s *FreeQuotaService) DeleteTrialCard(ctx context.Context, id int64) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.DeleteTrialCard(ctx, id)
}

func (s *FreeQuotaService) MarkPaymentSucceeded(ctx context.Context, userID int64, amount float64) error {
	if s == nil || s.repo == nil || userID <= 0 {
		return nil
	}
	transfer := false
	if settings, err := s.GetSettings(ctx); err == nil {
		transfer = settings.TransferOnPayment
	}
	return s.repo.MarkPaymentSucceeded(ctx, userID, amount, transfer)
}

func (s *FreeQuotaService) HasBillableCredit(ctx context.Context, userID int64, balance float64, groupID int64) bool {
	if balance > 0 {
		return true
	}
	available, err := s.GetAvailableForGroup(ctx, userID, groupID)
	return err == nil && available > 0
}

func (s *FreeQuotaService) CanUnpaidUserBindGroup(ctx context.Context, user *User, groupID int64, isExclusive bool) bool {
	if user == nil {
		return false
	}
	if user.Balance > 0 {
		return user.CanBindGroup(groupID, isExclusive)
	}
	settings, err := s.GetSettings(ctx)
	if err != nil || len(settings.GroupIDs) == 0 {
		return false
	}
	return containsInt64(settings.GroupIDs, groupID)
}

func (s *FreeQuotaService) GetSettings(ctx context.Context) (*FreeQuotaSettings, error) {
	settings := &FreeQuotaSettings{}
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil {
		return settings, nil
	}
	values, err := s.settingService.settingRepo.GetMultiple(ctx, []string{
		SettingKeyFreeQuotaInvitationEnabled,
		SettingKeyFreeQuotaInvitationAmount,
		SettingKeyFreeQuotaGroupIDs,
		SettingKeyFreeQuotaShowLockedGroups,
		SettingKeyFreeQuotaTransferOnPayment,
	})
	if err != nil {
		return nil, fmt.Errorf("get free quota settings: %w", err)
	}
	settings.InvitationEnabled = values[SettingKeyFreeQuotaInvitationEnabled] == "true"
	settings.InvitationAmount, _ = strconv.ParseFloat(strings.TrimSpace(values[SettingKeyFreeQuotaInvitationAmount]), 64)
	settings.GroupIDs = parseInt64ListSetting(values[SettingKeyFreeQuotaGroupIDs])
	settings.ShowLockedGroups = values[SettingKeyFreeQuotaShowLockedGroups] == "true"
	settings.TransferOnPayment = values[SettingKeyFreeQuotaTransferOnPayment] == "true"
	return settings, nil
}

func (s *FreeQuotaService) UpdateSettings(ctx context.Context, settings *FreeQuotaSettings) error {
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil || settings == nil {
		return nil
	}
	rawGroups, err := json.Marshal(settings.GroupIDs)
	if err != nil {
		return fmt.Errorf("marshal free quota groups: %w", err)
	}
	updates := map[string]string{
		SettingKeyFreeQuotaInvitationEnabled: strconv.FormatBool(settings.InvitationEnabled),
		SettingKeyFreeQuotaInvitationAmount:  strconv.FormatFloat(settings.InvitationAmount, 'f', -1, 64),
		SettingKeyFreeQuotaGroupIDs:          string(rawGroups),
		SettingKeyFreeQuotaShowLockedGroups:  strconv.FormatBool(settings.ShowLockedGroups),
		SettingKeyFreeQuotaTransferOnPayment: strconv.FormatBool(settings.TransferOnPayment),
	}
	return s.settingService.settingRepo.SetMultiple(ctx, updates)
}

func parseInt64ListSetting(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var ids []int64
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &ids); err == nil {
			return dedupePositiveInt64(ids)
		}
	}
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return dedupePositiveInt64(ids)
}

func dedupePositiveInt64(in []int64) []int64 {
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, id := range in {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
