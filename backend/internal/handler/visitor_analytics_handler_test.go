package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVisitorClientIPOnlyTrustsProxyHeadersFromPrivatePeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		remoteAddr string
		xRealIP    string
		want       string
	}{
		{name: "local caddy", remoteAddr: "127.0.0.1:55123", xRealIP: "203.0.113.8", want: "203.0.113.8"},
		{name: "docker bridge", remoteAddr: "172.18.0.1:55123", xRealIP: "198.51.100.9", want: "198.51.100.9"},
		{name: "direct public cannot spoof", remoteAddr: "203.0.113.20:55123", xRealIP: "198.51.100.9", want: "203.0.113.20"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/analytics/visit", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Real-IP", tt.xRealIP)
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = req
			if got := visitorClientIP(ctx); got != tt.want {
				t.Fatalf("visitorClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVisitorTrustsProxyHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		remoteAddr string
		want       bool
	}{
		{remoteAddr: "127.0.0.1:55123", want: true},
		{remoteAddr: "172.18.0.1:55123", want: true},
		{remoteAddr: "203.0.113.20:55123", want: false},
	} {
		req := httptest.NewRequest("POST", "/api/v1/analytics/visit", nil)
		req.RemoteAddr = tt.remoteAddr
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = req
		if got := visitorTrustsProxyHeaders(ctx); got != tt.want {
			t.Fatalf("visitorTrustsProxyHeaders(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
		}
	}
}
