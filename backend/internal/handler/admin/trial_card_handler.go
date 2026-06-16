package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type TrialCardHandler struct {
	freeQuotaService *service.FreeQuotaService
}

func NewTrialCardHandler(freeQuotaService *service.FreeQuotaService) *TrialCardHandler {
	return &TrialCardHandler{freeQuotaService: freeQuotaService}
}

type CreateTrialCardRequest struct {
	Code           string     `json:"code" binding:"required"`
	Name           string     `json:"name"`
	Amount         float64    `json:"amount" binding:"required"`
	MaxRedemptions int        `json:"max_redemptions" binding:"required"`
	PerUserLimit   int        `json:"per_user_limit"`
	Status         string     `json:"status"`
	Notes          string     `json:"notes"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

type UpdateTrialCardRequest struct {
	Name           *string    `json:"name"`
	Amount         *float64   `json:"amount"`
	MaxRedemptions *int       `json:"max_redemptions"`
	PerUserLimit   *int       `json:"per_user_limit"`
	Status         *string    `json:"status"`
	Notes          *string    `json:"notes"`
	ExpiresAt      *time.Time `json:"expires_at"`
	ClearExpiresAt bool       `json:"clear_expires_at"`
}

func (h *TrialCardHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.freeQuotaService.ListTrialCards(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *TrialCardHandler) Create(c *gin.Context) {
	var req CreateTrialCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	card, err := h.freeQuotaService.CreateTrialCard(c.Request.Context(), &service.CreateTrialCardInput{
		Code:           strings.TrimSpace(req.Code),
		Name:           req.Name,
		Amount:         req.Amount,
		MaxRedemptions: req.MaxRedemptions,
		PerUserLimit:   req.PerUserLimit,
		Status:         req.Status,
		Notes:          req.Notes,
		ExpiresAt:      req.ExpiresAt,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, card)
}

func (h *TrialCardHandler) Update(c *gin.Context) {
	id, ok := parseTrialCardID(c)
	if !ok {
		return
	}
	var req UpdateTrialCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	var expiresAt **time.Time
	if req.ClearExpiresAt {
		var nilTime *time.Time
		expiresAt = &nilTime
	} else if req.ExpiresAt != nil {
		expiresAt = &req.ExpiresAt
	}
	card, err := h.freeQuotaService.UpdateTrialCard(c.Request.Context(), id, &service.UpdateTrialCardInput{
		Name:           req.Name,
		Amount:         req.Amount,
		MaxRedemptions: req.MaxRedemptions,
		PerUserLimit:   req.PerUserLimit,
		Status:         req.Status,
		Notes:          req.Notes,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, card)
}

func (h *TrialCardHandler) Delete(c *gin.Context) {
	id, ok := parseTrialCardID(c)
	if !ok {
		return
	}
	if err := h.freeQuotaService.DeleteTrialCard(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Trial card deactivated successfully"})
}

func (h *TrialCardHandler) GetSettings(c *gin.Context) {
	settings, err := h.freeQuotaService.GetSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *TrialCardHandler) UpdateSettings(c *gin.Context) {
	var req service.FreeQuotaSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.freeQuotaService.UpdateSettings(c.Request.Context(), &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, req)
}

func parseTrialCardID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid trial card id")
		return 0, false
	}
	return id, true
}
