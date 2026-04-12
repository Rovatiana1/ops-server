package handlers

import (
	"io"
	"net/http"

	"ops-server/pkg/helper"

	"github.com/gin-gonic/gin"
)

// MetricsHealth godoc
// @Summary Metrics health summary
// @Description Return parsed Prometheus metrics in human readable form
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics/health [get]
func MetricsHealth(c *gin.Context) {

	resp, err := http.Get("http://localhost:8080/api/v1/metrics/prometheus")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	snap := helper.ParsePrometheusMetrics(string(body))
	health := helper.EvaluateHealth(snap)

	c.JSON(200, gin.H{
		"status": health.Status,
		"score":  health.Score,
		"metrics": gin.H{
			"goroutines": snap.Goroutines,
			"memory_mb": gin.H{
				"alloc":    snap.Memory.AllocMB,
				"heap":     snap.Memory.HeapMB,
				"resident": snap.Memory.ResidentMB,
			},
			"fds":         snap.FDS,
			"cpu_seconds": snap.CPU,
		},
		"issues": health.Issues,
	})
}
