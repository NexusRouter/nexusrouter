package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register 注册业务路由。
func Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
