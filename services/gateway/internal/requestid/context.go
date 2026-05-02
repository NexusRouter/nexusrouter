package requestid

import "context"

type idKey struct{}

// WithID 将请求 ID 写入 ctx，便于 http.Client、WithTimeout 等子上下文继承同一 ID。
// ctx 不得为 nil，否则 panic。id 为空字符串时原样返回 ctx。
func WithID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		panic("requestid: nil Context")
	}
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, idKey{}, id)
}

// FromContext 返回 ctx 中的请求 ID；未设置或类型不匹配时返回空字符串。
// ctx 不得为 nil，否则 panic。
func FromContext(ctx context.Context) string {
	if ctx == nil {
		panic("requestid: nil Context")
	}
	v, _ := ctx.Value(idKey{}).(string)
	return v
}
