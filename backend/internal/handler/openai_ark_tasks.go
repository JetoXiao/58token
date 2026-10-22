package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ArkTasks is the entry point for the Seedance/Ark asynchronous task API.
// The route is registered separately so task creation and polling cannot be
// mistaken for OpenAI image billing. The task scheduler/billing implementation
// is intentionally left explicit until the service-layer task store is wired.
func (h *OpenAIGatewayHandler) ArkTasks(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"type":    "not_implemented",
			"message": "Seedance/Ark task scheduling is not implemented in this build",
		},
	})
}
