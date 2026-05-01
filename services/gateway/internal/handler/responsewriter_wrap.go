package handler

import "github.com/gin-gonic/gin"

// closeNotifyResponseWriter 为 httptest 等不实现 http.CloseNotifier 的 ResponseWriter
// 提供占位 CloseNotify，避免 Go 1.25+ ReverseProxy 在探测连接时 panic。
type closeNotifyResponseWriter struct {
	gin.ResponseWriter
}

func (w *closeNotifyResponseWriter) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}
