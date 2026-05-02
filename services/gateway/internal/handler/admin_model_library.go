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

// registerModelLibraryRoutes 注册 /api/admin/v1/model-library/*（四表聚合）。
func registerModelLibraryRoutes(g, gw *gin.RouterGroup, cfg *config.Config, rt *runtime.Store, db *gorm.DB, log *zap.Logger) {
	if db == nil || rt == nil {
		return
	}
	// vendors
	g.GET("/model-library/vendors", adminListVendors(db))
	g.GET("/model-library/vendors/:id", adminGetVendor(db))
	gw.POST("/model-library/vendors", adminCreateVendor(db))
	gw.PUT("/model-library/vendors/:id", adminUpdateVendor(db))
	gw.DELETE("/model-library/vendors/:id", adminDeleteVendor(db))
	// bases
	g.GET("/model-library/bases", adminListBases(db))
	g.GET("/model-library/bases/:id", adminGetBase(db))
	gw.POST("/model-library/bases", adminCreateBase(db))
	gw.PUT("/model-library/bases/:id", adminUpdateBase(db))
	gw.DELETE("/model-library/bases/:id", adminDeleteBase(db))
	// upstreams
	g.GET("/model-library/upstreams", adminListUpstreams(db))
	g.GET("/model-library/upstreams/:id", adminGetUpstream(db))
	gw.POST("/model-library/upstreams", adminCreateUpstream(db))
	gw.PUT("/model-library/upstreams/:id", adminUpdateUpstream(db))
	gw.DELETE("/model-library/upstreams/:id", adminDeleteUpstream(db))
	// instances
	g.GET("/model-library/instances", adminListInstances(db))
	g.GET("/model-library/instances/:id", adminGetInstance(db))
	gw.POST("/model-library/instances", adminCreateInstance(db))
	gw.PUT("/model-library/instances/:id", adminUpdateInstance(db))
	gw.DELETE("/model-library/instances/:id", adminDeleteInstance(db))
	// sync
	gw.POST("/model-library/sync", adminSyncUpstreamModelsDB(cfg, log, db))
	gw.POST("/model-library/vendor-logo", adminUploadVendorLogo(cfg, log))
}

func parseID64(c *gin.Context, key string) (int64, bool) {
	s := strings.TrimSpace(c.Param(key))
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// --- vendors ---

func adminListVendors(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows []repository.ModelVendor
		q := db.Model(&repository.ModelVendor{})
		if v := strings.TrimSpace(c.Query("vendor_code")); v != "" {
			q = q.Where("vendor_code = ?", v)
		}
		if s := strings.TrimSpace(c.Query("status")); s != "" {
			if s == "0" || s == "1" {
				q = q.Where("status = ?", s)
			}
		}
		if err := q.Order("id asc").Find(&rows).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取厂商失败")
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			out = append(out, vendorToJSON(r))
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
	}
}

func vendorToJSON(r repository.ModelVendor) gin.H {
	return gin.H{
		"id":          r.ID,
		"vendor_name": r.VendorName,
		"vendor_type": r.VendorType,
		"vendor_code": r.VendorCode,
		"logo":        nullStr(r.Logo),
		"status":      r.Status,
		"created_at":  r.CreatedAt,
		"updated_at":  r.UpdatedAt,
	}
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func adminGetVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var r repository.ModelVendor
		if err := db.First(&r, id).Error; err != nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "厂商不存在")
			return
		}
		c.JSON(http.StatusOK, vendorToJSON(r))
	}
}

type vendorBody struct {
	VendorName string `json:"vendor_name"`
	VendorType int8   `json:"vendor_type"`
	VendorCode string `json:"vendor_code"`
	Logo       string `json:"logo"`
	Status     *int8  `json:"status"`
}

func adminCreateVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body vendorBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		st := int8(1)
		if body.Status != nil {
			st = *body.Status
		}
		row := repository.ModelVendor{
			VendorName: strings.TrimSpace(body.VendorName),
			VendorType: body.VendorType,
			VendorCode: strings.TrimSpace(body.VendorCode),
			Logo:       strings.TrimSpace(body.Logo),
			Status:     st,
		}
		if row.VendorName == "" || row.VendorCode == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "vendor_name 与 vendor_code 必填")
			return
		}
		if err := db.Create(&row).Error; err != nil {
			WriteGatewayError(c, http.StatusConflict, "CONFLICT", "创建失败或编码冲突")
			return
		}
		c.JSON(http.StatusCreated, vendorToJSON(row))
	}
}

func adminUpdateVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var body vendorBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		updates := map[string]any{
			"vendor_name": strings.TrimSpace(body.VendorName),
			"vendor_type": body.VendorType,
			"vendor_code": strings.TrimSpace(body.VendorCode),
			"logo":        strings.TrimSpace(body.Logo),
		}
		if body.Status != nil {
			updates["status"] = *body.Status
		}
		res := db.Model(&repository.ModelVendor{}).Where("id = ?", id).Updates(updates)
		if res.Error != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "更新失败")
			return
		}
		if res.RowsAffected == 0 {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "厂商不存在")
			return
		}
		var row repository.ModelVendor
		_ = db.First(&row, id).Error
		c.JSON(http.StatusOK, vendorToJSON(row))
	}
}

func adminDeleteVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		if err := db.Delete(&repository.ModelVendor{}, id).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "删除失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// --- bases ---

func adminListBases(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := db.Model(&repository.ModelBase{})
		if s := strings.TrimSpace(c.Query("model_code")); s != "" {
			q = q.Where("model_code = ?", s)
		}
		if s := strings.TrimSpace(c.Query("status")); s == "0" || s == "1" {
			q = q.Where("status = ?", s)
		}
		var rows []repository.ModelBase
		if err := q.Order("sort asc, id asc").Find(&rows).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取逻辑模型失败")
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			out = append(out, baseToJSON(r))
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
	}
}

func baseToJSON(r repository.ModelBase) gin.H {
	return gin.H{
		"id":         r.ID,
		"model_name": r.ModelName,
		"model_code": r.ModelCode,
		"model_type": r.ModelType,
		"capability": nullJSON(r.Capability),
		"sort":       r.Sort,
		"status":     r.Status,
		"created_at": r.CreatedAt,
		"updated_at": r.UpdatedAt,
	}
}

func nullJSON(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return json.RawMessage(s)
}

func adminGetBase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var r repository.ModelBase
		if err := db.First(&r, id).Error; err != nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "逻辑模型不存在")
			return
		}
		c.JSON(http.StatusOK, baseToJSON(r))
	}
}

type baseBody struct {
	ModelName  string `json:"model_name"`
	ModelCode  string `json:"model_code"`
	ModelType  int8   `json:"model_type"`
	Capability string `json:"capability"`
	Sort       *int   `json:"sort"`
	Status     *int8  `json:"status"`
}

func adminCreateBase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body baseBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		row := repository.ModelBase{
			ModelName:  strings.TrimSpace(body.ModelName),
			ModelCode:  strings.TrimSpace(body.ModelCode),
			ModelType:  body.ModelType,
			Capability: body.Capability,
			Sort:       0,
			Status:     1,
		}
		if body.Sort != nil {
			row.Sort = *body.Sort
		}
		if body.Status != nil {
			row.Status = *body.Status
		}
		if row.ModelName == "" || row.ModelCode == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "model_name 与 model_code 必填")
			return
		}
		if err := db.Create(&row).Error; err != nil {
			WriteGatewayError(c, http.StatusConflict, "CONFLICT", "创建失败或 model_code 冲突")
			return
		}
		c.JSON(http.StatusCreated, baseToJSON(row))
	}
}

func adminUpdateBase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var body baseBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		updates := map[string]any{
			"model_name": strings.TrimSpace(body.ModelName),
			"model_code": strings.TrimSpace(body.ModelCode),
			"model_type": body.ModelType,
			"capability": body.Capability,
		}
		if body.Sort != nil {
			updates["sort"] = *body.Sort
		}
		if body.Status != nil {
			updates["status"] = *body.Status
		}
		res := db.Model(&repository.ModelBase{}).Where("id = ?", id).Updates(updates)
		if res.Error != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "更新失败")
			return
		}
		if res.RowsAffected == 0 {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "逻辑模型不存在")
			return
		}
		var row repository.ModelBase
		_ = db.First(&row, id).Error
		c.JSON(http.StatusOK, baseToJSON(row))
	}
}

func adminDeleteBase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		if err := db.Delete(&repository.ModelBase{}, id).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "删除失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// --- upstreams ---

func upstreamToJSON(r repository.ModelUpstream, maskKey bool) gin.H {
	h := gin.H{
		"id":             r.ID,
		"vendor_id":      r.VendorID,
		"upstream_name":  r.UpstreamName,
		"base_url":       r.BaseURL,
		"timeout":        r.Timeout,
		"max_concurrent": r.MaxConcurrent,
		"status":         r.Status,
		"created_at":     r.CreatedAt,
		"updated_at":     r.UpdatedAt,
		"api_key_set":    strings.TrimSpace(r.APIKey) != "",
	}
	if !maskKey {
		h["api_key"] = r.APIKey
	}
	return h
}

