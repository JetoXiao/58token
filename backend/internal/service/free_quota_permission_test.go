package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type freeQuotaPermissionSettingRepo struct {
	values map[string]string
}

func (r *freeQuotaPermissionSettingRepo) Get(_ context.Context, key string) (*Setting, error) {
	return &Setting{Key: key, Value: r.values[key]}, nil
}

func (r *freeQuotaPermissionSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	return r.values[key], nil
}

func (r *freeQuotaPermissionSettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *freeQuotaPermissionSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = r.values[key]
	}
	return out, nil
}

func (r *freeQuotaPermissionSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *freeQuotaPermissionSettingRepo) GetAll(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r.values))
	for key, value := range r.values {
		out[key] = value
	}
	return out, nil
}

func (r *freeQuotaPermissionSettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type freeQuotaPermissionAPIKeyRepo struct {
	key     *APIKey
	created *APIKey
	updated *APIKey
}

func (r *freeQuotaPermissionAPIKeyRepo) Create(_ context.Context, key *APIKey) error {
	clone := *key
	r.created = &clone
	return nil
}

func (r *freeQuotaPermissionAPIKeyRepo) GetByID(context.Context, int64) (*APIKey, error) {
	if r.key == nil {
		return nil, ErrAPIKeyNotFound
	}
	clone := *r.key
	return &clone, nil
}

func (r *freeQuotaPermissionAPIKeyRepo) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) Update(_ context.Context, key *APIKey) error {
	clone := *key
	r.updated = &clone
	return nil
}
func (r *freeQuotaPermissionAPIKeyRepo) Delete(context.Context, int64) error { panic("unexpected") }
func (r *freeQuotaPermissionAPIKeyRepo) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) ExistsByKey(context.Context, string) (bool, error) {
	return false, nil
}
func (r *freeQuotaPermissionAPIKeyRepo) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) ListKeysByUserID(context.Context, int64) ([]string, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionAPIKeyRepo) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected")
}

type freeQuotaPermissionUserRepo struct {
	user *User
}

func (r *freeQuotaPermissionUserRepo) Create(context.Context, *User) error { panic("unexpected") }
func (r *freeQuotaPermissionUserRepo) GetByID(context.Context, int64) (*User, error) {
	if r.user == nil {
		return nil, ErrUserNotFound
	}
	clone := *r.user
	return &clone, nil
}
func (r *freeQuotaPermissionUserRepo) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) Update(context.Context, *User) error { panic("unexpected") }
func (r *freeQuotaPermissionUserRepo) Delete(context.Context, int64) error { panic("unexpected") }
func (r *freeQuotaPermissionUserRepo) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserRepo) EnableTotp(context.Context, int64) error { panic("unexpected") }
func (r *freeQuotaPermissionUserRepo) DisableTotp(context.Context, int64) error {
	panic("unexpected")
}

type freeQuotaPermissionGroupRepo struct {
	group  *Group
	groups []Group
}

func (r *freeQuotaPermissionGroupRepo) Create(context.Context, *Group) error { panic("unexpected") }
func (r *freeQuotaPermissionGroupRepo) GetByID(context.Context, int64) (*Group, error) {
	if r.group == nil {
		return nil, ErrGroupNotFound
	}
	clone := *r.group
	return &clone, nil
}
func (r *freeQuotaPermissionGroupRepo) GetByIDLite(context.Context, int64) (*Group, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) Update(context.Context, *Group) error { panic("unexpected") }
func (r *freeQuotaPermissionGroupRepo) Delete(context.Context, int64) error  { panic("unexpected") }
func (r *freeQuotaPermissionGroupRepo) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) ListActive(context.Context) ([]Group, error) {
	if r.groups != nil {
		out := make([]Group, len(r.groups))
		copy(out, r.groups)
		return out, nil
	}
	if r.group == nil {
		return nil, nil
	}
	return []Group{*r.group}, nil
}
func (r *freeQuotaPermissionGroupRepo) ListActiveByPlatform(context.Context, string) ([]Group, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) GetAccountCount(context.Context, int64) (int64, int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionGroupRepo) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected")
}

type freeQuotaPermissionUserSubRepo struct {
	activeSubscriptions []UserSubscription
}

