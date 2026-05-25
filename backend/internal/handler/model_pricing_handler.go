package handler

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const avemujicaPricingURL = "https://api.avemujica.moe/api/pricing"

type ModelPricingHandler struct {
	client *http.Client
}

func NewModelPricingHandler() *ModelPricingHandler {
	return &ModelPricingHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *ModelPricingHandler) GetPublicPricing(c *gin.Context) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, avemujicaPricingURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create pricing request"})
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "UseAiForMe/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch model pricing"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read model pricing"})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.Data(http.StatusBadGateway, "application/json; charset=utf-8", body)
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}
