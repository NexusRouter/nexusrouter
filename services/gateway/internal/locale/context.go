package locale

import "context"

type tagKey struct{}

// WithLocale 将语言标签写入 ctx，子上下文（如 WithTimeout）可经 FromContext 读取同一标签。
func WithLocale(ctx context.Context, tag string) context.Context {
	if ctx == nil || tag == "" {
		return ctx
	}
	return context.WithValue(ctx, tagKey{}, tag)
}

// FromContext 返回 ctx 中的语言标签；未设置或类型不匹配时返回空字符串。
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(tagKey{}).(string)
	return v
}
