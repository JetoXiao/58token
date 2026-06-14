package middleware

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func SubAdminAdminPermissionGuard() gin.HandlerFunc {
	return subAdminPermissionGuard(true)
}

func SubAdminUserPermissionGuard() gin.HandlerFunc {
	return subAdminPermissionGuard(false)
}

func subAdminPermissionGuard(adminScope bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := GetUserRoleFromContext(c)
		if role != service.RoleSubAdmin {
			c.Next()
			return
		}

		if !isReadOnlyMethod(c.Request.Method) && !isReadOnlyQueryEndpoint(c.Request.Method, c.Request.URL.Path, adminScope) {
			AbortWithError(c, http.StatusForbidden, "READ_ONLY_ADMIN", "This admin role has read-only access")
			return
		}
		if isAuxiliaryReadEndpoint(c.Request.Method, c.Request.URL.Path, adminScope) {
			c.Next()
			return
		}

		key := ""
		if adminScope {
			key = adminMenuKeyForAPIPath(c.Request.URL.Path)
		} else {
			key = userMenuKeyForAPIPath(c.Request.URL.Path)
		}
		if key == "" || !hasAdminMenuPermission(c, key) {
			AbortWithError(c, http.StatusForbidden, "MENU_FORBIDDEN", "Menu permission required")
			return
		}

		c.Next()
	}
}

func isReadOnlyMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func isReadOnlyQueryEndpoint(method, path string, adminScope bool) bool {
	if method != http.MethodPost {
		return false
	}
	p := trimAPIV1(path)
	if adminScope {
		switch p {
		case "/admin/dashboard/users-usage",
			"/admin/dashboard/api-keys-usage",
			"/admin/accounts/today-stats/batch",
			"/admin/user-attributes/batch":
			return true
		default:
			return false
		}
	}
	return p == "/usage/dashboard/api-keys-usage"
}

func isAuxiliaryReadEndpoint(method, path string, adminScope bool) bool {
	if !isReadOnlyMethod(method) || adminScope {
		return false
	}
	p := trimAPIV1(path)
	return p == "/user/profile" || strings.HasPrefix(p, "/announcements")
}

func hasAdminMenuPermission(c *gin.Context, key string) bool {
	items, ok := GetAdminMenuPermissionsFromContext(c)
	if !ok || key == "" {
		return false
	}
	for _, item := range items {
		if item == key {
			return true
		}
	}
	return false
}

func adminMenuKeyForAPIPath(path string) string {
	p := trimAPIV1(path)
	if !strings.HasPrefix(p, "/admin/") {
		return ""
	}
	p = strings.TrimPrefix(p, "/admin")

	switch {
	case strings.HasPrefix(p, "/dashboard"):
		return "admin_dashboard"
	case strings.HasPrefix(p, "/ops/response-cache"):
		return "admin_response_cache"
	case strings.HasPrefix(p, "/ops/requests"):
		return "admin_requests"
	case strings.HasPrefix(p, "/ops"):
		return "admin_ops"
	case strings.HasPrefix(p, "/users") || strings.HasPrefix(p, "/user-attributes"):
		return "admin_users"
	case strings.HasPrefix(p, "/groups"):
		return "admin_groups"
	case strings.HasPrefix(p, "/channels"):
		return "admin_channel_pricing"
	case strings.HasPrefix(p, "/channel-monitors") || strings.HasPrefix(p, "/channel-monitor-templates"):
		return "admin_channel_monitor"
	case strings.HasPrefix(p, "/subscriptions"):
		return "admin_subscriptions"
	case strings.HasPrefix(p, "/accounts") || strings.HasPrefix(p, "/openai") || strings.HasPrefix(p, "/gemini") || strings.HasPrefix(p, "/antigravity") || strings.HasPrefix(p, "/api-keys"):
		return "admin_accounts"
	case strings.HasPrefix(p, "/announcements"):
		return "admin_announcements"
	case strings.HasPrefix(p, "/proxies"):
		return "admin_proxies"
	case strings.HasPrefix(p, "/risk-control"):
		return "admin_risk_control"
	case strings.HasPrefix(p, "/redeem-codes"):
		return "admin_redeem"
	case strings.HasPrefix(p, "/promo-codes"):
		return "admin_promo_codes"
	case strings.HasPrefix(p, "/affiliates/usage"):
		return "admin_affiliate_usage"
	case strings.HasPrefix(p, "/affiliates/partner-applications"):
		return "admin_affiliate_applications"
	case strings.HasPrefix(p, "/affiliates/invites"):
		return "admin_affiliate_invites"
	case strings.HasPrefix(p, "/affiliates/rebates"):
		return "admin_affiliate_rebates"
	case strings.HasPrefix(p, "/affiliates/transfers"):
		return "admin_affiliate_transfers"
	case strings.HasPrefix(p, "/affiliates"):
		return "admin_affiliate_usage"
	case strings.HasPrefix(p, "/payment/dashboard"):
		return "admin_order_dashboard"
	case strings.HasPrefix(p, "/payment/orders"):
		return "admin_orders"
	case strings.HasPrefix(p, "/payment/plans"):
		return "admin_order_plans"
	case strings.HasPrefix(p, "/payment/config") || strings.HasPrefix(p, "/payment/providers"):
		return "admin_order_dashboard"
	case strings.HasPrefix(p, "/usage"):
		return "admin_usage"
	case strings.HasPrefix(p, "/settings") || strings.HasPrefix(p, "/system") || strings.HasPrefix(p, "/data-management") || strings.HasPrefix(p, "/backups") || strings.HasPrefix(p, "/scheduled-test-plans") || strings.HasPrefix(p, "/error-passthrough-rules") || strings.HasPrefix(p, "/tls-fingerprint-profiles"):
		return "admin_settings"
	default:
		return ""
	}
}

func userMenuKeyForAPIPath(path string) string {
	p := trimAPIV1(path)
	switch {
	case strings.HasPrefix(p, "/usage/dashboard"):
		return "dashboard"
	case strings.HasPrefix(p, "/keys") || strings.HasPrefix(p, "/groups"):
		return "api_keys"
	case strings.HasPrefix(p, "/usage"):
		return "usage"
	case strings.HasPrefix(p, "/channels") || strings.HasPrefix(p, "/channel-monitors"):
		return "channel_status"
	case strings.HasPrefix(p, "/subscriptions"):
		return "subscriptions"
	case strings.HasPrefix(p, "/payment"):
		if strings.HasPrefix(p, "/payment/orders") {
			return "orders"
		}
		return "purchase"
	case strings.HasPrefix(p, "/redeem"):
		return "redeem"
	case strings.HasPrefix(p, "/user/aff"):
		return "affiliate"
	case strings.HasPrefix(p, "/user"):
		return "profile"
	case strings.HasPrefix(p, "/pages"):
		return "profile"
	default:
		return ""
	}
}

func trimAPIV1(path string) string {
	if strings.HasPrefix(path, "/api/v1") {
		return strings.TrimPrefix(path, "/api/v1")
	}
	return path
}
