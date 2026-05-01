package handler

import (
	"net/http"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo"
	"github.com/gin-gonic/gin"
)

// Health 返回进程存活、版本号与服务端当前时间，供外部监控探活。
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"version":     buildinfo.Version,
			"server_time": time.Now().UTC().Format(time.RFC3339Nano),
		})
	}
}
