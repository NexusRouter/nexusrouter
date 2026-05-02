package bodyreuse

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// ginKeyRequestBody 在 Gin 上下文中缓存已读取的请求体字节，供链路上多次解析同一 body。
const ginKeyRequestBody = "nexus_request_body_bytes"

// GetRequestBody 返回请求体字节：首次从 c.Request.Body 读取并缓存；之后直接返回缓存。
// 首次读取会耗尽并关闭原 Body，调用方在转发前须调用 ResetRequestBody 或自行设置 Body。
func GetRequestBody(c *gin.Context) ([]byte, error) {
	if v, ok := c.Get(ginKeyRequestBody); ok {
		if b, ok := v.([]byte); ok {
			return b, nil
		}
	}
	if c.Request.Body == nil {
		var empty []byte
		c.Set(ginKeyRequestBody, empty)
		return empty, nil
	}
	raw, err := io.ReadAll(c.Request.Body)
	_ = c.Request.Body.Close()
	if err != nil {
		return nil, err
	}
	c.Set(ginKeyRequestBody, raw)
	return raw, nil
}

// ResetRequestBody 将给定字节写回请求体，并更新缓存与 Content-Length，清空已解析的表单缓存。
func ResetRequestBody(c *gin.Context, body []byte) {
	c.Set(ginKeyRequestBody, body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
	c.Request.Form = nil
	c.Request.PostForm = nil
}

// UnmarshalBodyReusable 将 JSON（或其它 Content-Type）解析到 v，并把原始字节还原到 c.Request.Body，
// 以便后续处理器或反向代理再次读取同一 body。
func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return err
	}
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(strings.TrimSpace(contentType), "application/json") {
		err = json.Unmarshal(requestBody, v)
	} else {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		err = c.ShouldBind(v)
	}
	if err != nil {
		return err
	}
	ResetRequestBody(c, requestBody)
	return nil
}
