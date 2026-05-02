package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const maxUpstreamModelsBody = 10 << 20 // 10 MiB

// registerModelLibraryRoutes 注册 /api/admin/v1/model-library/* 。
func registerModelLibraryRoutes(g, gw *gin.RouterGroup, cfg *config.Config, rt *runtime.Store, db *gorm.DB, log *zap.Logger) {
	if db == nil || rt == nil {
		return
	}
	g.GET("/model-library/entries", adminListCatalogEntries(db))
	g.GET("/model-library/entries/:entry_id", adminGetCatalogEntry(db))
	g.GET("/model-library/entries/:entry_id/bindings", adminListBindings(db))
	gw.POST("/model-library/entries", adminCreateCatalogEntry(db))
	gw.PUT("/model-library/entries/:entry_id", adminUpdateCatalogEntry(db))
	gw.DELETE("/model-library/entries/:entry_id", adminDeleteCatalogEntry(db))
	gw.POST("/model-library/entries/:entry_id/bindings", adminCreateBinding(db))
	gw.PATCH("/model-library/bindings/:binding_id", adminPatchBinding(db))
	gw.DELETE("/model-library/bindings/:binding_id", adminDeleteBinding(db))
	gw.POST("/model-library/sync", adminSyncUpstreamModels(cfg, rt, log))
}

func adminListCatalogEntries(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		off, _ := strconv.Atoi(c.Query("offset"))
		lim, _ := strconv.Atoi(c.Query("limit"))
		rows, total, err := repository.ListModelCatalogEntries(db, off, lim)
		if err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取模型目录失败")
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": rows, "total": total})
	}
}

func adminGetCatalogEntry(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("entry_id"))
		row, err := repository.GetModelCatalogEntry(db, id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "目录项不存在")
				return
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取失败")
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

type createCatalogBody struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	OwnedBy     string `json:"owned_by"`
	Metadata    string `json:"metadata"`
}

func adminCreateCatalogEntry(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body createCatalogBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		body.ID = strings.TrimSpace(body.ID)
		if body.ID == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 必填")
			return
		}
		row := &repository.ModelCatalogEntry{
			ID:          body.ID,
			DisplayName: strings.TrimSpace(body.DisplayName),
			OwnedBy:     strings.TrimSpace(body.OwnedBy),
			Metadata:    body.Metadata,
		}
		if err := repository.CreateModelCatalogEntry(db, row); err != nil {
			WriteGatewayError(c, http.StatusConflict, "CONFLICT", "已存在相同 id")
			return
		}
		c.JSON(http.StatusCreated, row)
	}
}

type updateCatalogBody struct {
	DisplayName string `json:"display_name"`
	OwnedBy     string `json:"owned_by"`
	Metadata    string `json:"metadata"`
}

func adminUpdateCatalogEntry(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("entry_id"))
		var body updateCatalogBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		if err := repository.UpdateModelCatalogEntry(db, id, strings.TrimSpace(body.DisplayName), strings.TrimSpace(body.OwnedBy), body.Metadata); err != nil {
			if err == gorm.ErrRecordNotFound {
				WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "目录项不存在")
				return
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "更新失败")
			return
		}
		row, _ := repository.GetModelCatalogEntry(db, id)
		c.JSON(http.StatusOK, row)
	}
}

func adminDeleteCatalogEntry(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("entry_id"))
		if err := repository.DeleteModelCatalogEntry(db, id); err != nil {
			if err == gorm.ErrRecordNotFound {
				WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "目录项不存在")
				return
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "删除失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func adminListBindings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("entry_id"))
		rows, err := repository.ListBindingsForCatalog(db, id)
		if err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取绑定失败")
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": rows})
	}
}

type createBindingBody struct {
	UpstreamID  string  `json:"upstream_id"`
	Enabled     bool    `json:"enabled"`
	Priority    int64   `json:"priority"`
	ActualModel *string `json:"actual_model"`
}

