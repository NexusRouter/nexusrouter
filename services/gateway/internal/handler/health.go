package handler

import (
	"net/http"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo"
	"github.com/gin-gonic/gin"
)

// Health 返回进程存活、版本号、进程启动时间、已运行秒数与服务端当前时间，供外部监控探活。
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		start := buildinfo.ProcessStart
		uptime := now.Sub(start).Seconds()
		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"version":        buildinfo.Version,
			"start_time":     start.Format(time.RFC3339Nano),
			"uptime_seconds": uptime,
			"server_time":    now.Format(time.RFC3339Nano),
		})
	}
}
