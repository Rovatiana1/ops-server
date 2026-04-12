package handlers

import (
	"github.com/gin-gonic/gin"
)

// Metrics godoc
// @Summary Prometheus metrics
// @Description Expose Prometheus metrics
// @Tags system
// @Produce plain
// @Success 200 {string} string "metrics output"
// @Router /metrics/prometheus [get]
func MetricsDoc(c *gin.Context) {
	// no-op (just for swagger)
}
