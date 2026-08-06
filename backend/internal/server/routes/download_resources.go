package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterDownloadResourceRoutes(v1 *gin.RouterGroup, h *handler.Handlers, jwtAuth middleware.JWTAuthMiddleware) {
	if h == nil || h.DownloadResources == nil {
		return
	}
	resources := v1.Group("/public/download-resources")
	{
		resources.GET("", h.DownloadResources.ListPublished)
		resources.POST("/:id/download", middleware.OptionalJWTAuth(jwtAuth), h.DownloadResources.IssueDownload)
	}
}

func registerDownloadResourceAdminRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	if h == nil || h.DownloadResources == nil {
		return
	}
	resources := admin.Group("/download-resources")
	{
		resources.GET("", h.DownloadResources.ListAll)
		resources.GET("/downloads", h.DownloadResources.ListDownloads)
		if h.VisitorAnalytics != nil {
			resources.POST("/ip-lookup", h.VisitorAnalytics.LookupIP)
		}
		resources.GET("/storage", h.DownloadResources.GetS3Config)
		resources.PUT("/storage", h.DownloadResources.UpdateS3Config)
		resources.POST("/storage/test", h.DownloadResources.TestS3Config)
		resources.POST("/upload-url", h.DownloadResources.CreateUploadURL)
		resources.POST("", h.DownloadResources.Create)
		resources.PUT("/:id", h.DownloadResources.Update)
		resources.DELETE("/:id", h.DownloadResources.Delete)
	}
}
