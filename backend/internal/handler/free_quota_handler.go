package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type FreeQuotaHandler struct {
	freeQuotaService *service.FreeQuotaService
}

func NewFreeQuotaHandler(freeQuotaService *service.FreeQuotaService) *FreeQuotaHandler {
	return &FreeQuotaHandler{freeQuotaService: freeQuotaService}
}

type RedeemTrialCardRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *FreeQuotaHandler) RedeemTrialCard(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req RedeemTrialCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	card, ledger, err := h.freeQuotaService.RedeemTrialCard(c.Request.Context(), subject.UserID, req.Code)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"trial_card": card,
		"free_quota": ledger,
	})
}

func (h *FreeQuotaHandler) GetSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	summary, err := h.freeQuotaService.GetUserFreeQuotaSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, summary)
}

func (h *FreeQuotaHandler) GetSettings(c *gin.Context) {
	settings, err := h.freeQuotaService.GetSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}
