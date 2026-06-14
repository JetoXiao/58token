//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSubAdminAdminPermissionGuard(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		permissions []string
		method      string
		path        string
		wantStatus  int
		wantBody    string
	}{
		{
			name:       "super_admin_not_limited",
			role:       service.RoleAdmin,
			method:     http.MethodPost,
			path:       "/api/v1/admin/users",
			wantStatus: http.StatusOK,
		},
		{
			name:        "sub_admin_authorized_get_allowed",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_users"},
			method:      http.MethodGet,
			path:        "/api/v1/admin/users",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "sub_admin_query_post_allowed",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_dashboard"},
			method:      http.MethodPost,
			path:        "/api/v1/admin/dashboard/users-usage",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "sub_admin_missing_menu_denied",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_groups"},
			method:      http.MethodGet,
			path:        "/api/v1/admin/users",
			wantStatus:  http.StatusForbidden,
			wantBody:    "MENU_FORBIDDEN",
		},
		{
			name:        "sub_admin_write_denied",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_users"},
			method:      http.MethodPost,
			path:        "/api/v1/admin/users",
			wantStatus:  http.StatusForbidden,
			wantBody:    "READ_ONLY_ADMIN",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			code, body := runSubAdminGuardRequest(SubAdminAdminPermissionGuard(), tc.role, tc.permissions, tc.method, tc.path)
			require.Equal(t, tc.wantStatus, code)
			if tc.wantBody != "" {
				require.Contains(t, body, tc.wantBody)
			}
		})
	}
}

func TestSubAdminUserPermissionGuard(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		method      string
		path        string
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "authorized_get_allowed",
			permissions: []string{"orders"},
			method:      http.MethodGet,
			path:        "/api/v1/payment/orders",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "query_post_allowed",
			permissions: []string{"dashboard"},
			method:      http.MethodPost,
			path:        "/api/v1/usage/dashboard/api-keys-usage",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "auxiliary_announcements_get_allowed",
			permissions: nil,
			method:      http.MethodGet,
			path:        "/api/v1/announcements",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "auxiliary_profile_get_allowed",
			permissions: nil,
			method:      http.MethodGet,
			path:        "/api/v1/user/profile",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "missing_menu_denied",
			permissions: []string{"profile"},
			method:      http.MethodGet,
			path:        "/api/v1/payment/orders",
			wantStatus:  http.StatusForbidden,
			wantBody:    "MENU_FORBIDDEN",
		},
		{
			name:        "write_denied",
			permissions: []string{"api_keys"},
			method:      http.MethodPost,
			path:        "/api/v1/keys",
			wantStatus:  http.StatusForbidden,
			wantBody:    "READ_ONLY_ADMIN",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			code, body := runSubAdminGuardRequest(SubAdminUserPermissionGuard(), service.RoleSubAdmin, tc.permissions, tc.method, tc.path)
			require.Equal(t, tc.wantStatus, code)
			if tc.wantBody != "" {
				require.Contains(t, body, tc.wantBody)
			}
		})
	}
}

func runSubAdminGuardRequest(guard gin.HandlerFunc, role string, permissions []string, method string, path string) (int, string) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), role)
		if permissions != nil {
			c.Set(string(ContextKeyAdminMenuPermissions), permissions)
		}
		c.Next()
	})
	router.Use(guard)
	router.Any("/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	router.ServeHTTP(w, req)
	return w.Code, w.Body.String()
}
