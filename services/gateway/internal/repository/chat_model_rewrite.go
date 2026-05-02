package repository

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// RewriteChatCompletionsModelBody 按启用绑定将 body JSON 中的 model 替换为 actual_model（若非空）。
// 无法解析为 JSON 对象、缺少 model、或无匹配绑定时返回原始 body。
func RewriteChatCompletionsModelBody(body []byte, db *gorm.DB, upstreamID string) []byte {
	if db == nil || len(body) == 0 {
		return body
	}
	upstreamID = strings.TrimSpace(upstreamID)
	if upstreamID == "" {
		return body
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body
	}
	rawModel, ok := obj["model"]
	if !ok {
		return body
	}
	var modelStr string
	if err := json.Unmarshal(rawModel, &modelStr); err != nil {
		return body
	}
	modelStr = strings.TrimSpace(modelStr)
	if modelStr == "" {
		return body
	}
	var b ModelUpstreamBinding
	err := db.Where("catalog_entry_id = ? AND upstream_id = ? AND enabled = ?", modelStr, upstreamID, true).
		First(&b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return body
		}
		return body
	}
	newModel := modelStr
	if b.ActualModel != nil {
		if v := strings.TrimSpace(*b.ActualModel); v != "" {
			newModel = v
		}
	}
	if newModel == modelStr {
		return body
	}
	out, err := json.Marshal(newModel)
	if err != nil {
		return body
	}
	obj["model"] = out
	rewritten, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return rewritten
}
