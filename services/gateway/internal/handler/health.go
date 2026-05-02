package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo"
	"github.com/gin-gonic/gin"
)

// Health 返回进程存活、版本号、进程启动时间、已运行秒数与服务端当前时间，供外部监控探活。
// HEAD 与 GET 语义一致：成功均为 **200**；HEAD MUST NOT 携带响应体（可设置与 GET 等价的 **Content-Length** 与 **Content-Type**）。
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		start := buildinfo.ProcessStart
		uptime := now.Sub(start).Seconds()
		payload := gin.H{
			"status":         "ok",
			"version":        buildinfo.Version,
			"start_time":     start.Format(time.RFC3339Nano),
			"uptime_seconds": uptime,
			"server_time":    now.Format(time.RFC3339Nano),
		}
		if c.Request.Method == http.MethodHead {
			b, err := json.Marshal(payload)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("Content-Length", strconv.Itoa(len(b)))
			c.Status(http.StatusOK)
			return
		}
		c.JSON(http.StatusOK, payload)
	}
}