func (r *freeQuotaPermissionUserSubRepo) Create(context.Context, *UserSubscription) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) GetByID(context.Context, int64) (*UserSubscription, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) GetByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) GetActiveByUserIDAndGroupID(_ context.Context, userID, groupID int64) (*UserSubscription, error) {
	for i := range r.activeSubscriptions {
		sub := r.activeSubscriptions[i]
		if sub.UserID == userID && sub.GroupID == groupID {
			return &sub, nil
		}
	}
	return nil, ErrSubscriptionNotFound
}
func (r *freeQuotaPermissionUserSubRepo) Update(context.Context, *UserSubscription) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) Delete(context.Context, int64) error { panic("unexpected") }
func (r *freeQuotaPermissionUserSubRepo) ListByUserID(context.Context, int64) ([]UserSubscription, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ListVisibleByUserID(context.Context, int64) ([]UserSubscription, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ListActiveByUserID(_ context.Context, userID int64) ([]UserSubscription, error) {
	out := make([]UserSubscription, 0, len(r.activeSubscriptions))
	for _, sub := range r.activeSubscriptions {
		if sub.UserID == userID {
			out = append(out, sub)
		}
	}
	return out, nil
}
func (r *freeQuotaPermissionUserSubRepo) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) List(context.Context, pagination.PaginationParams, *int64, *int64, string, string, string, string) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ExistsByUserIDAndGroupID(context.Context, int64, int64) (bool, error) {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ExtendExpiry(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) UpdateStatus(context.Context, int64, string) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) UpdateNotes(context.Context, int64, string) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ActivateWindows(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ResetDailyUsage(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ResetWeeklyUsage(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) ResetMonthlyUsage(context.Context, int64, time.Time) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) IncrementUsage(context.Context, int64, float64) error {
	panic("unexpected")
}
func (r *freeQuotaPermissionUserSubRepo) BatchUpdateExpiredStatus(context.Context) (int64, error) {
	panic("unexpected")
}

type freeQuotaGrantRepo struct {
	grants []FreeQuotaLedger
}

func (r *freeQuotaGrantRepo) Grant(_ context.Context, grant *FreeQuotaLedger) error {
	if grant == nil {
		return nil
	}
	r.grants = append(r.grants, *grant)
	return nil
}
func (r *freeQuotaGrantRepo) GetAvailableForGroup(context.Context, int64, int64) (float64, error) {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) RedeemTrialCard(context.Context, int64, string, []int64) (*TrialCard, *FreeQuotaLedger, error) {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) ListTrialCards(context.Context, int, int) ([]TrialCard, int64, error) {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) CreateTrialCard(context.Context, *CreateTrialCardInput) (*TrialCard, error) {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) UpdateTrialCard(context.Context, int64, *UpdateTrialCardInput) (*TrialCard, error) {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) DeleteTrialCard(context.Context, int64) error { panic("unexpected") }
func (r *freeQuotaGrantRepo) MarkPaymentSucceeded(context.Context, int64, float64, bool) error {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) GetUserBalance(context.Context, int64) (float64, error) {
	panic("unexpected")
}
func (r *freeQuotaGrantRepo) GetUserFreeQuotaSummary(context.Context, int64) (*FreeQuotaSummary, error) {
	panic("unexpected")
}

func newFreeQuotaGrantService(t *testing.T, repo *freeQuotaGrantRepo, enabled bool, amount float64, freeGroupIDs []int64) *FreeQuotaService {
	t.Helper()

	rawGroups, err := json.Marshal(freeGroupIDs)
	require.NoError(t, err)

	return NewFreeQuotaService(repo, &SettingService{
		settingRepo: &freeQuotaPermissionSettingRepo{values: map[string]string{
			SettingKeyFreeQuotaInvitationEnabled: strconv.FormatBool(enabled),
			SettingKeyFreeQuotaInvitationAmount:  strconv.FormatFloat(amount, 'f', -1, 64),
			SettingKeyFreeQuotaGroupIDs:          string(rawGroups),
		}},
	})
}

func newFreeQuotaPermissionService(t *testing.T, freeGroupIDs []int64) *FreeQuotaService {
	t.Helper()

	rawGroups, err := json.Marshal(freeGroupIDs)
	require.NoError(t, err)

	return NewFreeQuotaService(nil, &SettingService{
		settingRepo: &freeQuotaPermissionSettingRepo{values: map[string]string{
			SettingKeyFreeQuotaGroupIDs: string(rawGroups),
		}},
	})
}

func newFreeQuotaPermissionServiceWithVisibility(t *testing.T, freeGroupIDs []int64, showLockedGroups bool) *FreeQuotaService {
	t.Helper()

	rawGroups, err := json.Marshal(freeGroupIDs)
	require.NoError(t, err)

	return NewFreeQuotaService(nil, &SettingService{
		settingRepo: &freeQuotaPermissionSettingRepo{values: map[string]string{
			SettingKeyFreeQuotaGroupIDs:         string(rawGroups),
			SettingKeyFreeQuotaShowLockedGroups: strconv.FormatBool(showLockedGroups),
		}},
	})
}

func TestAPIKeyServiceCreateRejectsNonFreeQuotaGroupForZeroBalanceUser(t *testing.T) {
	ctx := context.Background()
	freeGroupID := int64(10)
	blockedGroupID := int64(20)
	user := &User{ID: 1, Status: StatusActive, Balance: 0}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               blockedGroupID,
			Status:           StatusActive,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		nil,
		nil,
		nil,
		nil,
	)
	svc.SetFreeQuotaService(newFreeQuotaPermissionService(t, []int64{freeGroupID}))

	_, err := svc.Create(ctx, user.ID, CreateAPIKeyRequest{
		Name:    "blocked",
		GroupID: &blockedGroupID,
	})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Nil(t, apiKeyRepo.created, "backend must not persist an API key bound to a non-free group")
}

func TestAPIKeyServiceCreateAllowsAssignedExclusiveGroup(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	user := &User{ID: 1, Status: StatusActive, Balance: 100, AllowedGroups: []int64{groupID}}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               groupID,
			Status:           StatusActive,
			IsExclusive:      true,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		nil,
		nil,
		nil,
		nil,
	)

	customKey := "assignedexclusive"
	_, err := svc.Create(ctx, user.ID, CreateAPIKeyRequest{
		Name:      "assigned-exclusive",
		GroupID:   &groupID,
		CustomKey: &customKey,
	})

	require.NoError(t, err)
	require.NotNil(t, apiKeyRepo.created)
	require.Equal(t, groupID, *apiKeyRepo.created.GroupID)
}

