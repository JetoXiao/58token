package service

import (
	"encoding/json"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

const paymentOrderSnapshotSubscriptionTargetKey = "subscription_target"

// PaymentOrderSubscriptionTarget is a non-sensitive snapshot of the user who
// receives a subscription purchased by another account.
type PaymentOrderSubscriptionTarget struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

func subscriptionTargetFromUser(user *User) *PaymentOrderSubscriptionTarget {
	if user == nil {
		return nil
	}
	return &PaymentOrderSubscriptionTarget{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
	}
}

func buildPaymentOrderSubscriptionTargetSnapshot(target *User) map[string]any {
	targetSnapshot := subscriptionTargetFromUser(target)
	if targetSnapshot == nil || targetSnapshot.UserID <= 0 {
		return nil
	}
	return map[string]any{
		paymentOrderSnapshotSubscriptionTargetKey: map[string]any{
			"user_id":  targetSnapshot.UserID,
			"email":    targetSnapshot.Email,
			"username": targetSnapshot.Username,
		},
	}
}

// PaymentOrderSubscriptionTargetSnapshot returns the target user snapshot for a
// subscription order. Legacy orders do not have this snapshot and should fall
// back to the order owner.
func PaymentOrderSubscriptionTargetSnapshot(order *dbent.PaymentOrder) *PaymentOrderSubscriptionTarget {
	if order == nil || len(order.ProviderSnapshot) == 0 {
		return nil
	}
	raw := order.ProviderSnapshot[paymentOrderSnapshotSubscriptionTargetKey]
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	target := &PaymentOrderSubscriptionTarget{
		UserID:   paymentOrderSnapshotInt64Value(m["user_id"]),
		Email:    strings.TrimSpace(paymentOrderSnapshotStringValue(m["email"])),
		Username: strings.TrimSpace(paymentOrderSnapshotStringValue(m["username"])),
	}
	if target.UserID <= 0 {
		return nil
	}
	return target
}

func paymentOrderSubscriptionTargetUserID(order *dbent.PaymentOrder) int64 {
	if target := PaymentOrderSubscriptionTargetSnapshot(order); target != nil && target.UserID > 0 {
		return target.UserID
	}
	if order == nil {
		return 0
	}
	return order.UserID
}

func paymentOrderSnapshotInt64Value(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(strings.TrimSpace(n), 10, 64)
		return i
	default:
		return 0
	}
}

func paymentOrderSnapshotStringValue(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case json.Number:
		return s.String()
	default:
		return ""
	}
}
