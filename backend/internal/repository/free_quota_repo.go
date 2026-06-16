package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type freeQuotaRepository struct {
	db *sql.DB
}

func NewFreeQuotaRepository(_ *dbent.Client, sqlDB *sql.DB) service.FreeQuotaRepository {
	return &freeQuotaRepository{db: sqlDB}
}

func (r *freeQuotaRepository) Grant(ctx context.Context, grant *service.FreeQuotaLedger) error {
	if r == nil || r.db == nil || grant == nil || grant.Amount <= 0 {
		return nil
	}
	if len(grant.AllowedGroupIDs) == 0 {
		return service.ErrFreeQuotaGroupsMissing
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_free_quota_ledger (
			user_id, source_type, source_id, amount, remaining_amount, allowed_group_ids, status, notes, expires_at
		) VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, NULLIF($8, ''), $9)
	`, grant.UserID, grant.SourceType, grant.SourceID, grant.Amount, grant.RemainingAmount, pq.Array(grant.AllowedGroupIDs), coalesceStatus(grant.Status), grant.Notes, grant.ExpiresAt)
	return err
}

func (r *freeQuotaRepository) GetAvailableForGroup(ctx context.Context, userID, groupID int64) (float64, error) {
	if r == nil || r.db == nil || userID <= 0 || groupID <= 0 {
		return 0, nil
	}
	var amount float64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(remaining_amount), 0)
		FROM user_free_quota_ledger
		WHERE user_id = $1
			AND status = $3
			AND remaining_amount > 0
			AND $2 = ANY(allowed_group_ids)
			AND (expires_at IS NULL OR expires_at > NOW())
	`, userID, groupID, service.StatusActive).Scan(&amount)
	return amount, err
}

func (r *freeQuotaRepository) GetUserFreeQuotaSummary(ctx context.Context, userID int64) (*service.FreeQuotaSummary, error) {
	if r == nil || r.db == nil || userID <= 0 {
		return &service.FreeQuotaSummary{}, nil
	}
	summary := &service.FreeQuotaSummary{}
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(u.balance, 0),
			COALESCE(SUM(q.remaining_amount) FILTER (
				WHERE q.status = $2
					AND q.remaining_amount > 0
					AND (q.expires_at IS NULL OR q.expires_at > NOW())
			), 0)
		FROM users u
		LEFT JOIN user_free_quota_ledger q ON q.user_id = u.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
		GROUP BY u.id, u.balance
	`, userID, service.StatusActive).Scan(&summary.BalanceAmount, &summary.FreeQuotaAmount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	summary.TotalAmount = summary.BalanceAmount + summary.FreeQuotaAmount
	return summary, nil
}

func (r *freeQuotaRepository) RedeemTrialCard(ctx context.Context, userID int64, code string, allowedGroupIDs []int64) (*service.TrialCard, *service.FreeQuotaLedger, error) {
	if r == nil || r.db == nil {
		return nil, nil, service.ErrTrialCardNotFound
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil, service.ErrTrialCardNotFound
	}
	if len(allowedGroupIDs) == 0 {
		return nil, nil, service.ErrFreeQuotaGroupsMissing
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	card, err := scanTrialCard(tx.QueryRowContext(ctx, `
		SELECT id, code, name, amount, max_redemptions, redeemed_count, per_user_limit,
			status, notes, expires_at, created_at, updated_at
		FROM trial_cards
		WHERE code = $1
		FOR UPDATE
	`, code))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, service.ErrTrialCardNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	if card.Status != service.StatusActive {
		return nil, nil, service.ErrTrialCardInactive
	}
	if card.ExpiresAt != nil && time.Now().After(*card.ExpiresAt) {
		return nil, nil, service.ErrTrialCardExpired
	}
	if card.RedeemedCount >= card.MaxRedemptions {
		return nil, nil, service.ErrTrialCardExhausted
	}

	var userRedeemed int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM trial_card_redemptions
		WHERE trial_card_id = $1 AND user_id = $2
	`, card.ID, userID).Scan(&userRedeemed); err != nil {
		return nil, nil, err
	}
	if userRedeemed >= card.PerUserLimit {
		return nil, nil, service.ErrTrialCardUserLimit
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE trial_cards
		SET redeemed_count = redeemed_count + 1, updated_at = NOW()
		WHERE id = $1 AND redeemed_count < max_redemptions
	`, card.ID)
	if err != nil {
		return nil, nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, nil, err
	}
	if affected == 0 {
		return nil, nil, service.ErrTrialCardExhausted
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO trial_card_redemptions (trial_card_id, user_id, amount)
		VALUES ($1, $2, $3)
	`, card.ID, userID, card.Amount); err != nil {
		return nil, nil, err
	}

	ledger := &service.FreeQuotaLedger{
		UserID:          userID,
		SourceType:      service.FreeQuotaSourceTrialCard,
		SourceID:        fmt.Sprint(card.ID),
		Amount:          card.Amount,
		RemainingAmount: card.Amount,
		AllowedGroupIDs: allowedGroupIDs,
		Status:          service.StatusActive,
	}
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO user_free_quota_ledger (
			user_id, source_type, source_id, amount, remaining_amount, allowed_group_ids, status
		) VALUES ($1, $2, $3, $4, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, userID, ledger.SourceType, ledger.SourceID, ledger.Amount, pq.Array(ledger.AllowedGroupIDs), ledger.Status).
		Scan(&ledger.ID, &ledger.CreatedAt, &ledger.UpdatedAt); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	tx = nil
	card.RedeemedCount++
	return card, ledger, nil
}