func adminListUpstreams(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := db.Model(&repository.ModelUpstream{})
		if s := strings.TrimSpace(c.Query("vendor_id")); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				q = q.Where("vendor_id = ?", n)
			}
		}
		if s := strings.TrimSpace(c.Query("status")); s == "0" || s == "1" {
			q = q.Where("status = ?", s)
		}
		var rows []repository.ModelUpstream
		if err := q.Order("id asc").Find(&rows).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取上游失败")
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			out = append(out, upstreamToJSON(r, true))
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
	}
}

func adminGetUpstream(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var r repository.ModelUpstream
		if err := db.First(&r, id).Error; err != nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "上游不存在")
			return
		}
		c.JSON(http.StatusOK, upstreamToJSON(r, true))
	}
}

type upstreamBody struct {
	VendorID      int64  `json:"vendor_id"`
	UpstreamName  string `json:"upstream_name"`
	BaseURL       string `json:"base_url"`
	APIKey        string `json:"api_key"`
	Timeout       *int   `json:"timeout"`
	MaxConcurrent *int   `json:"max_concurrent"`
	Status        *int8  `json:"status"`
}

func adminCreateUpstream(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body upstreamBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		row := repository.ModelUpstream{
			VendorID:      body.VendorID,
			UpstreamName:  strings.TrimSpace(body.UpstreamName),
			BaseURL:       strings.TrimSpace(body.BaseURL),
			APIKey:        body.APIKey,
			Timeout:       30,
			MaxConcurrent: 100,
			Status:        1,
		}
		if body.Timeout != nil {
			row.Timeout = *body.Timeout
		}
		if body.MaxConcurrent != nil {
			row.MaxConcurrent = *body.MaxConcurrent
		}
		if body.Status != nil {
			row.Status = *body.Status
		}
		if row.VendorID <= 0 || row.UpstreamName == "" || row.BaseURL == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "vendor_id、upstream_name、base_url 必填")
			return
		}
		if err := db.Create(&row).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "创建失败")
			return
		}
		c.JSON(http.StatusCreated, upstreamToJSON(row, true))
	}
}

func adminUpdateUpstream(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var body upstreamBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		updates := map[string]any{
			"vendor_id":     body.VendorID,
			"upstream_name": strings.TrimSpace(body.UpstreamName),
			"base_url":      strings.TrimSpace(body.BaseURL),
		}
		if body.APIKey != "" {
			updates["api_key"] = body.APIKey
		}
		if body.Timeout != nil {
			updates["timeout"] = *body.Timeout
		}
		if body.MaxConcurrent != nil {
			updates["max_concurrent"] = *body.MaxConcurrent
		}
		if body.Status != nil {
			updates["status"] = *body.Status
		}
		res := db.Model(&repository.ModelUpstream{}).Where("id = ?", id).Updates(updates)
		if res.Error != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "更新失败")
			return
		}
		if res.RowsAffected == 0 {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "上游不存在")
			return
		}
		var row repository.ModelUpstream
		_ = db.First(&row, id).Error
		c.JSON(http.StatusOK, upstreamToJSON(row, true))
	}
}

func adminDeleteUpstream(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		if err := db.Delete(&repository.ModelUpstream{}, id).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "删除失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// --- instances ---

func instanceToJSON(r repository.ModelInstance) gin.H {
	return gin.H{
		"id":                  r.ID,
		"base_model_id":       r.BaseModelID,
		"vendor_id":           r.VendorID,
		"upstream_id":         r.UpstreamID,
		"instance_name":       r.InstanceName,
		"provider_model_code": r.ProviderModelCode,
		"weight":              r.Weight,
		"priority":            r.Priority,
		"is_official":         r.IsOfficial,
		"status":              r.Status,
		"created_at":          r.CreatedAt,
		"updated_at":          r.UpdatedAt,
	}
}

func adminListInstances(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := db.Model(&repository.ModelInstance{})
		if s := strings.TrimSpace(c.Query("base_model_id")); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				q = q.Where("base_model_id = ?", n)
			}
		}
		if s := strings.TrimSpace(c.Query("vendor_id")); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				q = q.Where("vendor_id = ?", n)
			}
		}
		if s := strings.TrimSpace(c.Query("status")); s == "0" || s == "1" {
			q = q.Where("status = ?", s)
		}
		var rows []repository.ModelInstance
		if err := q.Order("id asc").Find(&rows).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取实例失败")
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			out = append(out, instanceToJSON(r))
		}
		c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
	}
}

