package routes

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterVisitorAnalyticsRoutes(v1 *gin.RouterGroup, h *handler.Handlers, redisClient *redis.Client) {
	if h == nil || h.VisitorAnalytics == nil {
		return
	}
	rateLimiter := middleware.NewRateLimiter(redisClient)
	analytics := v1.Group("/analytics")
	analytics.POST("/visit", rateLimiter.LimitWithOptions("visitor-analytics", 180, time.Minute, middleware.RateLimitOptions{
		FailureMode: middleware.RateLimitFailOpen,
	}), h.VisitorAnalytics.Track)
}