func TestAPIKeyServiceCreateRejectsUnassignedExclusiveGroup(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	user := &User{ID: 1, Status: StatusActive, Balance: 100}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               groupID,
			Status:           StatusActive,
			IsExclusive:      true,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := svc.Create(ctx, user.ID, CreateAPIKeyRequest{
		Name:    "unassigned-exclusive",
		GroupID: &groupID,
	})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Nil(t, apiKeyRepo.created, "backend must not persist an API key bound to an unassigned exclusive group")
}

func TestAPIKeyServiceCreateAllowsSubscribedExclusiveSubscriptionGroup(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	user := &User{ID: 1, Status: StatusActive, Balance: 0}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               groupID,
			Status:           StatusActive,
			IsExclusive:      true,
			SubscriptionType: SubscriptionTypeSubscription,
		}},
		&freeQuotaPermissionUserSubRepo{activeSubscriptions: []UserSubscription{{
			UserID:  user.ID,
			GroupID: groupID,
		}}},
		nil,
		nil,
		nil,
	)

	customKey := "subscribed-exclusive-key"
	_, err := svc.Create(ctx, user.ID, CreateAPIKeyRequest{
		Name:      "subscribed-exclusive",
		GroupID:   &groupID,
		CustomKey: &customKey,
	})

	require.NoError(t, err)
	require.NotNil(t, apiKeyRepo.created)
	require.Equal(t, groupID, *apiKeyRepo.created.GroupID)
}

func TestAPIKeyServiceUpdateRejectsNonFreeQuotaGroupForZeroBalanceUser(t *testing.T) {
	ctx := context.Background()
	freeGroupID := int64(10)
	blockedGroupID := int64(20)
	user := &User{ID: 1, Status: StatusActive, Balance: 0}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{
		key: &APIKey{ID: 99, UserID: user.ID, Key: "sk-test", Status: StatusActive},
	}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               blockedGroupID,
			Status:           StatusActive,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		nil,
		nil,
		nil,
		nil,
	)
	svc.SetFreeQuotaService(newFreeQuotaPermissionService(t, []int64{freeGroupID}))

	_, err := svc.Update(ctx, 99, user.ID, UpdateAPIKeyRequest{GroupID: &blockedGroupID})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Nil(t, apiKeyRepo.updated, "backend must not persist a forged group change")
}