func adminGetInstance(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var r repository.ModelInstance
		if err := db.First(&r, id).Error; err != nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "实例不存在")
			return
		}
		c.JSON(http.StatusOK, instanceToJSON(r))
	}
}

type instanceBody struct {
	BaseModelID       int64  `json:"base_model_id"`
	VendorID          int64  `json:"vendor_id"`
	UpstreamID        int64  `json:"upstream_id"`
	InstanceName      string `json:"instance_name"`
	ProviderModelCode string `json:"provider_model_code"`
	Weight            *int   `json:"weight"`
	Priority          *int8  `json:"priority"`
	IsOfficial        *int8  `json:"is_official"`
	Status            *int8  `json:"status"`
}

func adminCreateInstance(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body instanceBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		row := repository.ModelInstance{
			BaseModelID:       body.BaseModelID,
			VendorID:          body.VendorID,
			UpstreamID:        body.UpstreamID,
			InstanceName:      strings.TrimSpace(body.InstanceName),
			ProviderModelCode: strings.TrimSpace(body.ProviderModelCode),
			Weight:            10,
			Priority:          1,
			IsOfficial:        0,
			Status:            1,
		}
		if body.Weight != nil {
			row.Weight = *body.Weight
		}
		if body.Priority != nil {
			row.Priority = *body.Priority
		}
		if body.IsOfficial != nil {
			row.IsOfficial = *body.IsOfficial
		}
		if body.Status != nil {
			row.Status = *body.Status
		}
		if row.BaseModelID <= 0 || row.VendorID <= 0 || row.UpstreamID <= 0 || row.InstanceName == "" || row.ProviderModelCode == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "必填字段缺失")
			return
		}
		if err := db.Create(&row).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "创建失败")
			return
		}
		c.JSON(http.StatusCreated, instanceToJSON(row))
	}
}

func adminUpdateInstance(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		var body instanceBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		updates := map[string]any{
			"base_model_id":       body.BaseModelID,
			"vendor_id":           body.VendorID,
			"upstream_id":         body.UpstreamID,
			"instance_name":       strings.TrimSpace(body.InstanceName),
			"provider_model_code": strings.TrimSpace(body.ProviderModelCode),
		}
		if body.Weight != nil {
			updates["weight"] = *body.Weight
		}
		if body.Priority != nil {
			updates["priority"] = *body.Priority
		}
		if body.IsOfficial != nil {
			updates["is_official"] = *body.IsOfficial
		}
		if body.Status != nil {
			updates["status"] = *body.Status
		}
		res := db.Model(&repository.ModelInstance{}).Where("id = ?", id).Updates(updates)
		if res.Error != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "更新失败")
			return
		}
		if res.RowsAffected == 0 {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "实例不存在")
			return
		}
		var row repository.ModelInstance
		_ = db.First(&row, id).Error
		c.JSON(http.StatusOK, instanceToJSON(row))
	}
}

func adminDeleteInstance(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID64(c, "id")
		if !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "id 无效")
			return
		}
		if err := db.Delete(&repository.ModelInstance{}, id).Error; err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "删除失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type syncBodyDB struct {
	ModelUpstreamID int64  `json:"model_upstream_id"`
	Bearer          string `json:"bearer"`
}

type upstreamListJSON struct {
	Object string `json:"object"`
	Data   []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func adminSyncUpstreamModelsDB(cfg *config.Config, log *zap.Logger, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body syncBodyDB
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体无效")
			return
		}
		if body.ModelUpstreamID <= 0 {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "model_upstream_id 必填")
			return
		}
		var up repository.ModelUpstream
		if err := db.First(&up, body.ModelUpstreamID).Error; err != nil {
			WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "上游不存在")
			return
		}
		base, err := url.Parse(strings.TrimSpace(up.BaseURL))
		if err != nil || base.Scheme == "" || base.Host == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "base_url 非法")
			return
		}
		token := strings.TrimSpace(body.Bearer)
		if token == "" {
			token = strings.TrimSpace(up.APIKey)
		}
		if token == "" && cfg != nil {
			token = strings.TrimSpace(cfg.UpstreamAPIKey)
		}
		if token == "" {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "需提供 bearer、上游 api_key 或 NEXUSROUTER_UPSTREAM_API_KEY")
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
				log.Warn("model sync upstream request failed", zap.Int64("model_upstream_id", body.ModelUpstreamID), zap.Error(err))
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
			"model_upstream_id": body.ModelUpstreamID,
			"model_ids":         ids,
			"count":             len(ids),
		})
	}
}
