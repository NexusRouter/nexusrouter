package handler

import (
	"net/http"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo"
	"github.com/gin-gonic/gin"
)

// APIStatus 注册为 GET /api/status：无需鉴权，返回统一 success/data 封装的进程版本与启动时刻（RFC3339Nano），供外部脚本或控制台读取。
func APIStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := buildinfo.ProcessStart
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data": gin.H{
				"version":    buildinfo.Version,
				"start_time": start.Format(time.RFC3339Nano),
			},
		})
	}
}
