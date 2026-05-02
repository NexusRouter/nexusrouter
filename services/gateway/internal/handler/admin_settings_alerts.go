package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/alerts"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
)

func adminAuthMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(adminauth.CtxClaims)
		if !ok {
			WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "无令牌信息")
			return
		}
		cl, ok := raw.(*adminauth.Claims)
		if !ok || cl == nil {
			WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "无效令牌")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"role":    strings.TrimSpace(cl.Role),
			"subject": cl.Subject,
		})
	}
}

type settingField struct {
	Key        string `json:"key"`
	Value      any    `json:"value"`
	Mutability string `json:"mutability"` // hot_reload | restart_required | read_only
	Hint       string `json:"hint,omitempty"`
}

func adminSystemSettingsGet(cfg *config.Config, rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "配置未加载")
			return
		}
		var snap *runtime.Snapshot
		if rt != nil {
			snap = rt.Snapshot()
		}
		fields := []settingField{
			{Key: "http_listen_addr", Value: cfg.HTTPListenAddr, Mutability: "restart_required", Hint: "环境变量 NEXUSROUTER_HTTP_LISTEN_ADDR"},
			{Key: "upstream_timeout", Value: cfg.UpstreamTimeout.String(), Mutability: "restart_required", Hint: "NEXUSROUTER_UPSTREAM_TIMEOUT"},
			{Key: "gateway_config_file", Value: cfg.GatewayConfigFile, Mutability: "read_only", Hint: "修改须改环境变量并重启"},
		}
		if snap != nil {
			fields = append(fields,
				settingField{Key: "proxy_access_log_enabled", Value: snap.ProxyAccessLog.Enabled, Mutability: "hot_reload"},
				settingField{Key: "proxy_access_log_path", Value: snap.ProxyAccessLog.Path, Mutability: "hot_reload"},
				settingField{Key: "proxy_access_log_level", Value: snap.ProxyAccessLog.Level, Mutability: "hot_reload"},
			)
		}
		c.JSON(http.StatusOK, gin.H{"settings": fields})
	}
}

type putSystemSettingsBody struct {
	ProxyAccessLog *runtime.ProxyAccessLog `json:"proxy_access_log"`
	Persist        bool                    `json:"persist"`
}

func adminSystemSettingsPut(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rt == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "运行时未初始化")
			return
		}
		var body putSystemSettingsBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON body")
			return
		}
		if body.ProxyAccessLog == nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "至少提供 proxy_access_log")
			return
		}
		next := runtime.CloneSnapshot(rt.Snapshot())
		next.ProxyAccessLog = *body.ProxyAccessLog
		if next.ProxyAccessLog.Level == "" {
			next.ProxyAccessLog.Level = "info"
		}
		if err := runtime.ValidateSnapshot(next); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}
		if err := persistOrApply(rt, next, body.Persist); err != nil {
			WritePersistError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":           "ok",
			"restart_required": false,
			"note":             "监听端口与上游超时须改环境变量并重启进程后生效",
		})
	}
}

func adminAlertsStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		st := alerts.Current()
		c.JSON(http.StatusOK, gin.H{
			"level":   st.Level,
			"reasons": st.Reasons,
			"enabled": st.Enabled,
			"time":    time.Now().UTC().Format(time.RFC3339Nano),
		})
	}
}
