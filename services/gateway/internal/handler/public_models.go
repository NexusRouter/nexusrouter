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

// openAIModelJSON OpenAI List models 单条子集。
type openAIModelJSON struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ListModels 返回 GET /v1/models，仅含已发布（启用绑定且上游存在于快照）的目录项。
func ListModels(db *gorm.DB, rt *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil || rt == nil {
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": []any{}})
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
			out = append(out, openAIModelJSON{
				ID:      r.CatalogID,
				Object:  "model",
				Created: cr,
				OwnedBy: ob,
			})
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
		c.JSON(http.StatusOK, openAIModelJSON{
			ID:      ent.ID,
			Object:  "model",
			Created: cr,
			OwnedBy: ob,
		})
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
