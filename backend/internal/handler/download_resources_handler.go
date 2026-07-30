package handler

import (
	"errors"
	"net"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DownloadResourcesHandler struct {
	service *service.DownloadResourceService
}

func NewDownloadResourcesHandler(downloadService *service.DownloadResourceService) *DownloadResourcesHandler {
	return &DownloadResourcesHandler{service: downloadService}
}

func (h *DownloadResourcesHandler) ListPublished(c *gin.Context) {
	items, err := h.service.ListPublished(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load download resources")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *DownloadResourcesHandler) IssueDownload(c *gin.Context) {
	id, ok := downloadResourceIDParam(c)
	if !ok {
		return
	}
	ip := visitorClientIP(c)
	if net.ParseIP(ip) == nil {
		response.Error(c, http.StatusServiceUnavailable, "Download protection is temporarily unavailable")
		return
	}
	url, err := h.service.IssueDownload(c.Request.Context(), id, ip, c.GetHeader("User-Agent"), c.GetHeader("Referer"))
	if err != nil {
		handleDownloadResourceError(c, err)
		return
	}
	response.Success(c, gin.H{"url": url})
}

func (h *DownloadResourcesHandler) ListAll(c *gin.Context) {
	items, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load download resources")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *DownloadResourcesHandler) Create(c *gin.Context) {
	var input service.DownloadResourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid download resource payload")
		return
	}
	item, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		handleDownloadResourceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *DownloadResourcesHandler) Update(c *gin.Context) {
	id, ok := downloadResourceIDParam(c)
	if !ok {
		return
	}
	var input service.DownloadResourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid download resource payload")
		return
	}
	item, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		handleDownloadResourceError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DownloadResourcesHandler) Delete(c *gin.Context) {
	id, ok := downloadResourceIDParam(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleDownloadResourceError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *DownloadResourcesHandler) ListDownloads(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.ListDownloads(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to load download records")
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DownloadResourcesHandler) GetS3Config(c *gin.Context) {
	if isReadOnlyDownloadResourceAdmin(c) {
		response.Forbidden(c, "Download storage configuration requires a full administrator")
		return
	}
	cfg, err := h.service.GetS3Config(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load download storage configuration")
		return
	}
	response.Success(c, cfg)
}

func (h *DownloadResourcesHandler) UpdateS3Config(c *gin.Context) {
	if isReadOnlyDownloadResourceAdmin(c) {
		response.Forbidden(c, "Download storage configuration requires a full administrator")
		return
	}
	var cfg service.DownloadResourceS3Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "Invalid download storage configuration")
		return
	}
	updated, err := h.service.UpdateS3Config(c.Request.Context(), cfg)
	if err != nil {
		handleDownloadResourceError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *DownloadResourcesHandler) TestS3Config(c *gin.Context) {
	if isReadOnlyDownloadResourceAdmin(c) {
		response.Forbidden(c, "Download storage configuration requires a full administrator")
		return
	}
	var cfg service.DownloadResourceS3Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "Invalid download storage configuration")
		return
	}
	if err := h.service.TestS3Connection(c.Request.Context(), cfg); err != nil {
		response.BadRequest(c, "R2 connection test failed: "+err.Error())
		return
	}
	response.Success(c, gin.H{"connected": true})
}

func (h *DownloadResourcesHandler) CreateUploadURL(c *gin.Context) {
	if isReadOnlyDownloadResourceAdmin(c) {
		response.Forbidden(c, "Upload requires a full administrator")
		return
	}
	var request service.DownloadResourceUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid upload request")
		return
	}
	result, err := h.service.CreateUploadURL(c.Request.Context(), request)
	if err != nil {
		handleDownloadResourceError(c, err)
		return
	}
	response.Success(c, result)
}

func downloadResourceIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid download resource ID")
		return 0, false
	}
	return id, true
}

func isReadOnlyDownloadResourceAdmin(c *gin.Context) bool {
	role, _ := middleware.GetUserRoleFromContext(c)
	return role == service.RoleSubAdmin
}

func handleDownloadResourceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrDownloadResourceNotFound):
		response.NotFound(c, "Download resource not found")
	case errors.Is(err, service.ErrDownloadRateLimited):
		response.Error(c, http.StatusTooManyRequests, "Too many download requests. Please try again later.")
	case errors.Is(err, service.ErrDownloadStorageNotReady):
		response.Error(c, http.StatusServiceUnavailable, "Download storage is not ready")
	default:
		response.BadRequest(c, "Unable to process the download resource request")
	}
}