func adminCreateBinding(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cat := strings.TrimSpace(c.Param("entry_id"))
		var body createBindingBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		uid := strings.TrimSpace(body.UpstreamID)
		if uid == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "upstream_id 必填")
			return
		}
		row := &repository.ModelUpstreamBinding{
			CatalogEntryID: cat,
			UpstreamID:     uid,
			Enabled:        body.Enabled,
			Priority:       body.Priority,
			ActualModel:    body.ActualModel,
		}
		if err := repository.CreateModelUpstreamBinding(db, row); err != nil {
			WriteGatewayError(c, http.StatusConflict, "CONFLICT", "绑定已存在或非法")
			return
		}
		c.JSON(http.StatusCreated, row)
	}
}

type patchBindingBody struct {
	Enabled     bool    `json:"enabled"`
	Priority    int64   `json:"priority"`
	ActualModel *string `json:"actual_model"`
}

func adminPatchBinding(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id64, err := strconv.ParseUint(strings.TrimSpace(c.Param("binding_id")), 10, 64)
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "binding_id 无效")
			return
		}
		var body patchBindingBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		if err := repository.UpdateModelUpstreamBinding(db, uint(id64), body.Enabled, body.Priority, body.ActualModel); err != nil {
			if err == gorm.ErrRecordNotFound {
				WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "绑定不存在")
				return
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "更新失败")
			return
		}
		row, _ := repository.GetModelUpstreamBinding(db, uint(id64))
		c.JSON(http.StatusOK, row)
	}
}

func adminDeleteBinding(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id64, err := strconv.ParseUint(strings.TrimSpace(c.Param("binding_id")), 10, 64)
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "binding_id 无效")
			return
		}
		if err := repository.DeleteModelUpstreamBinding(db, uint(id64)); err != nil {
			if err == gorm.ErrRecordNotFound {
				WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "绑定不存在")
				return
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "删除失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type syncBody struct {
	UpstreamID string `json:"upstream_id"`
	Bearer     string `json:"bearer"` // 单次同步可选；不落库
}

type upstreamListJSON struct {
	Object string `json:"object"`
	Data   []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func adminSyncUpstreamModels(cfg *config.Config, rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body syncBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		uid := strings.TrimSpace(body.UpstreamID)
		if uid == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "upstream_id 必填")
			return
		}
		snap := rt.Snapshot()
		var base *url.URL
		for _, u := range snap.Upstreams {
			if strings.TrimSpace(u.ID) == uid {
				var err error
				base, err = url.Parse(strings.TrimSpace(u.BaseURL))
				if err != nil || base.Scheme == "" || base.Host == "" {
					WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "上游 base_url 非法")
					return
				}
				break
			}
		}
		if base == nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "上游不存在于当前快照")
			return
		}
		token := strings.TrimSpace(body.Bearer)
		if token == "" && cfg != nil {
			token = strings.TrimSpace(cfg.UpstreamAPIKey)
		}
		if token == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "需提供 bearer 或配置 NEXUSROUTER_UPSTREAM_API_KEY")
			return
		}
		listURL := base.ResolveReference(&url.URL{Path: "/v1/models"})
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, listURL.String(), nil)
		if err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "构造请求失败")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if log != nil {
				log.Warn("model sync upstream request failed", zap.String("upstream_id", uid), zap.Error(err))
			}
			WriteGatewayError(c, http.StatusBadGateway, "UPSTREAM_UNREACHABLE", "无法连接上游模型列表")
			return
		}
		defer func() { _ = resp.Body.Close() }()
		limited := io.LimitReader(resp.Body, maxUpstreamModelsBody+1)
		raw, err := io.ReadAll(limited)
		if err != nil {
			WriteGatewayError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "读取上游响应失败")
			return
		}
		if len(raw) > maxUpstreamModelsBody {
			WriteGatewayError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "上游响应过大")
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			WriteGatewayError(c, http.StatusBadGateway, "UPSTREAM_ERROR", fmt.Sprintf("上游状态 %d", resp.StatusCode))
			return
		}
		var parsed upstreamListJSON
		if err := json.Unmarshal(raw, &parsed); err != nil {
			WriteGatewayError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "上游响应非预期 JSON")
			return
		}
		ids := make([]string, 0, len(parsed.Data))
		for _, it := range parsed.Data {
			s := strings.TrimSpace(it.ID)
			if s != "" {
				ids = append(ids, s)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"upstream_id": uid,
			"model_ids":   ids,
			"count":       len(ids),
		})
	}
}
