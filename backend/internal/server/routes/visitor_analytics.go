package routes

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	appmiddleware "github.com/Wei-Shaw/sub2api/internal/middleware"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterVisitorAnalyticsRoutes(v1 *gin.RouterGroup, h *handler.Handlers, jwtAuth servermiddleware.JWTAuthMiddleware, redisClient *redis.Client) {
	if h == nil || h.VisitorAnalytics == nil {
		return
	}
	rateLimiter := appmiddleware.NewRateLimiter(redisClient)
	analytics := v1.Group("/analytics")
	analytics.POST("/visit", servermiddleware.OptionalJWTAuth(jwtAuth), rateLimiter.LimitWithOptions("visitor-analytics", 180, time.Minute, appmiddleware.RateLimitOptions{
		FailureMode: appmiddleware.RateLimitFailOpen,
	}), h.VisitorAnalytics.Track)
}
