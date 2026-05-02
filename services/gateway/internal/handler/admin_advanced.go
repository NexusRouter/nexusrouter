package handler

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/accesslog"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/ipaccess"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var methodToken = regexp.MustCompile(`^[A-Z][A-Z0-9-]*$`)

func splitBulkLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, ",", "\n")
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func validateCORS(c runtime.CORS) error {
	for _, m := range c.AllowMethods {
		m = strings.TrimSpace(m)
		if m == "" {
			return fmt.Errorf("空方法名")
		}
		if !methodToken.MatchString(m) {
			return fmt.Errorf("非法 HTTP 方法: %q", m)
		}
	}
	for _, h := range c.AllowHeaders {
		h = strings.TrimSpace(h)
		if h == "" {
			return fmt.Errorf("空头名称")
		}
		for _, r := range h {
			if r > unicode.MaxASCII || (!unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_') {
				return fmt.Errorf("非法头名: %q", h)
			}
		}
	}
	return nil
}

func registerAdminAdvanced(g, gw *gin.RouterGroup, rt *runtime.Store, log *zap.Logger) {
	if rt == nil {
		return
	}
	g.GET("/gateway/cors", adminGetCORS(rt))
	gw.PUT("/gateway/cors", adminPutCORS(rt))
	g.GET("/gateway/rate-limit-rules", adminGetRateRules(rt))
	gw.PUT("/gateway/rate-limit-rules", adminPutRateRules(rt))
	g.GET("/security/ip-access", adminGetIPAccess(rt))
	gw.PUT("/security/ip-access", adminPutIPAccess(rt))
	gw.PATCH("/security/ip-access", adminPatchIPAccess(rt))
	g.GET("/logs/query", adminLogsQuery(rt, log))
	g.GET("/logs/export.csv", adminLogsExportCSV(rt, log))
}

func adminGetCORS(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := rt.Snapshot()
		c.JSON(http.StatusOK, s.CORS)
	}
}

