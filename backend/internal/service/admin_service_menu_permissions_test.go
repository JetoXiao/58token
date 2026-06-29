//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminService_UpdateUser_KeepsUserMenuPermissionsForRegularUser(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 43, Email: "partner@example.com", Role: RoleUser}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	permissions := []string{"admin_users", "affiliate_usage", "custom:user:99", "custom:admin:3"}
	updated, err := svc.UpdateUser(context.Background(), 43, &UpdateUserInput{
		AdminMenuPermissions: &permissions,
	})

	require.NoError(t, err)
	require.Equal(t, []string{"affiliate_usage", "custom:user:99"}, updated.AdminMenuPermissions)
	require.Equal(t, []string{"affiliate_usage", "custom:user:99"}, repo.lastUpdated.AdminMenuPermissions)
	require.Equal(t, []int64{43}, invalidator.userIDs)
}

func TestAdminService_UpdateUser_ClearsUserMenuPermissionsWhenPromotedToAdmin(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 44, Email: "admin@example.com", Role: RoleUser, AdminMenuPermissions: []string{"affiliate_usage"}}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{
		userRepo:       repo,
		redeemCodeRepo: &redeemRepoStub{},
	}

	role := RoleAdmin
	updated, err := svc.UpdateUser(context.Background(), 44, &UpdateUserInput{
		Role: &role,
	})

	require.NoError(t, err)
	require.Nil(t, updated.AdminMenuPermissions)
	require.Nil(t, repo.lastUpdated.AdminMenuPermissions)
}