func (r *freeQuotaRepository) ListTrialCards(ctx context.Context, page, pageSize int) ([]service.TrialCard, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 20
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trial_cards`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, amount, max_redemptions, redeemed_count, per_user_limit,
			status, notes, expires_at, created_at, updated_at
		FROM trial_cards
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]service.TrialCard, 0)
	for rows.Next() {
		card, err := scanTrialCardRows(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *card)
	}
	return out, total, rows.Err()
}

func (r *freeQuotaRepository) CreateTrialCard(ctx context.Context, input *service.CreateTrialCardInput) (*service.TrialCard, error) {
	if r == nil || r.db == nil || input == nil {
		return nil, service.ErrTrialCardNotFound
	}
	status := coalesceStatus(input.Status)
	perUserLimit := input.PerUserLimit
	if perUserLimit <= 0 {
		perUserLimit = 1
	}
	maxRedemptions := input.MaxRedemptions
	if maxRedemptions <= 0 {
		maxRedemptions = 1
	}
	return scanTrialCard(r.db.QueryRowContext(ctx, `
		INSERT INTO trial_cards (
			code, name, amount, max_redemptions, per_user_limit, status, notes, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8)
		RETURNING id, code, name, amount, max_redemptions, redeemed_count, per_user_limit,
			status, notes, expires_at, created_at, updated_at
	`, strings.TrimSpace(input.Code), input.Name, input.Amount, maxRedemptions, perUserLimit, status, input.Notes, input.ExpiresAt))
}

func (r *freeQuotaRepository) UpdateTrialCard(ctx context.Context, id int64, input *service.UpdateTrialCardInput) (*service.TrialCard, error) {
	if r == nil || r.db == nil || input == nil {
		return nil, service.ErrTrialCardNotFound
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	card, err := scanTrialCard(tx.QueryRowContext(ctx, `
		SELECT id, code, name, amount, max_redemptions, redeemed_count, per_user_limit,
			status, notes, expires_at, created_at, updated_at
		FROM trial_cards WHERE id = $1 FOR UPDATE
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrTrialCardNotFound
	}
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		card.Name = *input.Name
	}
	if input.Amount != nil {
		card.Amount = *input.Amount
	}
	if input.MaxRedemptions != nil {
		card.MaxRedemptions = *input.MaxRedemptions
	}
	if input.PerUserLimit != nil {
		card.PerUserLimit = *input.PerUserLimit
	}
	if input.Status != nil {
		card.Status = *input.Status
	}
	if input.Notes != nil {
		card.Notes = *input.Notes
	}
	if input.ExpiresAt != nil {
		card.ExpiresAt = *input.ExpiresAt
	}
	if card.PerUserLimit <= 0 || card.MaxRedemptions < card.RedeemedCount {
		return nil, service.ErrTrialCardUserLimit
	}
	card, err = scanTrialCard(tx.QueryRowContext(ctx, `
		UPDATE trial_cards
		SET name = $2, amount = $3, max_redemptions = $4, per_user_limit = $5,
			status = $6, notes = NULLIF($7, ''), expires_at = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING id, code, name, amount, max_redemptions, redeemed_count, per_user_limit,
			status, notes, expires_at, created_at, updated_at
	`, id, card.Name, card.Amount, card.MaxRedemptions, card.PerUserLimit, card.Status, card.Notes, card.ExpiresAt))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return card, nil
}

func (r *freeQuotaRepository) DeleteTrialCard(ctx context.Context, id int64) error {
	if r == nil || r.db == nil {
		return nil
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE trial_cards
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, id, "inactive")
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrTrialCardNotFound
	}
	return nil
}

func (r *freeQuotaRepository) MarkPaymentSucceeded(ctx context.Context, userID int64, amount float64, transferFreeQuota bool) error {
	if r == nil || r.db == nil || userID <= 0 {
		return nil
	}
	if !transferFreeQuota {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	var remaining float64
	if err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(remaining_amount), 0)
			FROM user_free_quota_ledger
			WHERE user_id = $1 AND status = $2 AND remaining_amount > 0
	`, userID, service.StatusActive).Scan(&remaining); err != nil {
		return err
	}
	if remaining > 0 {
		if _, err := tx.ExecContext(ctx, `
				UPDATE user_free_quota_ledger
				SET remaining_amount = 0, updated_at = NOW()
				WHERE user_id = $1 AND status = $2 AND remaining_amount > 0
		`, userID, service.StatusActive); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
				UPDATE users
				SET balance = balance + $1, updated_at = NOW()
				WHERE id = $2 AND deleted_at IS NULL
		`, remaining, userID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return service.ErrUserNotFound
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (r *freeQuotaRepository) GetUserBalance(ctx context.Context, userID int64) (float64, error) {
	if r == nil || r.db == nil || userID <= 0 {
		return 0, nil
	}
	var balance float64
	err := r.db.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1 AND deleted_at IS NULL`, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUserNotFound
	}
	return balance, err
}

type trialCardScanner interface {
	Scan(dest ...any) error
}

func scanTrialCard(row trialCardScanner) (*service.TrialCard, error) {
	card := &service.TrialCard{}
	var notes sql.NullString
	err := row.Scan(
		&card.ID,
		&card.Code,
		&card.Name,
		&card.Amount,
		&card.MaxRedemptions,
		&card.RedeemedCount,
		&card.PerUserLimit,
		&card.Status,
		&notes,
		&card.ExpiresAt,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if notes.Valid {
		card.Notes = notes.String
	}
	return card, nil
}

func scanTrialCardRows(rows *sql.Rows) (*service.TrialCard, error) {
	return scanTrialCard(rows)
}

func coalesceStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return service.StatusActive
	}
	return status
}