func TestAPIKeyServiceUpdateAllowsAssignedExclusiveGroup(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	user := &User{ID: 1, Status: StatusActive, Balance: 100, AllowedGroups: []int64{groupID}}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{
		key: &APIKey{ID: 99, UserID: user.ID, Key: "sk-test", Status: StatusActive},
	}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               groupID,
			Status:           StatusActive,
			IsExclusive:      true,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := svc.Update(ctx, 99, user.ID, UpdateAPIKeyRequest{GroupID: &groupID})

	require.NoError(t, err)
	require.NotNil(t, apiKeyRepo.updated)
	require.Equal(t, groupID, *apiKeyRepo.updated.GroupID)
}

func TestAPIKeyServiceUpdateRejectsUnassignedExclusiveGroup(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	user := &User{ID: 1, Status: StatusActive, Balance: 100}

	apiKeyRepo := &freeQuotaPermissionAPIKeyRepo{
		key: &APIKey{ID: 99, UserID: user.ID, Key: "sk-test", Status: StatusActive},
	}
	svc := NewAPIKeyService(
		apiKeyRepo,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{group: &Group{
			ID:               groupID,
			Status:           StatusActive,
			IsExclusive:      true,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := svc.Update(ctx, 99, user.ID, UpdateAPIKeyRequest{GroupID: &groupID})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Nil(t, apiKeyRepo.updated, "backend must not persist a forged unassigned exclusive group change")
}

func TestAPIKeyServiceCanUserUseGroupRejectsNonFreeQuotaGroupAtRuntime(t *testing.T) {
	ctx := context.Background()
	freeGroupID := int64(10)
	blockedGroupID := int64(20)

	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	svc.SetFreeQuotaService(newFreeQuotaPermissionService(t, []int64{freeGroupID}))

	user := &User{ID: 1, Status: StatusActive, Balance: 0}
	require.True(t, svc.CanUserUseGroup(ctx, user, &Group{ID: freeGroupID, SubscriptionType: SubscriptionTypeStandard}))
	require.False(t, svc.CanUserUseGroup(ctx, user, &Group{ID: blockedGroupID, SubscriptionType: SubscriptionTypeStandard}))

	user.Balance = 1
	require.True(t, svc.CanUserUseGroup(ctx, user, &Group{ID: blockedGroupID, SubscriptionType: SubscriptionTypeStandard}))
}

func TestAPIKeyServiceCanUserUseGroupChecksExclusiveAssignment(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)

	user := &User{ID: 1, Status: StatusActive, Balance: 100}
	require.False(t, svc.CanUserUseGroup(ctx, user, &Group{
		ID:               groupID,
		IsExclusive:      true,
		SubscriptionType: SubscriptionTypeStandard,
	}))

	user.AllowedGroups = []int64{groupID}
	require.True(t, svc.CanUserUseGroup(ctx, user, &Group{
		ID:               groupID,
		IsExclusive:      true,
		SubscriptionType: SubscriptionTypeStandard,
	}))
}

func TestFreeQuotaServiceGrantInvitationQuotaRequiresEnabledSetting(t *testing.T) {
	repo := &freeQuotaGrantRepo{}
	svc := newFreeQuotaGrantService(t, repo, false, 10, []int64{10})

	err := svc.GrantInvitationQuota(context.Background(), 1)

	require.NoError(t, err)
	require.Empty(t, repo.grants)
}

func TestFreeQuotaServiceGrantInvitationQuotaWhenEnabled(t *testing.T) {
	repo := &freeQuotaGrantRepo{}
	svc := newFreeQuotaGrantService(t, repo, true, 10, []int64{10})

	err := svc.GrantInvitationQuota(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, repo.grants, 1)
	require.Equal(t, int64(1), repo.grants[0].UserID)
	require.Equal(t, FreeQuotaSourceInvitation, repo.grants[0].SourceType)
	require.Equal(t, 10.0, repo.grants[0].Amount)
	require.Equal(t, []int64{10}, repo.grants[0].AllowedGroupIDs)
}

func TestAPIKeyServiceGetAvailableGroupsHidesLockedGroupsWhenSettingDisabled(t *testing.T) {
	ctx := context.Background()
	freeGroupID := int64(10)
	blockedGroupID := int64(20)
	user := &User{ID: 1, Status: StatusActive, Balance: 0}

	svc := NewAPIKeyService(
		nil,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{groups: []Group{
			{ID: freeGroupID, Name: "free", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			{ID: blockedGroupID, Name: "blocked", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
		}},
		&freeQuotaPermissionUserSubRepo{},
		nil,
		nil,
		nil,
	)
	svc.SetFreeQuotaService(newFreeQuotaPermissionServiceWithVisibility(t, []int64{freeGroupID}, false))

	groups, err := svc.GetAvailableGroups(ctx, user.ID)

	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Equal(t, freeGroupID, groups[0].ID)
	require.False(t, groups[0].Locked)
}

func TestAPIKeyServiceGetAvailableGroupsIncludesOnlyAssignedExclusiveGroups(t *testing.T) {
	ctx := context.Background()
	publicGroupID := int64(10)
	assignedExclusiveGroupID := int64(20)
	unassignedExclusiveGroupID := int64(30)
	user := &User{
		ID:            1,
		Status:        StatusActive,
		Balance:       100,
		AllowedGroups: []int64{assignedExclusiveGroupID},
	}

	svc := NewAPIKeyService(
		nil,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{groups: []Group{
			{ID: publicGroupID, Name: "public", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			{ID: assignedExclusiveGroupID, Name: "assigned", Status: StatusActive, IsExclusive: true, SubscriptionType: SubscriptionTypeStandard},
			{ID: unassignedExclusiveGroupID, Name: "unassigned", Status: StatusActive, IsExclusive: true, SubscriptionType: SubscriptionTypeStandard},
		}},
		&freeQuotaPermissionUserSubRepo{},
		nil,
		nil,
		nil,
	)

	groups, err := svc.GetAvailableGroups(ctx, user.ID)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, publicGroupID, groups[0].ID)
	require.False(t, groups[0].IsExclusive)
	require.Equal(t, assignedExclusiveGroupID, groups[1].ID)
	require.True(t, groups[1].IsExclusive)
}

func TestAPIKeyServiceGetAvailableGroupsIncludesSubscribedExclusiveSubscriptionGroup(t *testing.T) {
	ctx := context.Background()
	publicGroupID := int64(10)
	subscribedSubscriptionGroupID := int64(20)
	unsubscribedSubscriptionGroupID := int64(30)
	user := &User{ID: 1, Status: StatusActive, Balance: 0}

	svc := NewAPIKeyService(
		nil,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{groups: []Group{
			{ID: publicGroupID, Name: "public", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			{ID: subscribedSubscriptionGroupID, Name: "subscribed", Status: StatusActive, IsExclusive: true, SubscriptionType: SubscriptionTypeSubscription},
			{ID: unsubscribedSubscriptionGroupID, Name: "unsubscribed", Status: StatusActive, IsExclusive: true, SubscriptionType: SubscriptionTypeSubscription},
		}},
		&freeQuotaPermissionUserSubRepo{activeSubscriptions: []UserSubscription{{
			UserID:  user.ID,
			GroupID: subscribedSubscriptionGroupID,
		}}},
		nil,
		nil,
		nil,
	)

	groups, err := svc.GetAvailableGroups(ctx, user.ID)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, publicGroupID, groups[0].ID)
	require.Equal(t, subscribedSubscriptionGroupID, groups[1].ID)
	require.True(t, groups[1].IsExclusive)
	require.Equal(t, SubscriptionTypeSubscription, groups[1].SubscriptionType)
}

func TestAPIKeyServiceGetAvailableGroupsShowsLockedGroupsWhenSettingEnabled(t *testing.T) {
	ctx := context.Background()
	freeGroupID := int64(10)
	blockedGroupID := int64(20)
	user := &User{ID: 1, Status: StatusActive, Balance: 0}

	svc := NewAPIKeyService(
		nil,
		&freeQuotaPermissionUserRepo{user: user},
		&freeQuotaPermissionGroupRepo{groups: []Group{
			{ID: freeGroupID, Name: "free", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
			{ID: blockedGroupID, Name: "blocked", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard},
		}},
		&freeQuotaPermissionUserSubRepo{},
		nil,
		nil,
		nil,
	)
	svc.SetFreeQuotaService(newFreeQuotaPermissionServiceWithVisibility(t, []int64{freeGroupID}, true))

	groups, err := svc.GetAvailableGroups(ctx, user.ID)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, freeGroupID, groups[0].ID)
	require.False(t, groups[0].Locked)
	require.Equal(t, blockedGroupID, groups[1].ID)
	require.True(t, groups[1].Locked)
	require.Equal(t, "payment_required", groups[1].LockReason)
}
