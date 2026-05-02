package router

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/gin-gonic/gin"
)

// openAIV1ExplicitNotImplementedRoutes 与 OpenAI 兼容面中「已识别、明确未实现」的路径及方法（不含已实现的 chat/models）。
var openAIV1ExplicitNotImplementedRoutes = []struct {
	method string
	path   string
}{
	{http.MethodPost, "/v1/completions"},
	{http.MethodPost, "/v1/images/edits"},
	{http.MethodPost, "/v1/images/variations"},
	{http.MethodGet, "/v1/files"},
	{http.MethodPost, "/v1/files"},
	{http.MethodDelete, "/v1/files/:id"},
	{http.MethodGet, "/v1/files/:id"},
	{http.MethodGet, "/v1/files/:id/content"},
	{http.MethodPost, "/v1/fine_tuning/jobs"},
	{http.MethodGet, "/v1/fine_tuning/jobs"},
	{http.MethodGet, "/v1/fine_tuning/jobs/:id"},
	{http.MethodPost, "/v1/fine_tuning/jobs/:id/cancel"},
	{http.MethodGet, "/v1/fine_tuning/jobs/:id/events"},
	{http.MethodDelete, "/v1/models/:model"},
	{http.MethodPost, "/v1/assistants"},
	{http.MethodGet, "/v1/assistants/:id"},
	{http.MethodPost, "/v1/assistants/:id"},
	{http.MethodDelete, "/v1/assistants/:id"},
	{http.MethodGet, "/v1/assistants"},
	{http.MethodPost, "/v1/assistants/:id/files"},
	{http.MethodGet, "/v1/assistants/:id/files/:fileId"},
	{http.MethodDelete, "/v1/assistants/:id/files/:fileId"},
	{http.MethodGet, "/v1/assistants/:id/files"},
	{http.MethodPost, "/v1/threads"},
	{http.MethodGet, "/v1/threads/:id"},
	{http.MethodPost, "/v1/threads/:id"},
	{http.MethodDelete, "/v1/threads/:id"},
	{http.MethodPost, "/v1/threads/:id/messages"},
	{http.MethodGet, "/v1/threads/:id/messages/:messageId"},
	{http.MethodPost, "/v1/threads/:id/messages/:messageId"},
	{http.MethodGet, "/v1/threads/:id/messages/:messageId/files/:filesId"},
	{http.MethodGet, "/v1/threads/:id/messages/:messageId/files"},
	{http.MethodPost, "/v1/threads/:id/runs"},
	{http.MethodGet, "/v1/threads/:id/runs/:runsId"},
	{http.MethodPost, "/v1/threads/:id/runs/:runsId"},
	{http.MethodGet, "/v1/threads/:id/runs"},
	{http.MethodPost, "/v1/threads/:id/runs/:runsId/submit_tool_outputs"},
	{http.MethodPost, "/v1/threads/:id/runs/:runsId/cancel"},
	{http.MethodGet, "/v1/threads/:id/runs/:runsId/steps/:stepId"},
	{http.MethodGet, "/v1/threads/:id/runs/:runsId/steps"},
}

func registerOpenAIV1NotImplementedRoutes(r *gin.Engine, d Deps) {
	log, ks, store, col := d.Log, d.KeyStore, d.Runtime, d.Metrics
	h := handler.OpenAINotImplemented()
	for _, route := range openAIV1ExplicitNotImplementedRoutes {
		r.Handle(route.method, route.path,
			handler.GatewayAuth(ks, col),
			handler.KeyRateLimit(store, log, col),
			h,
		)
	}
}