type putCORSBody struct {
	Enabled          bool     `json:"enabled"`
	AllowOrigins     []string `json:"allow_origins"`
	AllowOriginsBulk string   `json:"allow_origins_bulk"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	MaxAgeSeconds    int      `json:"max_age_seconds"`
	Persist          bool     `json:"persist"`
}

func adminPutCORS(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body putCORSBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON body")
			return
		}
		next := runtime.CloneSnapshot(rt.Snapshot())
		next.CORS.Enabled = body.Enabled
		orig := append([]string(nil), body.AllowOrigins...)
		orig = append(orig, splitBulkLines(body.AllowOriginsBulk)...)
		next.CORS.AllowOrigins = dedupeStrings(orig)
		next.CORS.AllowMethods = dedupeStrings(body.AllowMethods)
		next.CORS.AllowHeaders = dedupeStrings(body.AllowHeaders)
		next.CORS.MaxAgeSeconds = body.MaxAgeSeconds
		if err := validateCORS(next.CORS); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CORS", err.Error())
			return
		}
		if err := persistOrApply(rt, next, body.Persist); err != nil {
			WritePersistError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "cors": next.CORS})
	}
}

func adminGetRateRules(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := rt.Snapshot()
		c.JSON(http.StatusOK, gin.H{"rules": s.RateLimitRules, "rate_limit": s.RateLimit})
	}
}

type putRateRulesBody struct {
	Rules   []runtime.RateLimitRule `json:"rules"`
	Persist bool                    `json:"persist"`
}

func adminPutRateRules(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body putRateRulesBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON：rules")
			return
		}
		next := runtime.CloneSnapshot(rt.Snapshot())
		next.RateLimitRules = append([]runtime.RateLimitRule(nil), body.Rules...)
		if err := runtime.ValidateSnapshot(next); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}
		if err := persistOrApply(rt, next, body.Persist); err != nil {
			WritePersistError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "rules": next.RateLimitRules})
	}
}

func adminGetIPAccess(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := rt.Snapshot()
		c.JSON(http.StatusOK, s.IPAccess)
	}
}

type putIPAccessBody struct {
	Mode    string   `json:"mode"`
	CIDRs   []string `json:"cidrs"`
	Persist bool     `json:"persist"`
}

func adminPutIPAccess(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body putIPAccessBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON：mode、cidrs")
			return
		}
		next := runtime.CloneSnapshot(rt.Snapshot())
		next.IPAccess.Mode = strings.TrimSpace(body.Mode)
		next.IPAccess.CIDRs = dedupeStrings(body.CIDRs)
		if err := runtime.ValidateSnapshot(next); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}
		if err := persistOrApply(rt, next, body.Persist); err != nil {
			WritePersistError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "ip_access": next.IPAccess})
	}
}

type patchIPAccessBody struct {
	Add     []string `json:"add"`
	Remove  []string `json:"remove"`
	Mode    *string  `json:"mode"`
	Persist bool     `json:"persist"`
}

func adminPatchIPAccess(rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body patchIPAccessBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "JSON：add、remove、mode?")
			return
		}
		next := runtime.CloneSnapshot(rt.Snapshot())
		if body.Mode != nil {
			next.IPAccess.Mode = strings.TrimSpace(*body.Mode)
		}
		cur := map[string]struct{}{}
		for _, x := range next.IPAccess.CIDRs {
			cur[strings.TrimSpace(x)] = struct{}{}
		}
		for _, x := range body.Add {
			x = strings.TrimSpace(x)
			if x != "" {
				cur[x] = struct{}{}
			}
		}
		for _, x := range body.Remove {
			x = strings.TrimSpace(x)
			delete(cur, x)
		}
		out := make([]string, 0, len(cur))
		for x := range cur {
			out = append(out, x)
		}
		sort.Strings(out)
		next.IPAccess.CIDRs = out
		if _, err := ipaccess.Compile(next.IPAccess.CIDRs); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CIDR", err.Error())
			return
		}
		if err := runtime.ValidateSnapshot(next); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}
		if err := persistOrApply(rt, next, body.Persist); err != nil {
			WritePersistError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "ip_access": next.IPAccess})
	}
}

func adminLogsQuery(rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := rt.Snapshot()
		f := accesslog.LogFilters{
			FromRFC3339:  c.Query("from"),
			ToRFC3339:    c.Query("to"),
			PathPrefix:   c.Query("path_prefix"),
			APIKeyFP:     c.Query("api_key_fp"),
			ClientIP:     c.Query("client_ip"),
			Cursor:       c.Query("cursor"),
			MaxScanBytes: int64(parseIntDefault(c.Query("max_scan_mb"), 0) * (1 << 20)),
		}
		if v := c.Query("status_min"); v != "" {
			f.StatusMin, _ = strconv.Atoi(v)
		}
		if v := c.Query("status_max"); v != "" {
			f.StatusMax, _ = strconv.Atoi(v)
		}
		f.Limit = parseIntDefault(c.Query("limit"), 100)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
		defer cancel()
		items, next, truncated, err := accesslog.QueryJSONLines(ctx, s, f)
		if err != nil {
			if log != nil {
				log.Warn("logs query", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusBadRequest, "LOG_QUERY_FAILED", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"items":            items,
			"next_cursor":      next,
			"scan_truncated":   truncated,
			"proxy_log_enabled": s.ProxyAccessLog.Enabled,
		})
	}
}

func adminLogsExportCSV(rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := rt.Snapshot()
		f := accesslog.LogFilters{
			FromRFC3339:  c.Query("from"),
			ToRFC3339:    c.Query("to"),
			PathPrefix:   c.Query("path_prefix"),
			APIKeyFP:     c.Query("api_key_fp"),
			ClientIP:     c.Query("client_ip"),
			MaxScanBytes: int64(parseIntDefault(c.Query("max_scan_mb"), 0) * (1 << 20)),
		}
		if v := c.Query("status_min"); v != "" {
			f.StatusMin, _ = strconv.Atoi(v)
		}
		if v := c.Query("status_max"); v != "" {
			f.StatusMax, _ = strconv.Atoi(v)
		}
		f.Limit = 5000
		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()
		items, _, _, err := accesslog.QueryJSONLines(ctx, s, f)
		if err != nil {
			if log != nil {
				log.Warn("logs export", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusBadRequest, "LOG_EXPORT_FAILED", err.Error())
			return
		}
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", `attachment; filename="access_logs.csv"`)
		_ = accesslog.WriteCSV(c.Writer, items)
	}
}

func parseIntDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func persistOrApply(rt *runtime.Store, next *runtime.Snapshot, persist bool) error {
	if !persist {
		return rt.ApplySnapshot(next)
	}
	return rt.PersistSnapshot(next)
}

// WritePersistError 映射持久化错误到 HTTP。
func WritePersistError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "未配置 NEXUSROUTER_GATEWAY_CONFIG_FILE") {
		WriteGatewayError(c, http.StatusBadRequest, "PERSIST_DISABLED", msg)
		return
	}
	WriteGatewayError(c, http.StatusInternalServerError, "PERSIST_FAILED", msg)
}
