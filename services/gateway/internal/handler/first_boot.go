package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterFirstBootRoutes 注册首次初始化公开/受控接口（路径固定为 /api/bootstrap/v1/*）。
func RegisterFirstBootRoutes(r *gin.Engine, cfg *config.Config, db *gorm.DB, auth *adminauth.Service, log *zap.Logger) {
	if r == nil || db == nil {
		return
	}
	r.GET("/api/bootstrap/v1/status", firstBootStatus(db))
	r.POST("/api/bootstrap/v1/complete", firstBootComplete(cfg, db, log))
	if auth == nil {
		r.POST("/api/bootstrap/v1/reset", func(c *gin.Context) {
			WriteGatewayError(c, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "管理控制台未启用，无法执行重置")
		})
		return
	}
	r.POST("/api/bootstrap/v1/reset", adminJWTMiddleware(auth), firstBootResetAdminOnly(), firstBootReset(db, log))
}

// firstBootStatus 返回初始化状态与阶段。
//
//	@Summary	首次初始化状态
//	@Produce	json
//	@Success	200	{object}	map[string]any
//	@Router		/api/bootstrap/v1/status [get]
func firstBootStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		st, err := repository.GetBootstrapStatus(db, time.Now().UTC())
		if err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "读取初始化状态失败")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"initialized": st.Initialized,
			"phase":       string(st.Phase),
		})
	}
}

type firstBootCompleteBody struct {
	AdminUsername   string `json:"admin_username"`
	AdminPassword   string `json:"admin_password"`
	SiteDisplayName string `json:"site_display_name"`
}

// firstBootComplete 提交首次初始化（匿名，至多成功一次）。
//
//	@Summary	完成首次初始化
//	@Accept		json
//	@Produce	json
//	@Param		body	body		firstBootCompleteBody	true	"初始化载荷"
//	@Success	200		{object}	map[string]any
//	@Router		/api/bootstrap/v1/complete [post]
func firstBootComplete(cfg *config.Config, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil {
			WriteGatewayError(c, http.StatusBadRequest, "BOOTSTRAP_JWT_MISSING", "网关配置不可用")
			return
		}
		if !cfg.EnableAdminConsole {
			WriteGatewayError(c, http.StatusBadRequest, "ADMIN_DISABLED", "管理控制台已关闭：请将 NEXUSROUTER_ENABLE_ADMIN_CONSOLE 设为 true（或删除该环境变量使用默认开启）后重启网关")
			return
		}
		if strings.TrimSpace(cfg.AdminJWTSecret) == "" {
			WriteGatewayError(c, http.StatusBadRequest, "BOOTSTRAP_JWT_MISSING", "请在环境中配置 NEXUSROUTER_ADMIN_JWT_SECRET 后再完成初始化")
			return
		}
		var body firstBootCompleteBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体须为 JSON：admin_username、admin_password、可选 site_display_name")
			return
		}
		err := repository.CompleteFirstBoot(db, time.Now().UTC(), repository.CompleteFirstBootInput{
			AdminUsername:   body.AdminUsername,
			AdminPassword:   body.AdminPassword,
			SiteDisplayName: body.SiteDisplayName,
		})
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrBootstrapAlreadyCompleted):
				WriteGatewayError(c, http.StatusConflict, "BOOTSTRAP_ALREADY_COMPLETED", "系统已完成首次初始化")
			case errors.Is(err, repository.ErrBootstrapInProgress):
				WriteGatewayError(c, http.StatusConflict, "BOOTSTRAP_IN_PROGRESS", "其他请求正在执行初始化，请稍后重试")
			default:
				if log != nil {
					log.Warn("first boot complete failed", zap.Error(err))
				}
				msg := err.Error()
				if strings.Contains(msg, "用户名") || strings.Contains(msg, "密码") {
					WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", msg)
					return
				}
				WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "初始化失败")
			}
			return
		}
		if log != nil {
			log.Info("first boot completed", zap.String("admin_username", strings.TrimSpace(body.AdminUsername)))
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func firstBootResetAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(adminauth.CtxClaims)
		if !ok {
			WriteGatewayError(c, http.StatusForbidden, "FORBIDDEN", "缺少管理员身份")
			c.Abort()
			return
		}
		cl, ok := raw.(*adminauth.Claims)
		if !ok || cl == nil || !strings.EqualFold(strings.TrimSpace(cl.Role), "admin") {
			WriteGatewayError(c, http.StatusForbidden, "FORBIDDEN", "仅超级管理员可重置首次初始化状态")
			c.Abort()
			return
		}
		c.Next()
	}
}

// firstBootReset 将系统恢复为未初始化并清空管理用户表。
//
//	@Summary	重置首次初始化（超级管理员）
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{object}	map[string]any
//	@Router		/api/bootstrap/v1/reset [post]
func firstBootReset(db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := repository.ResetFirstBoot(db); err != nil {
			if log != nil {
				log.Error("first boot reset failed", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "重置失败")
			return
		}
		if log != nil {
			log.Warn("first boot reset executed")
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
