package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type healthHandler struct {
}

func NewHealthHandler() *healthHandler {
	return &healthHandler{}
}

func (h *healthHandler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

func (h *healthHandler) RegisterRoutes(group gin.IRoutes) {
	group.GET("/health", h.health)
}
