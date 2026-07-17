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
			name:        "sub_admin_help_center_uses_settings_permission",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_settings"},
			method:      http.MethodGet,
			path:        "/api/v1/admin/help-center",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "sub_admin_ttft_analysis_uses_dedicated_permission",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_ttft_analysis"},
			method:      http.MethodGet,
			path:        "/api/v1/admin/ops/ttft-analysis",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "sub_admin_settings_help_center_alias_uses_settings_permission",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_settings"},
			method:      http.MethodGet,
			path:        "/api/v1/admin/settings/help-center",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "sub_admin_trial_cards_uses_redeem_permission",
			role:        service.RoleSubAdmin,
			permissions: []string{"admin_redeem"},
			method:      http.MethodGet,
			path:        "/api/v1/admin/trial-cards",
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
