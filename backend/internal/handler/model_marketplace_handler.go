package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelMarketplaceHandler struct {
	service *service.ModelMarketplaceService
}

func NewModelMarketplaceHandler(service *service.ModelMarketplaceService) *ModelMarketplaceHandler {
	return &ModelMarketplaceHandler{service: service}
}

func (h *ModelMarketplaceHandler) ListPublic(c *gin.Context) {
	items, err := h.service.ListEnabled(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	c.Header("Cache-Control", "public, max-age=60")
	response.Success(c, gin.H{
		"items": items,
	})
}
