package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// openAIModelPermissionJSON 与常见 OpenAI model 列表项中 permission 元素形状对齐。
type openAIModelPermissionJSON struct {
	ID                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int64   `json:"created"`
	AllowCreateEngine  bool    `json:"allow_create_engine"`
	AllowSampling      bool    `json:"allow_sampling"`
	AllowLogprobs      bool    `json:"allow_logprobs"`
	AllowSearchIndices bool    `json:"allow_search_indices"`
	AllowView          bool    `json:"allow_view"`
	AllowFineTuning    bool    `json:"allow_fine_tuning"`
	Organization       string  `json:"organization"`
	Group              *string `json:"group"`
	IsBlocking         bool    `json:"is_blocking"`
}

// openAIModelJSON OpenAI List/Retrieve models 单条（含 permission、root，便于旧版客户端解析）。
type openAIModelJSON struct {
	ID         string                      `json:"id"`
	Object     string                      `json:"object"`
	Created    int64                       `json:"created"`
	OwnedBy    string                      `json:"owned_by"`
	Permission []openAIModelPermissionJSON `json:"permission"`
	Root       string                      `json:"root"`
	Parent     *string                     `json:"parent,omitempty"`
}

var defaultOpenAIModelPermissionEntry = openAIModelPermissionJSON{
	ID:                 "modelperm-nexusrouter-default",
	Object:             "model_permission",
	Created:            1626777600,
	AllowCreateEngine:  true,
	AllowSampling:      true,
	AllowLogprobs:      true,
	AllowSearchIndices: false,
	AllowView:          true,
	AllowFineTuning:    false,
	Organization:       "*",
	Group:              nil,
	IsBlocking:         false,
}

func newOpenAIModelItem(id, ownedBy string, created int64) openAIModelJSON {
	return openAIModelJSON{
		ID:         id,
		Object:     "model",
		Created:    created,
		OwnedBy:    ownedBy,
		Permission: []openAIModelPermissionJSON{defaultOpenAIModelPermissionEntry},
		Root:       id,
		Parent:     nil,
	}
}

// ListModels 返回 GET /v1/models。
func ListModels(db *gorm.DB, rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil || rt == nil {
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": []any{}})
			return
		}
		if repository.UseDatabaseModelLibrary(db) {
			rows, err := repository.ListPublishedModelsAggregation(db)
			if err != nil {
				WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取模型库失败")
				return
			}
			out := make([]openAIModelJSON, 0, len(rows))
			for _, r := range rows {
				ob := r.OwnedBy
				if ob == "" {
					ob = "nexusrouter"
				}
				cr := r.CreatedAt
				if cr == 0 {
					cr = 1626777600
				}
				out = append(out, newOpenAIModelItem(r.ModelCode, ob, cr))
			}
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": out})
			return
		}
		snap := rt.Snapshot()
		valid := upstreamIDSet(snap)
		rows, err := repository.ListPublishedModels(db, valid)
		if err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取模型库失败")
			return
		}
		out := make([]openAIModelJSON, 0, len(rows))
		for _, r := range rows {
			ob := r.OwnedBy
			if ob == "" {
				ob = "nexusrouter"
			}
			cr := r.CreatedAt
			if cr == 0 {
				cr = 1626777600
			}
			out = append(out, newOpenAIModelItem(r.CatalogID, ob, cr))
		}
		c.JSON(http.StatusOK, gin.H{"object": "list", "data": out})
	}
}

// RetrieveModel GET /v1/models/:model
func RetrieveModel(db *gorm.DB, rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		mid := c.Param("model")
		if db == nil || rt == nil {
			WriteGatewayError(c, http.StatusNotFound, "MODEL_NOT_FOUND", "模型不存在")
			return
		}
		if repository.UseDatabaseModelLibrary(db) {
			ent, err := repository.GetPublishedModelBase(db, mid)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					WriteGatewayError(c, http.StatusNotFound, "MODEL_NOT_FOUND", "模型不存在")
					return
				}
				WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取模型库失败")
				return
			}
			ob := ""
			var inst0 repository.ModelInstance
			if err := db.Where("base_model_id = ? AND status = ?", ent.ID, 1).First(&inst0).Error; err == nil {
				var v repository.ModelVendor
				if err := db.First(&v, inst0.VendorID).Error; err == nil {
					ob = v.VendorName
				}
			}
			if ob == "" {
				ob = "nexusrouter"
			}
			cr := ent.CreatedAt.Unix()
			if cr == 0 {
				cr = 1626777600
			}
			c.JSON(http.StatusOK, newOpenAIModelItem(ent.ModelCode, ob, cr))
			return
		}
		snap := rt.Snapshot()
		valid := upstreamIDSet(snap)
		ok, ent, err := repository.IsCatalogPublished(db, mid, valid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				WriteGatewayError(c, http.StatusNotFound, "MODEL_NOT_FOUND", "模型不存在")
				return
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "读取模型库失败")
			return
		}
		if ent == nil || !ok {
			WriteGatewayError(c, http.StatusNotFound, "MODEL_NOT_FOUND", "模型不存在")
			return
		}
		ob := ent.OwnedBy
		if ob == "" {
			ob = "nexusrouter"
		}
		cr := ent.CreatedAt.Unix()
		if cr == 0 {
			cr = 1626777600
		}
		c.JSON(http.StatusOK, newOpenAIModelItem(ent.ID, ob, cr))
	}
}

func upstreamIDSet(s *runtime.Snapshot) map[string]struct{} {
	m := make(map[string]struct{})
	if s == nil {
		return m
	}
	for _, u := range s.Upstreams {
		id := strings.TrimSpace(u.ID)
		if id != "" {
			m[id] = struct{}{}
		}
	}
	return m
}
