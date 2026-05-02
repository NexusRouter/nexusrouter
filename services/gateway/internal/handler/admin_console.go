package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterAdminConsole 注册 /api/admin/v1/*（需 cfg.AdminConsoleConfigured() 且依赖非空）。
func RegisterAdminConsole(
	r *gin.Engine,
	cfg *config.Config,
	auth *adminauth.Service,
	col *metrics.Collector,
	rt *runtime.Store,
	ks *keystore.Store,
	log *zap.Logger,
) {
	if cfg == nil || !cfg.AdminConsoleConfigured() || auth == nil {
		return
	}

	pub := r.Group("/api/admin/v1")
	{
		pub.POST("/auth/login", adminLogin(cfg, auth, col, log))
		pub.GET("/auth/password-reset-info", adminPasswordResetInfo(cfg))
	}

	g := r.Group("/api/admin/v1", adminJWTMiddleware(auth))
	gw := g.Group("", adminOperatorWriteGuard())
	{
		g.GET("/auth/me", adminAuthMe())
		g.POST("/auth/logout", adminLogout())
		g.GET("/metrics/summary", adminMetricsSummary(col))
		g.GET("/gateway/snapshot", adminGatewaySnapshot(rt))
		g.GET("/keys", adminListKeys(ks))
		g.GET("/system/settings", adminSystemSettingsGet(cfg, rt))
		g.GET("/alerts/status", adminAlertsStatus())
		gw.PUT("/gateway/active-upstream", adminSetActiveUpstreamPersist(cfg, rt, log))
		gw.PUT("/gateway/config", adminPutGatewayConfig(cfg, rt, log))
		gw.POST("/keys", adminCreateKey(ks, log))
		gw.PATCH("/keys/:id", adminPatchKey(ks, log))
		gw.DELETE("/keys/:id", adminDeleteKey(ks, log))
		gw.POST("/keys/batch-disable", adminBatchDisable(ks, log))
		gw.POST("/keys/batch-delete", adminBatchDelete(ks, log))
		gw.PUT("/system/settings", adminSystemSettingsPut(rt))
		registerAdminAdvanced(g, gw, rt, log)
	}
}

type adminLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func adminJWTMiddleware(auth *adminauth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth == nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "管理控制台未启用")
			return
		}
		cl, err := auth.Parse(c.GetHeader("Authorization"))
		if err != nil {
			WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "管理令牌无效或缺失")
			return
		}
		c.Set(adminauth.CtxClaims, cl)
		c.Next()
	}
}

func adminLogin(cfg *config.Config, auth *adminauth.Service, col *metrics.Collector, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body adminLoginBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体须为 JSON：username、password")
			return
		}
		tok, exp, role, err := auth.Login(body.Username, body.Password, body.Remember)
		if err != nil {
			if col != nil {
				col.RecordGatewayError("LOGIN_FAILED")
			}
			WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "用户名或密码错误")
			return
		}
		if log != nil {
			log.Info("admin login ok", zap.String("role", role))
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token": tok,
			"token_type":   "Bearer",
			"expires_at":   exp.UTC().Format(time.RFC3339Nano),
			"role":         role,
		})
	}
}

func adminLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	}
}

func adminPasswordResetInfo(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		hasSMTP := strings.TrimSpace(cfg.AdminPasswordResetSMTP) != ""
		c.JSON(http.StatusOK, gin.H{
			"email_configured": hasSMTP,
			"hint":               "若未配置邮件，请联系运维在主机上更新 NEXUSROUTER_ADMIN_PASSWORD_BCRYPT 或通过安全渠道重置。",
			"doc":                "详见 services/gateway/README.md 管理控制台章节。",
		})
	}
}

func adminMetricsSummary(col *metrics.Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		if col == nil {
			c.JSON(http.StatusOK, gin.H{"online": false})
			return
		}
		c.JSON(http.StatusOK, col.SummaryJSON())
	}
}

func adminGatewaySnapshot(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rt == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "运行时未初始化")
			return
		}
		s := rt.Snapshot()
		c.JSON(http.StatusOK, gin.H{
			"upstreams":         s.Upstreams,
			"routing":           s.Routing,
			"cors":              s.CORS,
			"rate_limit":        s.RateLimit,
			"rate_limit_rules":  s.RateLimitRules,
			"ip_access":         s.IPAccess,
			"proxy_access_log":  s.ProxyAccessLog,
			"config_file":       rt.Path(),
			"config_file_set":   strings.TrimSpace(rt.Path()) != "",
		})
	}
}

type putActiveBody struct {
	ActiveUpstreamID string `json:"active_upstream_id"`
	Persist          bool   `json:"persist"`
}

func adminSetActiveUpstreamPersist(cfg *config.Config, rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rt == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "运行时未初始化")
			return
		}
		var body putActiveBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON：active_upstream_id")
			return
		}
		if err := rt.SetActiveUpstream(body.ActiveUpstreamID); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_UPSTREAM", err.Error())
			return
		}
		if body.Persist {
			if err := rt.PersistCurrent(); err != nil {
				if log != nil {
					log.Error("persist active upstream failed", zap.Error(err))
				}
				WriteGatewayError(c, http.StatusInternalServerError, "PERSIST_FAILED", err.Error())
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "active_upstream_id": strings.TrimSpace(body.ActiveUpstreamID)})
	}
}

type putGatewayBody struct {
	Upstreams []runtime.Upstream `json:"upstreams"`
	Routing   runtime.Routing    `json:"routing"`
	Persist   bool               `json:"persist"`
}

func adminPutGatewayConfig(cfg *config.Config, rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rt == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "运行时未初始化")
			return
		}
		var body putGatewayBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON：upstreams、routing")
			return
		}
		next, err := runtime.NewSnapshotFromBase(rt.Snapshot(), body.Upstreams, body.Routing)
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}
		if !body.Persist {
			if err := rt.ApplySnapshot(next); err != nil {
				WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok", "persisted": false})
			return
		}
		if err := rt.PersistSnapshot(next); err != nil {
			if log != nil {
				log.Error("persist gateway yaml failed", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "PERSIST_FAILED", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "persisted": true})
	}
}

func adminListKeys(ks *keystore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ks == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "密钥库未初始化")
			return
		}
		list, err := ks.ListPublic()
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "KEYS_FILE_REQUIRED", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

type createKeyBody struct {
	Secret    string `json:"secret"`
	ExpiresAt string `json:"expires_at"`
}

func adminCreateKey(ks *keystore.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ks == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "密钥库未初始化")
			return
		}
		var body createKeyBody
		_ = c.ShouldBindJSON(&body)
		sec := strings.TrimSpace(body.Secret)
		if sec == "" {
			sec = keystore.NewRandomSecret()
		}
		var exp *time.Time
		if strings.TrimSpace(body.ExpiresAt) != "" {
			t, err := time.Parse(time.RFC3339Nano, body.ExpiresAt)
			if err != nil {
				t, err = time.Parse(time.RFC3339, body.ExpiresAt)
			}
			if err != nil {
				WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "expires_at 须为 RFC3339")
				return
			}
			tu := t.UTC()
			exp = &tu
		}
		now := time.Now().UTC()
		rec := keystore.Record{
			ID:        newKeyID(),
			Secret:    sec,
			Disabled:  false,
			ExpiresAt: exp,
			CreatedAt: &now,
		}

		cur := ks.SnapshotRecords()
		cur = append(cur, rec)
		if err := ks.ReplaceAllRecords(cur); err != nil {
			if log != nil {
				log.Error("create api key failed", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "KEY_WRITE_FAILED", err.Error())
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"id":         rec.ID,
			"secret":     sec,
			"expires_at": rec.ExpiresAt,
			"created_at": rec.CreatedAt,
			"warning":    "请立即保存 secret，之后仅显示脱敏值",
		})
	}
}

func newKeyID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "key_" + hex.EncodeToString(b)
}

type patchKeyBody struct {
	Disabled  *bool   `json:"disabled"`
	ExpiresAt *string `json:"expires_at"`
}

func adminPatchKey(ks *keystore.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("id"))
		var body patchKeyBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON body")
			return
		}
		cur := ks.SnapshotRecords()
		found := false
		for i := range cur {
			if strings.TrimSpace(cur[i].ID) == id || (id == "(legacy)" && strings.TrimSpace(cur[i].ID) == "") {
				found = true
				if body.Disabled != nil {
					cur[i].Disabled = *body.Disabled
				}
				if body.ExpiresAt != nil {
					if strings.TrimSpace(*body.ExpiresAt) == "" {
						cur[i].ExpiresAt = nil
					} else {
						t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*body.ExpiresAt))
						if err != nil {
							t, err = time.Parse(time.RFC3339, strings.TrimSpace(*body.ExpiresAt))
						}
						if err != nil {
							WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "expires_at RFC3339")
							return
						}
						tu := t.UTC()
						cur[i].ExpiresAt = &tu
					}
				}
				break
			}
		}
		if !found {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "密钥不存在")
			return
		}
		if err := ks.ReplaceAllRecords(cur); err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "KEY_WRITE_FAILED", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func adminDeleteKey(ks *keystore.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("id"))
		cur := ks.SnapshotRecords()
		next := make([]keystore.Record, 0, len(cur))
		ok := false
		for _, r := range cur {
			rid := strings.TrimSpace(r.ID)
			if rid == id || (id == "(legacy)" && rid == "") {
				ok = true
				continue
			}
			next = append(next, r)
		}
		if !ok {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "密钥不存在")
			return
		}
		if err := ks.ReplaceAllRecords(next); err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "KEY_WRITE_FAILED", err.Error())
			return
		}
		_ = log
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

type batchIDsBody struct {
	IDs []string `json:"ids"`
}

func adminBatchDisable(ks *keystore.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body batchIDsBody
		if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "ids 非空数组")
			return
		}
		set := map[string]struct{}{}
		for _, id := range body.IDs {
			set[strings.TrimSpace(id)] = struct{}{}
		}
		cur := ks.SnapshotRecords()
		for i := range cur {
			rid := strings.TrimSpace(cur[i].ID)
			if rid == "" {
				rid = "(legacy)"
			}
			if _, ok := set[rid]; ok {
				cur[i].Disabled = true
			}
		}
		if err := ks.ReplaceAllRecords(cur); err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "KEY_WRITE_FAILED", err.Error())
			return
		}
		_ = log
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func adminBatchDelete(ks *keystore.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body batchIDsBody
		if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "ids 非空数组")
			return
		}
		set := map[string]struct{}{}
		for _, id := range body.IDs {
			set[strings.TrimSpace(id)] = struct{}{}
		}
		cur := ks.SnapshotRecords()
		next := make([]keystore.Record, 0, len(cur))
		for _, r := range cur {
			rid := strings.TrimSpace(r.ID)
			if rid == "" {
				rid = "(legacy)"
			}
			if _, drop := set[rid]; drop {
				continue
			}
			next = append(next, r)
		}
		if err := ks.ReplaceAllRecords(next); err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "KEY_WRITE_FAILED", err.Error())
			return
		}
		_ = log
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
